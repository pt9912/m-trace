#!/usr/bin/env bash
#
# generate-postgres-schema.sh — erzeugt das PostgreSQL-Baseline-DDL
# (V1__m_trace.sql) für den optionalen Postgres-Runtime-Adapter
# (ADR-0006) aus `schema.yaml`.
#
# Seit dem schema.yaml-Refold (rolling-V1-Rekonsolidierung, 2026-07) ist
# `schema.yaml` die Single-Source-of-Truth für den vollen 13-Tabellen-
# Stand — symmetrisch zu `make schema-generate` (SQLite) wird das PG-DDL
# direkt via `export flyway --target postgresql --source schema.yaml`
# erzeugt. Kein Reverse der Live-SQLite mehr nötig (dieser Umweg
# existierte nur, solange schema.yaml V1-only war).
#
# Pipeline (rein d-migrate-nativ):
#   1. `export flyway --target postgresql --version 1`: erzeugt das
#      PG-DDL byte-deterministisch aus schema.yaml (kein Timestamp im
#      Header, nur der Versions-Stempel).
#
# Verifikation (fail-loud, siehe verify_output + verify_column_parity):
# 13 Tabellen, 18 CHECK-Constraints klammer-balanciert, beide 64-bit-PKs
# (BIGSERIAL, aus schema.yaml `--sqlite-autoincrement-width 64`-Reverse
# stammend), plus Cross-Dialekt-Spaltengleichheit gegen das eingecheckte
# SQLite-V1 (fängt target-spezifische Export-Divergenz).
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
: "${DMIGRATE_IMAGE:?DMIGRATE_IMAGE muss gesetzt sein (vom Makefile durchgereicht)}"

# Pfade relativ zu apps/api (Makefile-CURDIR).
STORAGE_DIR="internal/storage"
SCHEMA_YAML="${STORAGE_DIR}/schema.yaml"
MIGRATIONS_DIR="${STORAGE_DIR}/migrations"
SQLITE_V1="${MIGRATIONS_DIR}/V1__m_trace.sql"
PG_DIR="${MIGRATIONS_DIR}/postgres"
PG_FILE="${PG_DIR}/V1__m_trace.sql"

# Erwartete Invarianten (fail-loud-Guards).
EXPECT_TABLES=13
EXPECT_CHECKS=18
# Die zwei 64-bit-AUTOINCREMENT-PKs. Beide müssen im PG-DDL als bigint
# mit Identity/Serial erscheinen — nie als 32-bit SERIAL/INT.
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

# --- 1) export flyway --target postgresql aus schema.yaml ------------------
generate_ddl() {
  [[ -f "$SCHEMA_YAML" ]] || die "schema.yaml fehlt: $SCHEMA_YAML (aus apps/api aufrufen)"
  cp "$SCHEMA_YAML" "$WORK/schema.yaml"
  mkdir -p "$WORK/out"
  docker run --rm --user "$UID_GID" -v "$WORK:/work" -w /work \
    "$DMIGRATE_IMAGE" export flyway \
      --source /work/schema.yaml --target postgresql \
      --version 1 --output /work/out \
    || die "export flyway fehlgeschlagen"
  [[ -f "$WORK/out/V1__m_trace.sql" ]] || die "export erzeugte keine V1__m_trace.sql"
}

# --- Verifikation (fail-loud) ---------------------------------------------
verify_output() {
  local ddl="$1"
  local tables
  tables="$(grep -c '^CREATE TABLE ' "$ddl" || true)"
  [[ "$tables" -eq "$EXPECT_TABLES" ]] || die "erwartete $EXPECT_TABLES Tabellen, fand $tables"
  local checks
  checks="$(grep -c 'CHECK (' "$ddl" || true)"
  [[ "$checks" -eq "$EXPECT_CHECKS" ]] || die "erwartete $EXPECT_CHECKS CHECK-Constraints, fand $checks"
  local unbalanced
  unbalanced="$(grep 'CHECK (' "$ddl" | awk '{
    line=$0; o=gsub(/\(/,"("); c=gsub(/\)/,")"); if (o!=c) bad++
  } END { print bad+0 }')"
  [[ "$unbalanced" -eq 0 ]] || die "$unbalanced CHECK-Expression(s) klammer-unbalanciert"
  local col
  for col in "${IDENTITY_COLUMNS[@]}"; do
    local name="${col#*.}"
    grep -Eq "\"${name}\" (BIGSERIAL|BIGINT)" "$ddl" \
      || die "$col nicht als bigint gerendert (int32-Verengung?)"
  done
  log "Verifikation ok: $tables Tabellen, $checks CHECKs (balanciert), 64-bit-PKs intakt"
}

# Cross-Dialekt-Spaltengleichheit: das generierte PG-DDL und das
# eingecheckte SQLite-V1 stammen beide aus schema.yaml; ihre Tabellen-
# und Spaltenmengen (Namen + Reihenfolge) müssen übereinstimmen. Fängt
# target-spezifische Export-Divergenz (analog zur früheren Parität gegen
# die Live-SQLite, jetzt gegen die zweite Dialekt-Ableitung).
verify_column_parity() {
  local pg_ddl="$1"
  [[ -f "$SQLITE_V1" ]] || die "$SQLITE_V1 fehlt — erst 'make schema-generate' laufen lassen"
  python3 - "$SQLITE_V1" "$pg_ddl" <<'PY' || die "Cross-Dialekt-Spalten-Parität verletzt (SQLite-V1 vs PG-DDL)"
import sys, re
def parse(path):
    text = open(path).read()
    cols = {}
    for m in re.finditer(r'CREATE TABLE "(\w+)" \((.*?)\n\);', text, re.S):
        t, body = m.group(1), m.group(2)
        out = []
        for ln in body.split('\n'):
            ln = ln.strip().rstrip(',')
            cm = re.match(r'"(\w+)"\s', ln)
            if cm and not re.match(r'(CONSTRAINT|PRIMARY KEY|FOREIGN KEY|UNIQUE|CHECK)\b', ln):
                out.append(cm.group(1))
        cols[t] = out
    return cols
sqlite_cols = parse(sys.argv[1])
pg_cols = parse(sys.argv[2])
bad = 0
for t in sorted(set(sqlite_cols) | set(pg_cols)):
    if sqlite_cols.get(t, []) != pg_cols.get(t, []):
        bad += 1
        print(f"  [DIFF] {t}: SQLite={sqlite_cols.get(t)} PG={pg_cols.get(t)}", file=sys.stderr)
sys.exit(1 if bad else 0)
PY
  log "Spalten-Parität ok: PG-DDL und SQLite-V1 sind spalten- und reihenfolgegleich"
}

# --- Ablauf ----------------------------------------------------------------
main() {
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
