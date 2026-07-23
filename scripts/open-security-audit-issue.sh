#!/usr/bin/env bash
# ( extra-gates.md) —
# Eröffnet ein GitHub-Issue, wenn der Nightly-Security-Audit
# (govulncheck / pnpm audit / Trivy image scan) eine neue
# CRITICAL/HIGH-Vulnerability meldet. Aufgerufen aus
# `.github/workflows/security-audit.yml::Open security audit issue`.
#
# Hintergrund: Die `Security gates` im PR-/Push-Workflow brechen
# nur bei Pushes; zwischen zwei Pushes neu veröffentlichte
# Advisories (Beispiel: GHSA-77vg-94rm-hx3p devalue, 2026-05-17)
# bleiben unbemerkt, bis jemand pusht. Der Nightly schließt diese
# Lücke täglich.
#
# Erwartete Inputs:
# .tmp/security/vuln-check.log stdout/stderr von `make vuln-check`
# .tmp/security/audit-ts.log stdout/stderr von `make audit-ts`
# .tmp/security/image-scan.log stdout/stderr von `make image-scan`
# Env vars (gesetzt vom Workflow):
# VULN_CHECK_OUTCOME success|failure (steps.<id>.outcome)
# AUDIT_TS_OUTCOME success|failure
# IMAGE_SCAN_OUTCOME success|failure
# GH_TOKEN GitHub-Token mit `issues:write`
# RUN_URL Link auf den fehlgeschlagenen Workflow-Run
set -eu

title="Security audit findings ($(date -u +%Y-%m-%d))"

# Pro Check ein zusammengefasster Block. Die jeweils letzten
# 40 Zeilen reichen, um die eigentliche Vulnerability-Tabelle plus
# Exit-Trailer einzufangen, ohne dass das Issue-Body-Limit gerissen
# wird. Vollständige Logs liegen im Workflow-Artefakt.
extract_tail() {
  local file=$1
  if [ -f "$file" ]; then
    tail -n 40 "$file"
  else
    echo "<no log captured>"
  fi
}

vuln_tail="$(extract_tail .tmp/security/vuln-check.log)"
audit_tail="$(extract_tail .tmp/security/audit-ts.log)"
image_tail="$(extract_tail .tmp/security/image-scan.log)"

body=$(cat <<EOF
The nightly security audit workflow detected one or more findings.
Pro Pushes prueft \`build.yml::Security gates\` denselben Stack,
aber zwischen Pushes neu veroeffentlichte Advisories bleiben sonst
unbemerkt — dieser Nightly schliesst die Luecke (Hintergrund:
\`scripts/open-security-audit-issue.sh\` Header).

- Workflow run: ${RUN_URL}
- govulncheck (apps/api):       \`${VULN_CHECK_OUTCOME:-skipped}\`
- pnpm audit (TS workspace):    \`${AUDIT_TS_OUTCOME:-skipped}\`
- Trivy image scan (3 images):  \`${IMAGE_SCAN_OUTCOME:-skipped}\`

Full step output is attached as workflow artifact
\`security-audit-${RUN_URL##*/}\` (\`.tmp/security/*.log\`); the
tails below show the relevant Vulnerability-Tabelle.

## govulncheck (apps/api) — tail

\`\`\`
${vuln_tail}
\`\`\`

## pnpm audit (TS workspace) — tail

\`\`\`
${audit_tail}
\`\`\`

## Trivy image scan — tail

\`\`\`
${image_tail}
\`\`\`

## Reaction

1. Identify the failing check(s) above and read the full log in the
   artifact.
2. **govulncheck**: bump the offending Go dependency in
   \`apps/api/go.mod\`, run \`make vuln-check\` locally to confirm.
3. **pnpm audit**: either bump the offending package or add a
   \`pnpm.overrides\` entry in the root \`package.json\` (same
   pattern as \`picomatch\`/\`devalue\`). Re-run
   \`make lock-refresh && make audit-ts\` locally.
4. **Trivy image scan**: identify the OS package or layer responsible
   in the offending Dockerfile and bump the base image / package
   version. If a finding is a knowingly accepted risk, add an entry
   to \`.security/vulnignore.yaml\` with an \`expires\` date
   (max 30 days) and a justification — \`make image-scan\` regenerates
   the per-image \`.trivyignore\` automatically.
5. Push the fix; the next Nightly verifies that the gate is green
   again. If the issue stays open beyond 7 days, escalate via the
   Tranche-3 risks backlog (\`docs/plan/planning/in-progress/risks-backlog.md\`).
EOF
)

gh issue create \
  --title "$title" \
  --body "$body" \
  --label "security,audit"
