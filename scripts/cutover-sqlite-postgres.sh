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
#   profile — Phase 0 (data profile): Quality-Warnings + Target-Typ-Kompatibilität
#     der Quelle. Noch nicht verdrahtet (echter Aufruf ersetzt den Stub, s. cmd_profile).
#   bulk — Phase 1 (data transfer, Erstübertragung). Stub (Folge-Arbeit).
#   incremental — Phase 2 (data transfer --since, Delta). Stub (Folge-Arbeit).
#   switch — Phase 3 (Quiesce → finales Delta → Umschalten). Stub (Folge-Arbeit).
#
# ENV:
#   SQLITE_DB Host-Pfad der Quell-SQLite (z. B. /var/lib/mtrace/m-trace.db).
#   PG_DSN Ziel-Postgres-DSN (postgres://user:pass@host:port/db?sslmode=disable).
#   PG_NETWORK Optional: Docker-Netz, dem die Client-/d-migrate-Container
#                    beitreten, wenn der DSN-Host ein Container-Name ist.
#   DMIGRATE_IMAGE Override; Default = Single-Source-Pin aus apps/api/Makefile
#                    (`make -C apps/api -s print-dmigrate-image`).
#   SQLITE_IMAGE sqlite3-Client-Image (Default keinos/sqlite3).
#   PG_CLIENT_IMAGE psql-Client-Image (Default postgres:17-alpine).
#   CHUNK_SIZE data-transfer-Chunkgröße (Default 10000).
#
# Exit-Codes: 0 ok · 2 Config-/Nutzungsfehler · 3 Pre-Flight-Befund
#             (Ziel nicht bereit) · 4 noch nicht implementiert (Stub/blockiert).

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
SQLITE_IMAGE="${SQLITE_IMAGE:-keinos/sqlite3}"
PG_CLIENT_IMAGE="${PG_CLIENT_IMAGE:-postgres:17-alpine}"
CHUNK_SIZE="${CHUNK_SIZE:-10000}"
# Erwartetes Ziel-Schema (PG-DDL, ADR-0006): 13 Tabellen.
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
  mapfile -t net < <(docker_net_args)
  docker run --rm --user "$(id -u):$(id -g)" \
    -v "${src_dir}:/work" -w /work "${net[@]}" "$img" "$@"
}

# psql gegen das Ziel-PG (Client-Container). Gibt tuple-only aus.
pg_psql() {
  mapfile -t net < <(docker_net_args)
  docker run --rm -i "${net[@]}" "$PG_CLIENT_IMAGE" \
    psql "$PG_DSN" -tA -c "$1"
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
  if docker run --rm "$img" --version >/dev/null 2>&1; then
    log "  ✔ d-migrate-Container lauffähig ($(docker run --rm "$img" --version 2>/dev/null))"
  else
    warn "  ✘ d-migrate-Container startet nicht (Image gepullt?)"; rc=3
  fi

  # 2) Quelle: SQLite lesbar (integrity_check, ohne d-migrate).
  local src_dir base
  src_dir="$(cd "$(dirname "$SQLITE_DB")" && pwd)"
  base="$(basename "$SQLITE_DB")"
  local integ
  integ="$(docker run --rm --user 0:0 --entrypoint sqlite3 \
    -v "${src_dir}:/work" "$SQLITE_IMAGE" "/work/${base}" 'PRAGMA integrity_check;' 2>/dev/null || true)"
  if [ "$integ" = "ok" ]; then
    log "  ✔ Quell-SQLite lesbar (integrity_check ok): $SQLITE_DB"
  else
    warn "  ✘ Quell-SQLite nicht lesbar / integrity_check != ok ('$integ')"; rc=3
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

# --- Stubs: die d-migrate-getriebenen Phasen ------------------------------
cmd_profile() {
  require_source; require_target
  warn "Phase 0 (data profile) noch nicht verdrahtet (Folge-Arbeit)."
  # d-migrate 0.9.11 fixt die Integer-/Leer-Spalten-Crashes (verifiziert) —
  # der Pre-Flight kann jetzt gebaut werden. Vorgesehener Aufruf:
  #   run_dmigrate data profile \
  #     --source "sqlite:///work/$(basename "$SQLITE_DB")" \
  #     --format json --output /work/profile.json
  #   Danach targetCompatibility je Spalte auswerten (incompatibleCount>0 =>
  #   Abbruch); leere Tabellen sind unkritisch (nonNullCount=0).
  return 4
}

cmd_bulk() {
  require_source; require_target
  warn "Phase 1 (bulk) noch nicht implementiert (Folge-Arbeit)."
  # Vorgesehener Aufruf (data transfer ist bereits ausführungs-verifiziert):
  #   run_dmigrate data transfer \
  #     --source "sqlite:///work/$(basename "$SQLITE_DB")" \
  #     --target "$PG_DSN" \
  #     --sqlite-autoincrement-width 64 --on-conflict abort --chunk-size "$CHUNK_SIZE"
  #   Danach Watermark festhalten: SELECT MAX(ingest_sequence) FROM playback_events.
  return 4
}

cmd_incremental() {
  require_source; require_target
  warn "Phase 2 (incremental) noch nicht implementiert (Folge-Arbeit)."
  # Vorgesehener Aufruf:
  #   run_dmigrate data transfer \
  #     --source "sqlite:///work/$(basename "$SQLITE_DB")" --target "$PG_DSN" \
  #     --since-column ingest_sequence --since "$WATERMARK" --on-conflict skip
  return 4
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
  sed -n '2,29p' "$0" | sed 's/^# \{0,1\}//'
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
