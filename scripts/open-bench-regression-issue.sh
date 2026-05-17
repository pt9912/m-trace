#!/usr/bin/env bash
# plan-0.9.5 §3 Tranche 2 (RAK-Wave-2 / extra-gates.md §3.3) —
# Eröffnet ein GitHub-Issue, wenn der Nightly-Benchstat-Vergleich
# eine statistisch signifikante Regression (+15% bei p<0.05)
# meldet. Aufgerufen aus
# `.github/workflows/benchmark.yml::Open regression issue`.
#
# Erwartete Inputs:
#   - .tmp/bench/comparison.txt mit der benchstat-Ausgabe
#     (wird vom vorhergehenden `Compare with benchstat`-Step
#      erzeugt; fehlt sie, landet `<comparison missing>` im Body)
#   - $GH_TOKEN  GitHub-Token mit `issues:write`
#   - $RUN_URL   Link auf den fehlgeschlagenen Workflow-Run
set -eu

title="Benchmark regression detected ($(date -u +%Y-%m-%d))"
comparison="$(cat .tmp/bench/comparison.txt 2>/dev/null || echo '<comparison missing>')"

body=$(cat <<EOF
The nightly benchmark workflow (\`make api-benchmark-smoke\`
equivalent with \`-count=10\`) detected one or more statistically
significant regressions over the +15% threshold.

- Workflow run: ${RUN_URL}
- Threshold: +15% at p<0.05
- Baseline source: orphan-branch benchmark-baseline / benchmarks/api-bench.txt

## benchstat comparison

\`\`\`
${comparison}
\`\`\`

## Reaction

1. Open the run logs and read the regression detail
   (\`scripts/check-benchstat-regression.mjs\` output names every
   regressed Benchmark plus its delta and p-value).
2. Reproduce locally:
   \`\`\`
   cd apps/api
   go test -bench=. -benchmem -count=10 -benchtime=2s ./hexagon/... ./adapters/...
   \`\`\`
3. Either fix the regression and re-run the workflow, or
   accept the new performance characteristic and push the
   current bench output to the \`benchmark-baseline\` orphan
   branch (drift-acceptance path).
EOF
)

gh issue create \
  --title "$title" \
  --body "$body" \
  --label "performance,benchmark,plan-0.9.5"
