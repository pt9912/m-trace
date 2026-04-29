// coverage-overview liest ein Go-Coverage-Profil (mode=atomic) und
// erzeugt eine self-contained index.html mit Total und Per-Datei-
// Tabelle. Verlinkt auf coverage.html (Source-Drill-Down von
// `go tool cover -html`) und coverage-func.txt (Per-Funktion-Text).
//
// Aufruf:
//
//	go run ./scripts/coverage-overview \
//	    -profile /out/coverage.out \
//	    -threshold 90 \
//	    -out /out/index.html
//
// Ohne externe Abhängigkeiten — der Profil-Parser ist hand-rolled.
// Format einer Zeile (nach `mode:`-Header):
//
//	<file>:<startLine>.<startCol>,<endLine>.<endCol> <statements> <count>
package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"os"
	"sort"
	"strconv"
	"strings"
)

type fileStats struct {
	File           string
	Statements     int
	Covered        int
	CoveragePct    float64
	CoverageBucket string // "ok" / "warn" / "fail"
}

type pageData struct {
	TotalPct       float64
	TotalStmts     int
	TotalCovered   int
	TotalBucket    string
	Threshold      float64
	Files          []fileStats
	ProfileBaseURL string // optional Trim-Prefix für File-Display
}

const (
	bucketOK   = "ok"
	bucketWarn = "warn"
	bucketFail = "fail"
)

func bucketFor(pct, threshold float64) string {
	switch {
	case pct >= threshold:
		return bucketOK
	case pct >= threshold*0.9:
		return bucketWarn
	default:
		return bucketFail
	}
}

func main() {
	profilePath := flag.String("profile", "/out/coverage.out", "path to go test coverprofile")
	outPath := flag.String("out", "/out/index.html", "destination for the overview HTML")
	threshold := flag.Float64("threshold", 90.0, "coverage threshold in percent (also colors the bars)")
	stripPrefix := flag.String("strip-prefix", "github.com/pt9912/m-trace/apps/api/", "trim this prefix from file paths in the table")
	flag.Parse()

	stats, totalC, totalS, err := parseProfile(*profilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coverage-overview: parse %s: %v\n", *profilePath, err)
		os.Exit(2)
	}

	files := make([]fileStats, 0, len(stats))
	for f, s := range stats {
		pct := percent(s.covered, s.statements)
		display := strings.TrimPrefix(f, *stripPrefix)
		files = append(files, fileStats{
			File:           display,
			Statements:     s.statements,
			Covered:        s.covered,
			CoveragePct:    pct,
			CoverageBucket: bucketFor(pct, *threshold),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].CoveragePct != files[j].CoveragePct {
			return files[i].CoveragePct < files[j].CoveragePct // schwächste zuerst
		}
		return files[i].File < files[j].File
	})

	totalPct := percent(totalC, totalS)
	page := pageData{
		TotalPct:     totalPct,
		TotalStmts:   totalS,
		TotalCovered: totalC,
		TotalBucket:  bucketFor(totalPct, *threshold),
		Threshold:    *threshold,
		Files:        files,
	}

	out, err := os.Create(*outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coverage-overview: create %s: %v\n", *outPath, err)
		os.Exit(2)
	}
	defer func() { _ = out.Close() }()

	tpl := template.Must(template.New("overview").Funcs(template.FuncMap{
		"pct1": func(v float64) string { return fmt.Sprintf("%.1f%%", v) },
		"int":  func(v float64) int { return int(v) },
	}).Parse(htmlTemplate))
	if err := tpl.Execute(out, page); err != nil {
		fmt.Fprintf(os.Stderr, "coverage-overview: render: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("coverage-overview: %.1f%% total → %s\n", totalPct, *outPath)
}

type profileStats struct {
	statements int
	covered    int
}

type blockKey struct {
	file  string
	span  string // "start.col,end.col" — eindeutig pro Block
}

type blockState struct {
	statements int
	covered    bool
}

// parseProfile liest ein Go-Coverage-Profil und liefert Per-Datei-
// Statement- und Covered-Counts plus die Total-Werte.
//
// Wichtig: `go test ./... -coverprofile=…` mergt mehrere
// Test-Pakete in eine Datei und schreibt für denselben Block
// teilweise mehrfach Einträge (einmal pro Test-Paket, das ihn
// instrumentiert hat). Der Parser dedupliziert über
// `(file, span)` und markiert den Block als gedeckt, sobald **eine**
// Beobachtung count>0 hatte.
func parseProfile(path string) (map[string]profileStats, int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	blocks := make(map[blockKey]*blockState)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			if strings.HasPrefix(line, "mode:") {
				continue
			}
			// Profile ohne Mode-Header: durchfallen.
		}
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}
		stmts, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}
		colon := strings.Index(parts[0], ":")
		if colon < 0 {
			continue
		}
		key := blockKey{
			file: parts[0][:colon],
			span: parts[0][colon+1:],
		}
		b := blocks[key]
		if b == nil {
			b = &blockState{statements: stmts}
			blocks[key] = b
		}
		if count > 0 {
			b.covered = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, 0, err
	}

	stats := make(map[string]profileStats)
	var totalCovered, totalStatements int
	for k, b := range blocks {
		s := stats[k.file]
		s.statements += b.statements
		if b.covered {
			s.covered += b.statements
		}
		stats[k.file] = s

		totalStatements += b.statements
		if b.covered {
			totalCovered += b.statements
		}
	}
	return stats, totalCovered, totalStatements, nil
}

func percent(covered, total int) float64 {
	if total == 0 {
		return 0
	}
	return 100.0 * float64(covered) / float64(total)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="de">
<head>
<meta charset="UTF-8">
<title>m-trace API — Coverage Overview</title>
<style>
  body { font: 14px/1.5 system-ui, -apple-system, sans-serif; max-width: 1100px; margin: 2em auto; padding: 0 1em; color: #1a1a1a; }
  h1 { margin: 0 0 0.4em; font-size: 1.6em; }
  .links { margin: 0.5em 0 1.5em; }
  .links a { margin-right: 1em; }
  .total { display: inline-block; padding: 0.5em 1em; border-radius: 8px; margin: 0.3em 0 1em; }
  .total .pct { font-size: 2em; font-weight: 700; font-variant-numeric: tabular-nums; }
  .total .meta { font-size: 0.9em; color: #4b5563; }
  .total.ok   { background: #ecfdf5; }
  .total.ok   .pct { color: #047857; }
  .total.warn { background: #fffbeb; }
  .total.warn .pct { color: #b45309; }
  .total.fail { background: #fef2f2; }
  .total.fail .pct { color: #b91c1c; }
  table { width: 100%; border-collapse: collapse; }
  th, td { text-align: left; padding: 0.5em 0.75em; border-bottom: 1px solid #e5e7eb; }
  th { background: #f3f4f6; font-weight: 600; font-size: 0.9em; }
  tr:hover { background: #f9fafb; }
  .pct-cell { font-variant-numeric: tabular-nums; text-align: right; min-width: 5em; }
  .stmts { font-variant-numeric: tabular-nums; color: #6b7280; }
  .bar { display: inline-block; width: 140px; height: 0.7em; background: #e5e7eb; border-radius: 4px; overflow: hidden; vertical-align: middle; margin-left: 0.5em; }
  .bar > span { display: block; height: 100%; transition: width 0.2s; }
  .bar.ok   > span { background: #10b981; }
  .bar.warn > span { background: #f59e0b; }
  .bar.fail > span { background: #ef4444; }
  code { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 0.95em; }
  .small { color: #6b7280; font-size: 0.85em; }
  .threshold { color: #6b7280; font-size: 0.85em; margin-left: 0.5em; }
</style>
</head>
<body>
<h1>m-trace API — Coverage Overview</h1>

<div class="total {{.TotalBucket}}">
  <span class="pct">{{pct1 .TotalPct}}</span>
  <span class="meta">{{.TotalCovered}} / {{.TotalStmts}} Statements gedeckt</span>
  <span class="threshold">Threshold: {{pct1 .Threshold}}</span>
</div>

<div class="links">
  <a href="coverage.html">Source-Drill-Down ↗</a>
  <a href="coverage-func.txt">Per-Funktion-Text ↗</a>
</div>

<p class="small">Schwächste Dateien zuerst. Coverage-Range: <code>./hexagon/...</code>, <code>./adapters/...</code>; <code>cmd/api</code> ist Wiring/Signal-Handling und bleibt absichtlich draußen (siehe <code>docs/quality.md</code> §3).</p>

<table>
  <thead>
    <tr><th>Datei</th><th class="pct-cell">Coverage</th><th class="stmts">Stmts</th></tr>
  </thead>
  <tbody>
    {{range .Files}}
    <tr>
      <td><code>{{.File}}</code></td>
      <td class="pct-cell">{{pct1 .CoveragePct}}<span class="bar {{.CoverageBucket}}"><span style="width:{{int .CoveragePct}}%"></span></span></td>
      <td class="stmts">{{.Covered}} / {{.Statements}}</td>
    </tr>
    {{end}}
  </tbody>
</table>

</body>
</html>
`
