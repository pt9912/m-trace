# perl-base-Cohort Re-Review - 2026-08-16

## Kontext

Das perl-base-Cohort umfasst neun Perl-Vulnerabilities in der
gemeinsamen `node:22-trixie-slim`-Basis von Dashboard und
Analyzer-Service (`perl-base` 5.40.1-6):

- `CVE-2026-42496`, `CVE-2026-42497` (Archive::Tar Symlink/Hardlink)
- `CVE-2026-9538` (Archive::Tar Memory-Exhaustion)
- `CVE-2026-8376` (Regex-Compile Heap Buffer Overflow)
- `CVE-2026-48962`, `CVE-2026-48959` (IO::Compress / IO::Uncompress)
- `CVE-2026-13221` (silently incorrect Regex)
- `CVE-2026-57432` (S_measure_struct Integer Overflow)
- `CVE-2026-57433` (Storable signed Integer Overflow)

Anlass ist der bevorstehende `expires`-Ablauf am `2026-08-26`. Das
Cohort teilt Images, Mitigation und strukturellen Ausweg mit
[`R-13`](../plan/planning/risks-backlog.md#r-13), ist dort aber bis zu
diesem Re-Review nicht mitgeführt worden.

Die Single Source of Truth fuer die Suppressions bleibt
[`.security/vulnignore.yaml`](../../.security/vulnignore.yaml).

## Eingaben

| Feld | Wert |
| --- | --- |
| Re-Review-Zeitpunkt | `2026-08-16` |
| Trivy-Image | `aquasec/trivy:0.74.0` |
| Scan-Target API | `mtrace-api:scan` |
| Scan-Target Dashboard | `mtrace-dashboard:scan` |
| Scan-Target Analyzer-Service | `mtrace-analyzer-service:scan` |
| Ignore-Quelle | `.security/vulnignore.yaml` |
| Gerenderte Ignore-Datei | `.security/.trivyignore` |
| `expires` vorher | `2026-08-26` |
| `expires` nachher | `2026-11-02` |

## Kommandos

```sh
make image-scan
bash scripts/render-trivyignore.sh <image-scope>
```

## Ergebnis

| Image | Rendered Ignores | Ergebnis |
| --- | ---: | --- |
| `mtrace-api:scan` | 0 | keine CRITICAL/HIGH-Findings, Exit 0 |
| `mtrace-dashboard:scan` | 18 | keine unignorierten CRITICAL/HIGH-Findings, Exit 0 |
| `mtrace-analyzer-service:scan` | 18 | keine unignorierten CRITICAL/HIGH-Findings, Exit 0 |

Die API bleibt unbetroffen: ihr Runtime steht auf
`gcr.io/distroless/static-debian12:nonroot` und enthaelt kein
`perl-base`.

## Upstream-/Distro-Stand

Primaerquelle war diesmal der Debian Security Tracker
(`source-package/perl`, Spalte `trixie`), nicht die leere
Trivy-`Fixed Version`-Spalte: letztere belegt nur, dass Trivy keinen
Fix kennt, nicht warum. Stand `2026-08-16`:

| CVE | `trixie` | `forky` | `sid` |
| --- | --- | --- | --- |
| [`CVE-2026-42496`](https://security-tracker.debian.org/tracker/CVE-2026-42496) | `vulnerable (no DSA, postponed)` | `vulnerable` | `vulnerable` |
| [`CVE-2026-42497`](https://security-tracker.debian.org/tracker/CVE-2026-42497) | `vulnerable (no DSA, postponed)` | `vulnerable` | `vulnerable` |
| [`CVE-2026-9538`](https://security-tracker.debian.org/tracker/CVE-2026-9538) | `vulnerable (no DSA, postponed)` | `vulnerable` | `vulnerable` |
| [`CVE-2026-13221`](https://security-tracker.debian.org/tracker/CVE-2026-13221) | `vulnerable` | `vulnerable` | `vulnerable` |
| [`CVE-2026-8376`](https://security-tracker.debian.org/tracker/CVE-2026-8376) | `vulnerable (no DSA)` | `fixed` | `fixed` |
| [`CVE-2026-48962`](https://security-tracker.debian.org/tracker/CVE-2026-48962) | `vulnerable` | `fixed` | `fixed` |
| [`CVE-2026-48959`](https://security-tracker.debian.org/tracker/CVE-2026-48959) | `vulnerable` | `fixed` | `fixed` |
| [`CVE-2026-57432`](https://security-tracker.debian.org/tracker/CVE-2026-57432) | `vulnerable` | `fixed` | `fixed` |
| [`CVE-2026-57433`](https://security-tracker.debian.org/tracker/CVE-2026-57433) | `vulnerable` | `fixed` | `fixed` |

**Null von neun** sind in `trixie` gefixt. Zwei Lager:

- **Backport ausstehend** (in `forky`/`sid` gefixt): `CVE-2026-8376`,
  `CVE-2026-48962`, `CVE-2026-48959`, `CVE-2026-57432`,
  `CVE-2026-57433`. Der bisher notierte Backport-Trigger bleibt hier
  sinnvoll.
- **Nirgends gefixt**, auch nicht in `sid`: `CVE-2026-42496`,
  `CVE-2026-42497`, `CVE-2026-9538`, `CVE-2026-13221`. Fuer diese vier
  ist der Upstream-Fix-Trigger noch gar nicht erreichbar; drei davon
  fuehrt der Tracker als `no DSA, postponed` — das Security-Team plant
  dort aktiv keinen Backport.

Ein Base-Bump oder `apt-get upgrade` macht den Gate damit weiterhin
nicht gruen.

## Entscheidung

Verlaengerung statt Aufloesung; der Runtime-Vektor ist unveraendert
nicht erreichbar (Startpfad ist node-only, kein `perl`-Aufruf im
Runtime, keine tar-/ZIP-Extraktion und keine Regex-Kompilierung durch
`perl`).

- `expires` auf `2026-11-02` gesetzt — bewusst kuerzer als die
  90-Tage-Default-Frist (waere `2026-11-14`), um mit dem
  [`R-13`](../plan/planning/risks-backlog.md#r-13)-Cluster
  (ncurses/systemd/libcap) gleichzulaufen. Beide beschreiben dieselbe
  Lage in denselben zwei Images und haben denselben strukturellen
  Ausweg; ein gemeinsamer Re-Review-Termin statt zweier dicht
  aufeinanderfolgender. Die Verkuerzung ist die konservative Richtung.
- Das Cohort wird ab jetzt unter
  [`R-13`](../plan/planning/risks-backlog.md#r-13) mitgefuehrt statt
  unter einer eigenen Kennung: identischer Aufloesungs-Trigger, eine
  zweite Kennung waere ein Duplikat.
- Struktureller Ausweg bleibt `R-13` Folge-Trigger (c): ein
  perl-freies Runtime-Base (`gcr.io/distroless/nodejs22-debian12`)
  eliminiert dieses Cohort und den `R-13`-Cluster gemeinsam.

Nicht gangbar ist das Loeschen der perl-Dateien analog zum
npm-Rauswurf im Dashboard-Runtime: Trivy erkennt `perl-base` ueber die
dpkg-Datenbank, nicht ueber die Dateien. Das Finding bliebe bestehen,
und den dpkg-Eintrag zu entfernen waere keine Behebung, sondern eine
getarnte Suppression.
