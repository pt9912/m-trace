# R-13 Trivy Re-Review - 2026-05-13

## Kontext

`R-13` betrifft drei bekannte OS-CVEs in der gemeinsamen
`node:22-trixie-slim`-Basis von Dashboard und Analyzer-Service:

- `CVE-2025-69720`
- `CVE-2026-29111`
- `CVE-2026-4878`

Die Single Source of Truth fuer die Suppressions bleibt
[`../../../.security/vulnignore.yaml`](../../../.security/vulnignore.yaml).

## Eingaben

| Feld | Wert |
| --- | --- |
| Re-Review-Zeitpunkt | `2026-05-13T11:03:33+02:00` |
| Trivy-Image | `aquasec/trivy:0.59.1` |
| Scan-Target API | `mtrace-api:scan` |
| Scan-Target Dashboard | `mtrace-dashboard:scan` |
| Scan-Target Analyzer-Service | `mtrace-analyzer-service:scan` |
| Ignore-Quelle | `.security/vulnignore.yaml` |
| Gerenderte Ignore-Datei | `.security/.trivyignore` |
| Aktuelles `expires` | `2026-11-02` |

## Kommandos

```sh
make image-scan
```

Der Lauf baut alle drei Runtime-Images, rendert `.security/.trivyignore`
pro Image-Scope aus `.security/vulnignore.yaml` und scannt mit:

```sh
trivy image --severity CRITICAL,HIGH --exit-code 1 --no-progress --ignorefile /work/.trivyignore <image>
```

## Ergebnis

| Image | Rendered Ignores | Ergebnis |
| --- | ---: | --- |
| `mtrace-api:scan` | 0 | `Total: 0 (HIGH: 0, CRITICAL: 0)` |
| `mtrace-dashboard:scan` | 3 | `Total: 0 (HIGH: 0, CRITICAL: 0)` |
| `mtrace-analyzer-service:scan` | 3 | `Total: 0 (HIGH: 0, CRITICAL: 0)` |

Dashboard und Analyzer-Service melden weiterhin unterdrueckte bekannte
Findings; unignorierte HIGH-/CRITICAL-Findings sind nicht vorhanden.

## Upstream-/Distro-Stand

Debian Security Tracker wurde am 2026-05-13 gegengeprueft:

| CVE | Debian-Stand fuer `trixie` | Fix-Stand |
| --- | --- | --- |
| [`CVE-2025-69720`](https://security-tracker.debian.org/tracker/CVE-2025-69720) | `vulnerable`, `no-dsa` fuer `ncurses` | `forky`/`sid` fixed |
| [`CVE-2026-29111`](https://security-tracker.debian.org/tracker/CVE-2026-29111) | `vulnerable`, `no-dsa` fuer `systemd` | `forky`/`sid` fixed; `bullseye-security` fixed |
| [`CVE-2026-4878`](https://security-tracker.debian.org/tracker/CVE-2026-4878) | `vulnerable`, `no-dsa` fuer `libcap2` | `forky`/`sid` fixed |

## Entscheidung

`R-13` bleibt `continued`/offen bis zur Triggerschwelle:

- `expires` ist nicht erreicht (`2026-11-02`).
- Der aktuelle Scan liefert keine unignorierten HIGH-/CRITICAL-Findings.
- Der Debian-Stand fuer `trixie` zeigt weiterhin keinen eingespielten
  Fix fuer die betroffene Trixie-Slim-Basis.
- Distroless bleibt ein expliziter Folgeentscheid, nicht Teil von
  `0.18.0` Tranche 1.
