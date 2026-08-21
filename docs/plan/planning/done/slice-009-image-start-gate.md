# Slice 009: Runtime-Images im Gate starten (nicht nur bauen und scannen)

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle (Werkzeug-Fix).

**Bezug:** `373db24` (Befund beim Verifizieren des `pnpm deploy`-Fixes),
`apps/dashboard/Dockerfile`, `apps/analyzer-service/Dockerfile`,
`Makefile::gates` / `::security-gates`.

**Autor:** Nightly-Audit-Nachgang. **Datum:** 2026-08-21.

---

## 1. Ziel

Der Gate-Pfad stellt sicher, dass ein Runtime-Image **startet**. Heute tut das
kein Gate: `image-scan` baut die drei Images und scannt sie statisch, `gates`
fasst sie gar nicht an. Ein Image, das beim ersten `docker run` sofort stirbt,
passiert damit jede Prüfung grün.

Das ist keine Theorie. Genau so blieb unbemerkt, dass das
analyzer-service-Image nicht lauffähig war:

```
Error: Cannot find module '@pt9912/stream-analyzer'
```

Ursache war ein `pnpm deploy --legacy`-Bundle, dessen Workspace-Symlink aus dem
Bundle heraus zeigte (behoben in `373db24`). Aufgefallen ist es nur, weil beim
Verifizieren einer **anderen** Änderung zufällig ein Container gestartet wurde —
nicht durch eine Prüfung. Der Defekt hätte beliebig lange überdauert.

Verschärfend: `mtrace-dashboard` trug denselben Defekt und **überlebte**, weil
adapter-node sein Workspace-Paket zur Build-Zeit einbundelt. Ein Gate, das nur
ein Image stichprobenartig prüft, hätte hier also das falsche erwischt.

## 2. Definition of Done

1. Ein Target (Arbeitstitel `make image-start-check`) startet **jedes** der drei
   Runtime-Images und stellt fest, dass der Prozess oben bleibt.
2. Der Check ist an den Gate-Pfad gehängt — Ort ist Teil des Plans (§3, offen
   zwischen `gates` und `security-gates`).
3. Der Check schlägt gegen ein bewusst kaputtes Image nachweislich fehl
   (Negativ-Probe, nicht nur Grün-Lauf).
4. Er braucht **keinen** Compose-Stack: Laufzeit im Sekundenbereich, keine DB,
   kein MediaMTX, kein Netz zu anderen Services.
5. `make gates` bleibt grün.

## 3. Plan (vor Code)

**A) Liveness-Kriterium.** Kein Health-Endpoint als Pflicht — der API-Container
ist distroless und hat keine Shell, das Dashboard will einen Port, der Analyzer
bringt `/health` mit. Tragfähig für alle drei ist: Container starten, kurze
Karenz, prüfen dass er **noch läuft** (kein Exit). Genau das trennt den
Defekt-Fall (Exit 1 nach <1 s) vom gesunden Fall. Wo ein Health-Endpunkt
existiert, kann er additiv geprüft werden.

**B) Einhängeort.** `gates` baut heute keine Images, `security-gates` schon.
Kandidaten: (a) an `security-gates` hängen (Images sind dort ohnehin gebaut,
kein zweiter Build), (b) eigenes Target mit eigenem Build (teurer, aber
unabhängig von der Security-Linie). Entscheidung im Slice.

**C) Negativ-Probe.** Einmalig gegen ein absichtlich kaputtes Image
verifizieren (z. B. `--legacy`-Stand aus `373db24^`), damit der Check nicht
wie das Closure-Gate in `slice-005` grün läuft und nichts misst.

## 4. Trigger

- **`in-progress`:** jederzeit; der Befund steht, `373db24` ist gefixt.
- **Rückführung:** reines Werkzeug + Makefile; kein Produktcode berührt.

## 5. Closure-Trigger

DoD grün + Negativ-Probe belegt + `make gates` grün + Closure-Notiz;
`git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Startdauer vs. Aussagekraft.** Zu kurze Karenz übersieht späte Abstürze, zu
  lange bremst den Gate-Lauf. Der hier gejagte Fehlerfall stirbt in
  Millisekunden — eine kleine Karenz genügt und bleibt ehrlich, solange der
  Check nicht als „Image ist funktionsfähig" verkauft wird. Er beantwortet nur:
  *startet es überhaupt.*
- **Kein Ersatz für `smoke-analyzer`.** Der Compose-Smoke prüft fachliche
  Integration; dieser Check prüft Startfähigkeit. Beide bleiben.
- **Distroless-API.** Ohne Shell ist kein `docker exec`-Probing möglich; das
  Liveness-Kriterium aus (A) kommt bewusst ohne aus.

## 7. Closure-Notiz

Der Check ist `scripts/image-start-check.sh` + `make image-start-check`, gehängt
an `security-gates`. Drei Entscheidungen aus §3 sind so gefallen:

**(A) Liveness statt Health.** Das Kriterium ist „Container startet und **bleibt**
nach einer Karenz oben", nicht ein Health-Endpunkt. Grund ist das API-Image:
distroless, keine Shell, kein `docker exec`-Probing. Der gejagte Fehlerfall
stirbt ohnehin in Millisekunden, die Karenz (`GRACE`, Default 5 s) trennt ihn
sauber vom gesunden Fall. Der Check behauptet ausdrücklich **nicht**, dass ein
Image fachlich funktioniert — dafür bleiben die Compose-Smokes zuständig.

**(B) Einhängeort `security-gates`, nicht `gates`.** Dort werden die `:scan`-Tags
ohnehin gebaut, der Check braucht also keinen eigenen Build.

**Nachtrag (Korrektur am selben Tag).** Der erste Entwurf hängte das Target per
`image-start-check: image-scan` an den Scan — begründet damit, dass make ein
PHONY-Target pro Lauf nur einmal ausführt. Das stimmt lokal, ging an der
CI-Realität aber vorbei: `.github/workflows/build.yml` ruft die Security-Targets
**einzeln** auf (`make vuln-check`, `make audit-ts`, `make image-scan`), nie
`make security-gates`. Zwei Folgen, beide schlecht — der Check wäre in CI
überhaupt nicht gelaufen (kein Schritt rief ihn auf), und hätte man ihn naiv
ergänzt, hätte die Dependency dort in einem eigenen make-Lauf einen kompletten
zweiten Image-Build ausgelöst. Korrigiert: die Dependency ist entfernt, die
Reihenfolge sichert `security-gates`, ein fehlendes Image meldet das Script
klar, und `build.yml` hat einen eigenen Schritt nach dem Trivy-Scan, der die
dort gebauten `:scan`-Tags weiterverwendet.

Das war fast dieselbe Falle wie die, die dieser Slice schließt: ein Gate, das
lokal grün läuft und im entscheidenden Pfad gar nicht stattfindet. Aufgefallen
ist es nur, weil nach dem grünen CI-Lauf nachgeprüft wurde, **ob** der neue
Schritt in CI tatsächlich ausgeführt wurde — die Job-Ampel „Security gates:
success" bezog sich auf die drei alten Schritte.

**Nicht in den Nightly gehängt** (`security-audit.yml`), bewusst: der ist
strukturell auf genau drei Gates zugeschnitten (`id`/`outcome` je Schritt, das
Report-Script berichtet über drei ENV-Variablen, Fail-Bedingung listet drei
Outcomes). Und er jagt zeitgetrieben **neue Advisories**, während ein
Start-Check Regressionen fängt — die deckt `build.yml` bei jedem Push ab.

**(C) Negativ-Probe gegen den echten Vor-Fix-Stand.** Nicht gegen ein
synthetisch kaputtes Image: `pnpm-workspace.yaml`, `pnpm-lock.yaml` und
`apps/analyzer-service/Dockerfile` aus `373db24^` ausgecheckt, Image als
`:brokenprobe` gebaut, Arbeitsbaum sofort zurückgestellt. Ergebnis:
`FAIL … beendet innerhalb 5s (Status exited, Exit 1)` samt
`code: 'MODULE_NOT_FOUND'` in den mitgelieferten Container-Logs. Damit ist
belegt, dass der Check genau den Defekt gefangen hätte, der ihn ausgelöst hat.

**Lehre, direkt aus `slice-005` geerbt:** ein neues Gate ist erst dann etwas
wert, wenn es nachweislich auch **rot** werden kann. Der Positiv-Lauf allein
hätte hier nichts belegt — `image-scan` lief schließlich jahrelang grün, während
ein totes Image daneben stand. Deshalb ist die Negativ-Probe Teil der DoD (Punkt
3) und nicht bloß eine Fußnote.

Verifikation: `make security-gates` grün mit dem Check am Ende (`OK` für alle
drei Images), `make gates` grün, Negativ-Probe rot wie erwartet.

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Image-Start-Gate (Werkzeug)

Reines Werkzeug plus Makefile-Verdrahtung; kein Produktcode, keine Spec, kein
Requirement berührt. Damit ohne Welle (Modul 5 „Wartung/Architektur") und ohne
ADR — es wird kein Gate gesenkt, sondern eines ergänzt.
