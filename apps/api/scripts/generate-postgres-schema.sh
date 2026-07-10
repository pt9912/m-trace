#!/usr/bin/env bash
#
# generate-postgres-schema.sh — erzeugt das PostgreSQL-Baseline-DDL
# (V1__m_trace.sql) für den optionalen Postgres-Runtime-Adapter
# (ADR-0006) aus dem *live* SQLite-Schema.
#
# Warum nicht aus schema.yaml? schema.yaml ist V1-only (5 Tabellen);
# der Runtime-Store ist V1+V2–V7 (13 Tabellen, inkl. Rebuild-/ALTER-
# Migrationen). Der einzig driftfreie Weg zum vollständigen PG-Schema
# ist, die live migrierte SQLite zu introspizieren und nach Postgres
# zu exportieren — automatisiert, ohne Hand-Portage, ohne schema.yaml-
# Kollision.
#
# Pipeline (rein d-migrate-nativ, kein YAML-Nachpatchen):
#   1. live.db bauen: V1..V7 numerisch geordnet via SQLite-CLI anwenden
#      (derselbe Flyway-Reihenfolgen-Kontrakt wie storage/migrate.go).
#   2. `schema reverse --sqlite-autoincrement-width 64`: introspiziert
#      live.db in ein neutrales Schema. width=64 rendert SQLites 64-bit
#      AUTOINCREMENT-PKs (playback_events.ingest_sequence,
#      srt_health_samples.id) getreu als biginteger+Identity statt als
#      32-bit-SERIAL (ADR-0027 in d-migrate). Nativer
#      Output ist BIGSERIAL — semantisch die by_default-Identity
#      (bigint, explizite Inserts erlaubt, implizite nextval-Sequence).
#   3. `export flyway --target postgresql --version 1`: erzeugt das
#      PG-DDL byte-deterministisch (kein Timestamp im Header).
#
# Verifikation (fail-loud, siehe verify_output): 13 Tabellen, 18 CHECK-
# Constraints klammer-balanciert, beide 64-bit-PKs, Spaltengleichheit
# gegen die live-SQLite.
#
# Modi:
#   (default) schreibt internal/storage/migrations/postgres/V1__m_trace.sql
#   --check generiert nach temp und difft gegen die eingecheckte Datei
#     und gibt Exit 1 bei Drift (für das PG-DDL-Drift-Gate).
#
# Aufruf über `make schema-generate-postgres` bzw. `… -check`.
# DMIGRATE_IMAGE wird vom Makefile durchgereicht (Single-Source-Pin).

set -euo pipefail

# --- Konfiguration ---------------------------------------------------------
# d-migrate-Image: aus der Umgebung (Makefile-Pin). Kein Default, damit
# der Pin nicht still divergiert.
: "${DMIGRATE_IMAGE:?DMIGRATE_IMAGE muss gesetzt sein (vom Makefile durchgereicht)}"

# SQLite-CLI-Image zum Bauen der live.db. Per Digest gepinnt (analog
# d-migrate-/Trivy-/govulncheck-Pin), damit die live-SQLite und damit
# das PG-DDL reproduzierbar bleiben. keinos/sqlite3 hat keinen
# versionierten Tag, daher Digest-Pin (entspricht SQLite 3.53.0).
SQLITE_IMAGE="${SQLITE_IMAGE:-keinos/sqlite3@sha256:252363ef3cbbe11f1100dcbc734b89969b264df99a49008b34ca4578f503ff2a}"

# Pfade relativ zu apps/api (Makefile-CURDIR).
STORAGE_DIR="internal/storage"
MIGRATIONS_DIR="${STORAGE_DIR}/migrations"
PG_DIR="${MIGRATIONS_DIR}/postgres"
PG_FILE="${PG_DIR}/V1__m_trace.sql"

# Erwartete Invarianten (fail-loud-Guards).
EXPECT_TABLES=13
EXPECT_CHECKS=18
# Die zwei 64-bit-AUTOINCREMENT-PKs. Beide müssen im PG-DDL
# als bigint mit Identity/Serial erscheinen — nie als 32-bit SERIAL/INT.
IDENTITY_COLUMNS=("playback_events.ingest_sequence" "srt_health_samples.id")

MODE="write"
if [[ "${1:-}" == "--check" ]]; then
  MODE="check"
fi

UID_GID="$(id -u):$(id -g)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

log() { printf '[pg-schema] %s\n' "$*"; }
die() { printf '[pg-schema] FEHLER: %s\n' "$*" >&2; exit 1; }

# --- 1) live.db aus V1..V7 bauen ------------------------------------------
build_live_db() {
  [[ -d "$MIGRATIONS_DIR" ]] || die "Migrationsverzeichnis fehlt: $MIGRATIONS_DIR (aus apps/api aufrufen)"
  # Numerisch geordnete Migrationsliste (V1..V9, V10, … via sort -V) —
  # derselbe Kontrakt wie migrate.go (sort nach Integer-Version).
  local order
  order="$(cd "$MIGRATIONS_DIR" && ls V*.sql | sort -V)"
  [[ -n "$order" ]] || die "keine V*.sql-Migrationen gefunden"
  log "Migrationsreihenfolge: $(echo "$order" | tr '\n' ' ')"
  local order_oneline
  order_oneline="$(echo "$order" | tr '\n' ' ')"
  docker run --rm --user "$UID_GID" \
    -v "$(pwd)/${MIGRATIONS_DIR}:/src:ro" -v "$WORK:/work" -w /work \
    "$SQLITE_IMAGE" sh -c "for f in $order_oneline; do sqlite3 live.db < /src/\$f || exit 1; done" \
    || die "live.db-Aufbau fehlgeschlagen"
}

# --- 2) reverse + 3) export ------------------------------------------------
generate_ddl() {
  docker run --rm --user "$UID_GID" -v "$WORK:/work" -w /work \
    "$DMIGRATE_IMAGE" schema reverse \
      --source sqlite:///work/live.db --output /work/reverse.yaml \
      --name m-trace --version 1.0.0 \
      --sqlite-autoincrement-width 64 \
      --report /work/reverse.report.yaml \
    || die "schema reverse fehlgeschlagen"
  mkdir -p "$WORK/out"
  docker run --rm --user "$UID_GID" -v "$WORK:/work" -w /work \
    "$DMIGRATE_IMAGE" export flyway \
      --source /work/reverse.yaml --target postgresql \
      --version 1 --output /work/out \
    || die "export flyway fehlgeschlagen"
  [[ -f "$WORK/out/V1__m_trace.sql" ]] || die "export erzeugte keine V1__m_trace.sql"
}

# --- Verifikation (fail-loud) ---------------------------------------------
verify_output() {
  local ddl="$1"
  # 13 Tabellen
  local tables
  tables="$(grep -c '^CREATE TABLE ' "$ddl" || true)"
  [[ "$tables" -eq "$EXPECT_TABLES" ]] || die "erwartete $EXPECT_TABLES Tabellen, fand $tables"
  # 18 CHECK-Constraints, alle klammer-balanciert (der 0.9.9-Fix; mit
  # 0.9.8 wurden IN-Listen an der ersten ) trunkiert)
  local checks
  checks="$(grep -c 'CHECK (' "$ddl" || true)"
  [[ "$checks" -eq "$EXPECT_CHECKS" ]] || die "erwartete $EXPECT_CHECKS CHECK-Constraints, fand $checks"
  local unbalanced
  unbalanced="$(grep 'CHECK (' "$ddl" | awk '{
    line=$0; o=gsub(/\(/,"("); c=gsub(/\)/,")"); if (o!=c) bad++
  } END { print bad+0 }')"
  [[ "$unbalanced" -eq 0 ]] || die "$unbalanced CHECK-Expression(s) klammer-unbalanciert (Reverse-Trunkierung?)"
  # Beide 64-bit-PKs als bigint (BIGSERIAL oder BIGINT … IDENTITY), nie
  # als 32-bit SERIAL/INT.
  local col
  for col in "${IDENTITY_COLUMNS[@]}"; do
    local name="${col#*.}"
    grep -Eq "\"${name}\" (BIGSERIAL|BIGINT)" "$ddl" \
      || die "$col nicht als bigint gerendert (int32-Verengung?)"
  done
  log "Verifikation ok: $tables Tabellen, $checks CHECKs (balanciert), 64-bit-PKs intakt"
}

# Spaltengleichheit live-SQLite <-> PG-DDL (Reihenfolge + Namen), deckt
# insbesondere die V3/V5-Rebuild-Tabellen ab.
verify_column_parity() {
  local ddl="$1"
  docker run --rm --user "$UID_GID" -v "$WORK:/work" -w /work "$SQLITE_IMAGE" sh -c '
    for t in $(sqlite3 live.db "SELECT name FROM sqlite_master WHERE type=\"table\" AND name NOT LIKE \"sqlite_%\" ORDER BY name"); do
      cols=$(sqlite3 live.db "SELECT group_concat(name, \",\") FROM pragma_table_info(\"$t\")")
      echo "$t|$cols"
    done' > "$WORK/sqlite_cols.txt" || die "Spaltenexport aus live.db fehlgeschlagen"
  python3 - "$WORK/sqlite_cols.txt" "$ddl" <<'PY' || die "Spalten-Parität verletzt (live-SQLite vs PG-DDL)"
import sys, re
sqlite_cols = {}
for line in open(sys.argv[1]):
    line = line.strip()
    if not line or '|' not in line:
        continue
    t, cols = line.split('|', 1)
    sqlite_cols[t] = cols.split(',') if cols else []
pg = open(sys.argv[2]).read()
pg_cols = {}
for m in re.finditer(r'CREATE TABLE "(\w+)" \((.*?)\n\);', pg, re.S):
    t, body = m.group(1), m.group(2)
    cols = []
    for ln in body.split('\n'):
        ln = ln.strip().rstrip(',')
        cm = re.match(r'"(\w+)"\s', ln)
        if cm and not re.match(r'(CONSTRAINT|PRIMARY KEY|FOREIGN KEY|UNIQUE|CHECK)\b', ln):
            cols.append(cm.group(1))
    pg_cols[t] = cols
bad = 0
for t in sorted(set(sqlite_cols) | set(pg_cols)):
    if sqlite_cols.get(t, []) != pg_cols.get(t, []):
        bad += 1
        print(f"  [DIFF] {t}: SQLite={sqlite_cols.get(t)} PG={pg_cols.get(t)}", file=sys.stderr)
sys.exit(1 if bad else 0)
PY
  log "Spalten-Parität ok: alle Tabellen spalten- und reihenfolgegleich"
}

# --- Ablauf ----------------------------------------------------------------
main() {
  build_live_db
  generate_ddl
  local generated="$WORK/out/V1__m_trace.sql"
  verify_output "$generated"
  verify_column_parity "$generated"

  if [[ "$MODE" == "check" ]]; then
    [[ -f "$PG_FILE" ]] || die "$PG_FILE fehlt — erst 'make schema-generate-postgres' laufen lassen"
    if diff -u "$PG_FILE" "$generated"; then
      log "Drift-Check ok: $PG_FILE ist aktuell"
    else
      die "Drift: $PG_FILE weicht vom regenerierten PG-DDL ab (oben) — 'make schema-generate-postgres' + committen"
    fi
  else
    mkdir -p "$PG_DIR"
    cp "$generated" "$PG_FILE"
    log "DDL geschrieben: apps/api/$PG_FILE"
  fi
}

main
