#!/usr/bin/env bash
# cutover-sqlite-postgres.sh — optionaler, ops-/deploy-zeitiger
# SQLite→Postgres-Cutover (ADR-0007).
#
# Fährt d-migrate als ephemeren Ops-Container (die API-Runtime bleibt
# JDK-frei, ADR-0002). Vier Phasen:
#   Profile-Check (Pre-Flight) → Bulk → inkrementell → Switch.
#
# Subcommands:
#   doctor — profile-unabhängiger Pre-Flight: Tooling erreichbar, Quelle lesbar,
#     Ziel-PG erreichbar + Schema vorhanden + (für Bulk) leer. Voll implementiert.
#   profile — Phase 0 (data profile): self-type-Kompatibilitäts-Pre-Flight der
#     Quelle (Abbruch bei Wert-Typ-Korruption). Voll implementiert (s. cmd_profile).
#   bulk — Phase 1 (data transfer): Erstübertragung aller App-Tabellen +
#     Parität/Sequenz-Erhalt-Verifikation. Voll implementiert.
#   incremental — Phase 2 (data transfer --since): Delta seit Bulk nachziehen,
#     idempotent (--on-conflict skip). Voll implementiert.
#   switch — Phase 3 (Quiesce → finales Delta → Umschalten). Stub (Folge-Arbeit).
#
# ENV:
#   SQLITE_DB Host-Pfad der Quell-SQLite (z. B. /var/lib/mtrace/m-trace.db).
#   PG_DSN Ziel-Postgres-DSN (postgres://user:pass@host:port/db?sslmode=disable).
#   PG_NETWORK Optional: Docker-Netz, dem die Client-/d-migrate-Container
#                    beitreten, wenn der DSN-Host ein Container-Name ist.
#   DMIGRATE_IMAGE Override; Default = Single-Source-Pin aus apps/api/Makefile
#                    (`make -C apps/api -s print-dmigrate-image`). Ohne Repo-
#                    Checkout (Standalone-Deploy) DMIGRATE_IMAGE explizit setzen.
#   SQLITE_IMAGE sqlite3-Client-Image (Default keinos/sqlite3).
#   PG_CLIENT_IMAGE psql-Client-Image (Default postgres:17-alpine).
#   CHUNK_SIZE data-transfer-Chunkgröße (Default 10000).
#   SINCE incremental: ingest_sequence-Untergrenze; Default = Ziel-MAX (Auto-Resume).
#
# Exit-Codes: 0 ok · 1 hard FAIL (Transfer-/Verifikations-Fehler) · 2 Config-/
#             Nutzungsfehler · 3 Pre-Flight-Befund (Ziel nicht bereit) · 4 noch
#             nicht implementiert (Stub).

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
SQLITE_IMAGE="${SQLITE_IMAGE:-keinos/sqlite3}"
PG_CLIENT_IMAGE="${PG_CLIENT_IMAGE:-postgres:17-alpine}"
CHUNK_SIZE="${CHUNK_SIZE:-10000}"
# Erwartetes Ziel-Schema (PG-DDL, ADR-0006): 13 Tabellen. Source of Truth ist
# die eingecheckte PG-DDL + der schema-generate-postgres-check-Drift-Gate; hier
# nur eine grobe „Schema vorhanden?"-Untergrenze (>=). Bei Schema-Änderung
# mitziehen (spiegelt EXPECT_TABLES aus generate-postgres-schema.sh).
EXPECT_TABLES=13

log()  { printf '[cutover] %s\n' "$*"; }
warn() { printf '[cutover] WARN: %s\n' "$*" >&2; }
die()  { printf '[cutover] FEHLER: %s\n' "$*" >&2; exit "${2:-2}"; }

# d-migrate-Image aus dem Single-Source-Pin auflösen (override via ENV).
resolve_dmigrate_image() {
  if [ -n "${DMIGRATE_IMAGE:-}" ]; then
    printf '%s\n' "$DMIGRATE_IMAGE"
    return 0
  fi
  make -C "${REPO_ROOT}/apps/api" -s print-dmigrate-image 2>/dev/null
}

# Optionales `--network`-Argument, falls PG_NETWORK gesetzt.
docker_net_args() {
  if [ -n "${PG_NETWORK:-}" ]; then printf -- '--network\n%s\n' "$PG_NETWORK"; fi
}

# d-migrate im ephemeren Container ausführen (Quelle-Verzeichnis als /work
# gemountet). Aufrufer übergeben d-migrate-Argumente ($@).
run_dmigrate() {
  local img; img="$(resolve_dmigrate_image)"
  [ -n "$img" ] || die "DMIGRATE_IMAGE nicht auflösbar (apps/api print-dmigrate-image leer)"
  local src_dir; src_dir="$(cd "$(dirname "$SQLITE_DB")" && pwd)"
  local net extra
  mapfile -t net < <(docker_net_args)
  # Optionaler zweiter Mount (host:container) für Report-Output außerhalb des
  # Quell-Verzeichnisses (RUN_DMIGRATE_EXTRA_MOUNT).
  extra=(); [ -n "${RUN_DMIGRATE_EXTRA_MOUNT:-}" ] && extra=(-v "${RUN_DMIGRATE_EXTRA_MOUNT}")
  docker run --rm --user "$(id -u):$(id -g)" \
    -v "${src_dir}:/work" -w /work "${net[@]}" "${extra[@]}" "$img" "$@"
}

# psql gegen das Ziel-PG (Client-Container). Gibt tuple-only aus.
pg_psql() {
  local net
  mapfile -t net < <(docker_net_args)
  docker run --rm -i "${net[@]}" "$PG_CLIENT_IMAGE" \
    psql "$PG_DSN" -tA -c "$1"
}

# Read-Query gegen die Quell-SQLite (als d-migrate-User). Gibt den Wert aus.
sqlite_src() {
  local src_dir base
  src_dir="$(cd "$(dirname "$SQLITE_DB")" && pwd)"
  base="$(basename "$SQLITE_DB")"
  docker run --rm --user "$(id -u):$(id -g)" --entrypoint sqlite3 \
    -v "${src_dir}:/work" "$SQLITE_IMAGE" "/work/${base}" "$1"
}

# Sequenz-Erhalt prüfen: der nächste DB-vergebene Wert der PK-Sequenz muss
# ECHT über dem Quell-MAX liegen (sonst PK-Kollision beim ersten neuen Insert).
# Leere Quell-Tabelle (MAX 0, frische Sequenz) ist kollisionsfrei.
verify_sequence() { # table column -> 0 ok / 1 fail
  local tbl="$1" col="$2" seq lastv iscalled srcmax next
  seq="$(pg_psql "SELECT pg_get_serial_sequence('$tbl','$col');" 2>/dev/null || echo '')"
  [ -n "$seq" ] || { warn "  ✘ ${tbl}.${col}: keine PG-Sequenz gefunden"; return 1; }
  lastv="$(pg_psql "SELECT last_value FROM $seq;" 2>/dev/null || echo '')"
  iscalled="$(pg_psql "SELECT is_called FROM $seq;" 2>/dev/null || echo '')"
  srcmax="$(sqlite_src "SELECT COALESCE(MAX($col),0) FROM \"$tbl\";" 2>/dev/null || echo '')"
  if ! [[ "$lastv" =~ ^[0-9]+$ ]] || ! [[ "$srcmax" =~ ^[0-9]+$ ]]; then
    warn "  ✘ ${tbl}.${col}: Sequenz-/MAX-Abfrage fehlgeschlagen (last='${lastv}' max='${srcmax}')"; return 1
  fi
  if [ "$iscalled" = "t" ]; then next=$((lastv + 1)); else next=$lastv; fi
  if [ "$next" -gt "$srcmax" ]; then
    log "  ✔ ${tbl}.${col}: nächster Sequenzwert=${next} > MAX=${srcmax} (kein Kollision)"; return 0
  fi
  warn "  ✘ ${tbl}.${col}: nächster Sequenzwert=${next} <= MAX=${srcmax} — würde kollidieren"; return 1
}

# Transferierbare App-Tabellen der Quelle (ohne Migrations-Bookkeeping:
# schema_migrations wird auf beiden Seiten vom Runner verwaltet, sqlite_% intern).
app_tables() {
  sqlite_src "SELECT group_concat(name) FROM (SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name <> 'schema_migrations' ORDER BY name);"
}

# Row-Count-Parität je Tabelle (Quelle == Ziel). 0 = alle gleich, 1 = Abweichung.
verify_parity() { # tables(csv)
  local t src_c tgt_c tarr fail=0
  IFS=',' read -ra tarr <<< "$1"
  for t in "${tarr[@]}"; do
    src_c="$(sqlite_src "SELECT count(*) FROM \"$t\";" 2>/dev/null || echo ERR)"
    tgt_c="$(pg_psql "SELECT count(*) FROM \"$t\";" 2>/dev/null || echo ERR)"
    if [ "$src_c" != ERR ] && [ "$src_c" = "$tgt_c" ]; then
      log "  ✔ ${t}: ${tgt_c} Zeilen (Parität)"
    else
      warn "  ✘ ${t}: Quelle=${src_c} != Ziel=${tgt_c}"; fail=1
    fi
  done
  return "$fail"
}

# Duplikatfreiheit im Ziel: COUNT(DISTINCT col) == COUNT(*). 0 = ok.
verify_no_dup() { # table column
  local tbl="$1" col="$2" c d
  c="$(pg_psql "SELECT count(*) FROM \"$tbl\";" 2>/dev/null || echo ERR)"
  d="$(pg_psql "SELECT count(DISTINCT $col) FROM \"$tbl\";" 2>/dev/null || echo ERR)"
  if [ "$c" != ERR ] && [ "$c" = "$d" ]; then
    log "  ✔ ${tbl}: COUNT(DISTINCT ${col})=${d} == COUNT(*)=${c} (keine Duplikate)"; return 0
  fi
  warn "  ✘ ${tbl}: COUNT(DISTINCT ${col})=${d} != COUNT(*)=${c} — Duplikate!"; return 1
}

require_source() {
  [ -n "${SQLITE_DB:-}" ] || die "SQLITE_DB nicht gesetzt (Host-Pfad der Quell-SQLite)"
  [ -f "$SQLITE_DB" ]      || die "SQLITE_DB existiert nicht: $SQLITE_DB"
}
require_target() {
  [ -n "${PG_DSN:-}" ] || die "PG_DSN nicht gesetzt (Ziel-Postgres-DSN)"
}

# --- doctor: profile-unabhängiger Pre-Flight ------------------------------
cmd_doctor() {
  require_source
  require_target
  local rc=0

  # 1) Tooling: d-migrate-Image auflösbar + Container lauffähig.
  local img; img="$(resolve_dmigrate_image)"
  [ -n "$img" ] || die "DMIGRATE_IMAGE nicht auflösbar"
  log "d-migrate-Image: $img"
  local ver
  if ver="$(docker run --rm "$img" --version 2>/dev/null)"; then
    log "  ✔ d-migrate-Container lauffähig (${ver})"
  else
    warn "  ✘ d-migrate-Container startet nicht (Image gepullt?)"; rc=3
  fi

  # 2) Quelle: für den d-migrate-User (uid) lesbar UND read-write öffenbar.
  #    d-migrate/HikariCP öffnet SQLite RW — ein Read-Check (oder als root) würde
  #    das SQLITE_READONLY nicht fangen, an dem profile/bulk dann sterben. Daher
  #    denselben User + eine echte, zurückgerollte Write-Probe (BEGIN IMMEDIATE
  #    allein reicht NICHT: SQLite defert den Write-Fehler bis zum Statement).
  local src_dir base
  src_dir="$(cd "$(dirname "$SQLITE_DB")" && pwd)"
  base="$(basename "$SQLITE_DB")"
  local integ rwerr
  integ="$(docker run --rm --user "$(id -u):$(id -g)" --entrypoint sqlite3 \
    -v "${src_dir}:/work" "$SQLITE_IMAGE" "/work/${base}" 'PRAGMA integrity_check;' 2>/dev/null || true)"
  if [ "$integ" != "ok" ]; then
    warn "  ✘ Quell-SQLite nicht lesbar / integrity_check != ok ('$integ')"; rc=3
  else
    rwerr="$(docker run --rm --user "$(id -u):$(id -g)" --entrypoint sqlite3 \
      -v "${src_dir}:/work" "$SQLITE_IMAGE" "/work/${base}" \
      'BEGIN; CREATE TABLE IF NOT EXISTS __cutover_rwprobe__(x); ROLLBACK;' 2>&1 >/dev/null || true)"
    if [ -z "$rwerr" ]; then
      log "  ✔ Quell-SQLite lesbar + read-write öffenbar (uid $(id -u)): $SQLITE_DB"
    else
      warn "  ✘ Quell-SQLite nicht read-write für uid $(id -u) — d-migrate öffnet RW ('${rwerr}'); Quelle + Verzeichnis müssen schreibbar sein"; rc=3
    fi
  fi

  # 3) Ziel: PG erreichbar.
  if [ "$(pg_psql 'SELECT 1;' 2>/dev/null || true)" = "1" ]; then
    log "  ✔ Ziel-PG erreichbar"
  else
    warn "  ✘ Ziel-PG nicht erreichbar (PG_DSN/PG_NETWORK prüfen)"
    die "Pre-Flight abgebrochen: Ziel-PG nicht erreichbar" 3
  fi

  # 4) Ziel-Schema vorhanden (PG-DDL ADR-0006, vor dem Cutover angewendet).
  local ntables
  ntables="$(pg_psql "SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE';" 2>/dev/null || echo 0)"
  if [ "${ntables:-0}" -ge "$EXPECT_TABLES" ]; then
    log "  ✔ Ziel-Schema vorhanden (${ntables} Tabellen ≥ ${EXPECT_TABLES})"
  else
    warn "  ✘ Ziel-Schema unvollständig (${ntables} < ${EXPECT_TABLES}) — erst PG-DDL (ADR-0006) anwenden (migrations/postgres/)"; rc=3
  fi

  # 5) Ziel leer (Precondition für Bulk mit --on-conflict abort).
  local nevents
  nevents="$(pg_psql 'SELECT count(*) FROM playback_events;' 2>/dev/null || echo '?')"
  if [ "$nevents" = "0" ]; then
    log "  ✔ Ziel leer (playback_events=0) — bereit für Bulk (--on-conflict abort)"
  elif [ "$nevents" = "?" ]; then
    warn "  ? playback_events nicht abfragbar (Schema unvollständig?)"; rc=3
  else
    warn "  ! Ziel NICHT leer (playback_events=${nevents}) — Bulk mit --on-conflict abort würde brechen; inkrementell/--on-conflict skip nötig"
  fi

  if [ "$rc" -eq 0 ]; then
    log "doctor: Pre-Flight grün — Quelle + Ziel-PG bereit (Phase 0 'data profile' bleibt separat, s. profile)."
  else
    log "doctor: Pre-Flight-Befund (Exit 3) — siehe WARN oben."
  fi
  return "$rc"
}

# --- Phase 0: data profile (Pre-Flight, Toleranz self-type-only) -----
# `data profile` ist source-only (kein --target). Toleranz-Politik:
# Abbruch nur bei (a) Profile-Exit != 0 (Crash/Quelle unlesbar) oder (b) einer
# Spalte, deren self-type-Eintrag (targetType == logicalType) incompatibleCount
# > 0 hat. Cross-Type-Warnings / Null / leere Tabellen sind Info, nie Abbruch.
cmd_profile() {
  require_source
  command -v python3 >/dev/null 2>&1 || die "python3 nötig für die Profile-Auswertung"
  local base outdir
  base="$(basename "$SQLITE_DB")"
  # Report in ein separates Temp-Verzeichnis (Mount /out) schreiben, NICHT ins
  # Live-Quell-Verzeichnis — sonst landet profile.json neben der Produktions-DB.
  outdir="$(mktemp -d)"
  log "Phase 0: data profile (Quelle) — Toleranz self-type-only"
  if ! RUN_DMIGRATE_EXTRA_MOUNT="${outdir}:/out" run_dmigrate data profile \
       --source "sqlite:///work/${base}" --format json --output /out/profile.json \
       >"${outdir}/profile.err" 2>&1; then
    grep -viE 'HikariConfig|idleTimeout' "${outdir}/profile.err" >&2 || true
    rm -rf "$outdir"
    die "data profile fehlgeschlagen — (a). Hinweis: d-migrate öffnet SQLite read-write; die Quelle (+ ihr Verzeichnis) muss für den d-migrate-Container-User (uid $(id -u)) beschreibbar sein." 3
  fi
  if [ ! -f "${outdir}/profile.json" ]; then
    rm -rf "$outdir"; die "data profile erzeugte keinen Report" 3
  fi
  # Auswertung über tables[].columns[].targetCompatibility[].
  if ! PROFILE_JSON="${outdir}/profile.json" python3 - <<'PY'
import json, os, sys
d = json.load(open(os.environ["PROFILE_JSON"]))
fails, cross, empties, no_self = [], 0, [], []
for t in d.get("tables", []):
    if t.get("rowCount", 0) == 0:
        empties.append(t["name"])
    for c in t.get("columns", []):
        lt = c.get("logicalType")
        tc = c.get("targetCompatibility", []) or []
        self_e = next((e for e in tc if e.get("targetType") == lt), None)
        if self_e is None:
            no_self.append(f'{t["name"]}.{c["name"]} (logicalType={lt})')
            continue
        if self_e.get("incompatibleCount", 0) > 0:
            fails.append(f'{t["name"]}.{c["name"]}: {self_e["incompatibleCount"]} Wert(e) nicht nach {lt} abbildbar')
        cross += sum(1 for e in tc if e.get("targetType") != lt and e.get("incompatibleCount", 0) > 0)
print(f'[cutover] profile: {len(d.get("tables",[]))} Tabellen, {len(empties)} leer, '
      f'{cross} Cross-Type-Warnings (Info)')
if no_self:
    print(f'[cutover] profile: {len(no_self)} Spalte(n) ohne self-type-Eintrag (Info) — '
          f'Struktur via DDL/Drift-Check garantiert', file=sys.stderr)
if fails:
    print("[cutover] FAIL (b): self-type-Inkompatibilität — Werte nicht in ihren eigenen Zieltyp abbildbar:", file=sys.stderr)
    for f in fails:
        print(f'   - {f}', file=sys.stderr)
    sys.exit(1)
print("[cutover] profile: OK — keine self-type-Inkompatibilität (Tripwire still bei Gesundheit).")
PY
  then
    rm -rf "$outdir"
    die "Phase 0 Abbruch — self-type-Inkompatibilität (b), siehe oben" 3
  fi
  rm -rf "$outdir"
  log "Phase 0 (profile) grün."
}

# --- Phase 1: Bulk-Transfer + Sequenz-Erhalt ------------------------------
# `data transfer --sqlite-autoincrement-width 64` überträgt alle App-Tabellen
# und setzt die PG-BIGSERIAL-Zählerstände auf den SQLite-MAX fort (verifiziert:
# kein Neustart bei 1). schema_migrations + sqlite_%-Interna werden NICHT
# transferiert (auf beiden Seiten vom Migrations-Runner verwaltet).
cmd_bulk() {
  require_source; require_target
  local base tables wm out fail=0
  base="$(basename "$SQLITE_DB")"
  tables="$(app_tables 2>/dev/null || echo '')"
  [ -n "$tables" ] || die "keine transferierbaren Tabellen in der Quelle gefunden" 1
  log "Phase 1 (bulk): data transfer (--on-conflict abort) — Tabellen: ${tables}"
  out="$(mktemp)"
  if ! run_dmigrate data transfer --source "sqlite:///work/${base}" --target "$PG_DSN" \
       --tables "$tables" --sqlite-autoincrement-width 64 --on-conflict abort \
       --chunk-size "$CHUNK_SIZE" >"$out" 2>&1; then
    grep -viE 'HikariConfig|idleTimeout' "$out" >&2 || true
    rm -f "$out"
    die "data transfer fehlgeschlagen (Ziel nicht leer? DSN/Netz? siehe oben)" 1
  fi
  rm -f "$out"
  # Watermark für die inkrementelle Phase festhalten.
  wm="$(sqlite_src "SELECT COALESCE(MAX(ingest_sequence),0) FROM playback_events;" 2>/dev/null || echo '?')"
  log "Watermark (max ingest_sequence der Quelle): ${wm}"
  # Verifikation: Row-Count-Parität + Sequenz-Erhalt (kein PK-Kollision).
  verify_parity "$tables" || fail=1
  verify_sequence playback_events ingest_sequence || fail=1
  verify_sequence srt_health_samples id || fail=1
  [ "$fail" -eq 0 ] || die "Bulk-Verifikation fehlgeschlagen — siehe ✘ oben" 1
  log "Phase 1 (bulk) grün — Row-Count-Parität + Sequenz-Erhalt bestätigt. Watermark=${wm}."
}

# --- Phase 2: Inkrementelles Nachziehen (Delta seit Bulk) -----------------
# `--since-column ingest_sequence` filtert die High-Volume-Tabelle
# playback_events (nur Rows > SINCE); Tabellen OHNE ingest_sequence werden voll
# gescannt und per `--on-conflict skip` PK-basiert dedupliziert (idempotent).
# SINCE: explizit via ENV, sonst aus dem Ziel-MAX(ingest_sequence) abgeleitet
# (Auto-Resume vom Ziel-Stand). Wiederholbar. Der konservative Lookback für den
# Out-of-order-Commit-Fall gehört in die quiescte Switch-Phase (cmd_switch).
cmd_incremental() {
  require_source; require_target
  local base tables since out fail=0 newwm
  base="$(basename "$SQLITE_DB")"
  tables="$(app_tables 2>/dev/null || echo '')"
  [ -n "$tables" ] || die "keine transferierbaren Tabellen in der Quelle gefunden" 1
  if [ -n "${SINCE:-}" ]; then
    since="$SINCE"
  else
    since="$(pg_psql "SELECT COALESCE(MAX(ingest_sequence),0) FROM playback_events;" 2>/dev/null || echo '')"
    [[ "$since" =~ ^[0-9]+$ ]] || die "SINCE nicht gesetzt und Ziel-MAX(ingest_sequence) nicht ermittelbar" 1
  fi
  log "Phase 2 (incremental): --since-column ingest_sequence --since ${since} --on-conflict skip"
  out="$(mktemp)"
  if ! run_dmigrate data transfer --source "sqlite:///work/${base}" --target "$PG_DSN" \
       --tables "$tables" --since-column ingest_sequence --since "$since" \
       --on-conflict skip --chunk-size "$CHUNK_SIZE" >"$out" 2>&1; then
    grep -viE 'HikariConfig|idleTimeout' "$out" >&2 || true
    rm -f "$out"
    die "data transfer (incremental) fehlgeschlagen — siehe oben" 1
  fi
  rm -f "$out"
  # Verifikation: Parität + Duplikatfreiheit (PK) + Sequenz-Erhalt.
  verify_parity "$tables" || fail=1
  verify_no_dup playback_events ingest_sequence || fail=1
  verify_no_dup srt_health_samples id || fail=1
  verify_sequence playback_events ingest_sequence || fail=1
  verify_sequence srt_health_samples id || fail=1
  [ "$fail" -eq 0 ] || die "Incremental-Verifikation fehlgeschlagen — siehe ✘ oben" 1
  newwm="$(pg_psql "SELECT COALESCE(MAX(ingest_sequence),0) FROM playback_events;" 2>/dev/null || echo '?')"
  log "Phase 2 (incremental) grün — Parität + duplikatfrei + Sequenz-Erhalt. Neues Watermark=${newwm}."
}

cmd_switch() {
  require_source; require_target
  warn "Phase 3 (switch) noch nicht implementiert (Folge-Arbeit)."
  # Writer quiescen → finales incremental mit konservativem Lookback →
  # MTRACE_PERSISTENCE=postgres → Verifikation (Row-Counts + Watermark) →
  # Rollback = zurück auf SQLite.
  return 4
}

usage() {
  # Header-Kommentar (nach dem Shebang bis zur ersten Nicht-Kommentarzeile) —
  # robust gegen Header-Längenänderung statt fixer Zeilenspanne.
  awk 'NR==1{next} /^#/{sub(/^# ?/,""); print; next} {exit}' "$0"
}

main() {
  local cmd="${1:-help}"
  case "$cmd" in
    doctor)      cmd_doctor ;;
    profile)     cmd_profile ;;
    bulk)        cmd_bulk ;;
    incremental) cmd_incremental ;;
    switch)      cmd_switch ;;
    help | -h | --help) usage ;;
    *) die "unbekanntes Subcommand '$cmd' (doctor|profile|bulk|incremental|switch|help)" 2 ;;
  esac
}

main "$@"
