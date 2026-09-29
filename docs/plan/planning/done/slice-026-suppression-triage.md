# Slice 026: Suppression-Triage — abgelaufene `expires` in vulnignore.yaml

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle — Closure-Bedingung
ist die DoD dieses Slices, keine Wellen-Grenze.

**Bezug:** `.security/vulnignore.yaml`, `make image-scan`, Nightly-Audit-
Issues #36–#48 (GitHub), PR #38 (dependabot, extern).

**Autor:** Nightly-Befund 2026-09-29 (Issue #48) + Owner-Antrag. **Datum:** 2026-09-29.

---

## 1. Ziel und Abgrenzung

**Ziel:** Den `image-scan`-Gate wieder grün bekommen: die fünf abgelaufenen
Suppressions triagieren — die drei util-linux-Einträge (CVE-2026-53612/13/14)
ersatzlos entfernen, denn das Base `node:22-trixie-slim` trägt jetzt
`2.41.5-0+deb13u1` (dpkg-geprüft) und damit den dokumentierten
Re-Review-Trigger; die beiden acl/attr-Einträge (CVE-2026-54369/54371)
re-reviewen und verlängern (Base unverändert: `libacl1 2.3.2-2+b1`,
`libattr1 1:2.5.2-3`, kein Fix). Ergänzung im Lauf: **CVE-2026-41992
(gzip) läuft 2026-09-30 ab** — dieselbe Re-Review-Lage (`gzip
1.13-1+deb13u1` im Base, Fix >1.14 weiterhin nicht paketiert), wird mit
verlängert, sonst bricht der Gate morgen erneut.

**Ausdrücklich NICHT in diesem Slice:**

- **CVE-2026-53615** (util-linux parse_dos_extended, `expires` 2026-10-08) —
  eigener Re-Review-Trigger mit eigener Schwelle, 9 Tage entfernt; Fix-Status
  von deb13u1 ist ungeklärt (Changelog im slim-Image nicht verfügbar). Bleibt
  stehen, Triage an der Schwelle.
- **Die perl-Kohorte** (`expires` 2026-11-02, R-13) — gehört zur Produkt-Folge-
  well, nicht in die Suppression-Triage.
- **Schließen der 12 Nightly-Issues** gehört zum DoD dieses Slices (nach grünem
  Gate); die Nightly-Infrastruktur selbst (Issue-Erzeugung pro Fehllauf) bleibt
  unverändert.

## 2. Definition of Done

- [x] CVE-2026-53612/53613/53614 ersatzlos aus `.security/vulnignore.yaml`
      entfernt (Re-Review-Trigger eingetreten: Base trägt `2.41.5-0+deb13u1`).
- [x] CVE-2026-54369/54371 re-reviewt: frischer Re-Review-Vermerk mit Datum,
      `expires` auf 2026-10-29 verlängert.
- [x] CVE-2026-41992 (gzip) re-reviewt und auf 2026-10-29 verlängert
      (lief 2026-09-30 ab).
- [x] `make image-scan` grün (lokal ausgeführt, Output in Datei —
      render-trivyignore ohne Abbruch: api 0, dashboard/analyzer je 22
      gültige Einträge; Trivy exit 0).
- [x] Die 12 offenen Nightly-Issues (#36–#48) geschlossen mit Verweis auf den
      Fix-Commit (`f42426b`).
- [x] `make docs-check` grün.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.

## 3. Plan (vor Code)

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `.security/vulnignore.yaml` | update | 3 Einträge entfernt (Fix im Base), 2 verlängert (Re-Review) |

## 4. Trigger

- **Start (`next` → `in-progress`):** sofort — Owner-Antrag, die Gate-Lücke
  ist der Grund.
- **Rückführungen:** keine erwartet.

## 5. Closure-Trigger

DoD vollständig + `make image-scan` grün + Closure-Notiz + `git mv` nach
`done/`.

## 6. Risiken und offene Punkte

- **CVE-2026-53615 könnte ebenfalls durch deb13u1 gefixt sein** — ohne
  Changelog-Zugriff im slim-Base ist das nicht geprüft; die Suppression
  bleibt bis zur eigenen Schwelle (2026-10-08) stehen und maskiert den
  Befund. **Ausgang:** weiter offen → Triage an der `expires`-Schwelle.
- **12 Issues bulkgeschlossen** — wenn das nächste Nightly erneut rot läuft,
  öffnet es ein neues Issue; die Schließung verliert nichts.

## 7. Closure-Notiz

Triage vollzogen: util-linux-Trio (CVE-2026-53612/13/14) ersatzlos entfernt
— der in jedem Eintrag dokumentierte Re-Review-Trigger („Base zieht
`2.41.5-0+deb13u1` ein") ist eingetreten, dpkg-geprüft im frisch gepullten
Base. acl/attr (CVE-2026-54369/54371) und gzip (CVE-2026-41992) re-reviewt:
kein Fix, Vektor-Verschluss unverändert (runtime ruft die Pfade nie auf),
Re-Review auf 2026-10-29 verlängert. `make image-scan` lokal grün; die 12
Nightly-Issues (#36–#48) geschlossen.

**Was hat funktioniert:** Die Re-Review-Trigger standen in den Einträgen
selbst — der dpkg-Check im Base hat die Entfernungs-Entscheidung auf einen
Befund reduziert (Version-Zeile gegen die `reason`-Bedingung). Die
vollständige `expires`-Landschaft (`grep expires` + Sortierung) statt der
Issue-Diagnose zu arbeiten, hob den gzip-Fund (Ablauf einen Tag später) und
verhinderte den nächsten Gate-Bruch.

**Was ging anders als geplant:** Zwei Punkte. (1) gzip (CVE-2026-41992)
war im Plan nicht vorgesehen — der Generator meldet nur die *abgelaufenen*
Einträge, die kurz davor liegenden bleiben unsichtbar; gefunden durch die
Landschaftsprüfung. (2) Der Changelog-Check zu CVE-2026-53615 (ist deb13u1
auch der parse_dos_extended-Fix?) war ohne Paketchangelog im slim-Base
nicht führbar — der Eintrag bleibt bis zur eigenen Schwelle (2026-10-08)
stehen.

**Steering-Loop-Eintrag:** Guide geschärft: bei jeder Suppression-Triage
zuerst die vollständige `expires`-Liste sortiert ziehen und gegen *heute*
prüfen — der Gate-Generator meldet nur Ablaufs, und eine Triage, die nur
die gemeldeten Fälle bearbeitet, garantiert den nächsten Bruch. (Gezählt,
nicht verkörpert — Erstvorkommen.)

**Beobachtungs-Register (`../observations/`):** keine Beobachtung angefallen.
Review-Report slice-025 (F-1/F-2/F-3 — LOW/LOW/INFO) ist Erstvorkommen je
Klasse, kein Register-Eintrag (3×-Schwelle nicht erreicht).

**Folge-Slices:** keine. Nächste Suppression-Schwellen: 2026-10-08
(CVE-2026-53615), 2026-10-29 (acl/attr/gzip), 2026-11-02 (perl-Kohorte +
R-13).

**Risiken aus §6:** beide mit Ausgang — CVE-2026-53615 weiter offen
(Schwelle 2026-10-08, siehe oben); Bulk-Close vollzogen, Verlustszenario
durch die Issue-Automatik des Nightly abgedeckt.

## 8. Sub-Area-Prüfungen und Modus-Begründung

Reine Suppression-Triage nach der Wartungsregel des Scanners (fail-closed
Generator). Berührte Sub-Area: Security-Gate-Suppressions (`SECGATE`,
Brownfield, observable) — der in `harness/conventions.md` deklarierte
Graduierungspfad trifft hier nicht zu (trixie-slim-Base trägt weiterhin
transitive OS-CVEs), deshalb Re-Review statt Auflösung. Kein ADR — kein Gate
gesenkt: die Entfernung verschärft, die Verlängerung folgt dem
Re-Review-Muster der Einträge.
