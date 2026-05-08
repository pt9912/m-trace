# Security Policy

m-trace ist ein Observability-/Tracing-Werkzeug für lokale
Streaming-Labs (siehe [`README.md`](README.md) Abschnitt
„Was m-trace nicht ist"). Es ist nicht für Multi-Tenant-Produktion
oder öffentlich exponierte Deployments ausgelegt. Diese Policy
beschreibt trotzdem das Vorgehen für Sicherheitsmeldungen.

## 1. Unterstützte Versionen

| Version-Reihe | Status |
| ------------- | ------ |
| `0.9.x` | Aktiv unterstützt; Security-Fixes landen im jeweils nächsten Patch-Release. |
| `0.8.x` und älter | Nicht mehr aktiv unterstützt; bitte auf den aktuellen `0.9.x`-Stand aktualisieren. |

Die jeweils aktuelle Release-Reihe ergibt sich aus dem letzten
Tag (`git tag --list 'v*'`) und der Release-Übersicht in
[`docs/planning/in-progress/roadmap.md`](docs/planning/in-progress/roadmap.md)
§3.

## 2. Sicherheitslücke melden

- Sicherheitslücken **nicht** in öffentlichen GitHub-Issues, PRs,
  Diskussionen oder Commit-Nachrichten posten.
- Stattdessen privat an die Repository-Maintainer melden — etwa
  über GitHub Security Advisories („Report a vulnerability") oder
  per E-Mail an die im Repository-Profil hinterlegte
  Maintainer-Adresse.
- Eine Meldung soll enthalten: betroffene Version (Tag oder
  Commit-SHA), Kurzbeschreibung der Schwachstelle, Reproduktions-
  Pfad (idealerweise ein Compose-/`make`-Target), erwartete
  Auswirkung.
- Keine Exploit-Payloads oder Secrets in öffentliche Channels
  posten — auch nicht in Issue-Kommentaren.

Wir bestätigen den Eingang innerhalb von 7 Tagen und nennen einen
Zeitplan für Triage und Fix.

## 3. Scope

In Scope:

- Schwachstellen in produktivem Code unter `apps/`, `packages/`,
  `services/`, `scripts/` oder in CI-/Release-Prozessen, die
  Nutzer:innen oder Operator:innen unmittelbar betreffen.
- SSRF-/Injection-Lücken in API-, Analyzer- oder Dashboard-Code.
- Veröffentlichte Container-Images oder Pakete (sobald sie
  veröffentlicht sind — aktuell deferred laut
  [`docs/user/releasing.md`](docs/user/releasing.md)).

Außerhalb des Scope (Bug-Report, kein Security-Issue):

- Probleme nur in den Compose-Lab-Defaults (z. B. `demo-token`,
  Lab-CORS), wenn das Setup so dokumentiert ist.
- Findings in Drittanbieter-Images (MediaMTX, FFmpeg, Grafana,
  Prometheus, Tempo, Playwright). Diese werden über
  Upstream-Projekte gefixt; ein PR auf das Image-Pinning ist
  willkommen.

## 4. Hardening und Security-Gates

m-trace hat ab `0.8.5` (Quality-Gates Wave 1) PR-blockierende
Security-Gates: `vuln-check` (Go), `audit-ts` (npm/pnpm),
`image-scan` (gebaute Images) und ein Generated-Artifact-
Drift-Gate. Sie laufen parallel zu `make build` und sind in
[`docs/user/quality.md`](docs/user/quality.md) und im
Lieferstand-Plan
[`docs/planning/done/plan-0.8.5.md`](docs/planning/done/plan-0.8.5.md)
beschrieben. Eine sicherheitsrelevante Änderung muss diese Gates
weiterhin grün halten.

## 5. Disclosure

Nach einem Fix:

- Patch-Release entlang der Patch-Release-Konvention in
  [`docs/user/releasing.md`](docs/user/releasing.md) §3.1.
- Eintrag im [`CHANGELOG.md`](CHANGELOG.md) unter `Security` der
  jeweiligen Versions-Sektion mit Verweis auf das Advisory bzw.
  CVE, sobald vorhanden.
- Coordinated Disclosure auf Wunsch des Melders.
