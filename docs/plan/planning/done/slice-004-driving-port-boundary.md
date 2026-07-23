# Slice 004: Driving-Port-Boundary säubern (Handler ↛ `application`)

**Lifecycle:** Zustand = Verzeichnis. **Welle:** ohne Welle (Architektur-Refactor).

**Bezug:** `spec/architecture.md` §3.4, `slice-003` (a-check-Gate sichert das Ergebnis).

**Autor:** Harness-Migration. **Datum:** 2026-07-23.

---

## 1. Ziel

Die Driving-HTTP-Adapter dürfen die **konkrete `hexagon/application`-Schicht**
nicht mehr importieren — sie sprechen nur Driving-Ports + Domain. Danach fällt die
`adapters→application`-Kante aus `.a-check.yml` (von a-check verifiziert). Behebt
die Boundary-Erosion, die a-check in slice-003 sichtbar gemacht hat.

## 2. Definition of Done

- [x] Kein `hexagon/application`-Import in `adapters/driving/http/**` (nicht-Test).
- [x] `adapters→application`-Kante aus `.a-check.yml` entfernt; `make arch-check`
      (a-check) grün.
- [x] `make test` + `make gates` grün (echter Go-Refactor, Semantik unverändert).
- [x] Closure-Notiz.

## 3. Plan (vor Code) — drei Cluster

**A) Error-Sentinels → `domain`.** `ErrAnalyzeManifestEmpty`,
`ErrSrtHealthStreamUnknown` von `application/*` nach `hexagon/domain/` heben;
`application` gibt `domain.Err…` zurück, der Handler prüft `domain.Err…`.

**B) SSE-Broker → Port + Domain.** `EventAppendedFrame` nach `domain`; ein neues
Driving-Port-Interface (`port/driving`, `Subscribe(ctx, projectID) <-chan
domain.EventAppendedFrame`), das `application.EventBroker` erfüllt (Compile-Check).
`sse_stream.go`/`router.go` hängen am Port + Domain-Typ; `main.go` reicht den
Broker als Port durch.

**C) SRT-Health → `port/driving`.** Die heute **lokal** im HTTP-Adapter definierte
`SrtHealthInbound` samt Result-Typen (`SrtHealthSummary`, `SrtHealthHistoryItem`,
`SrtHealthHistoryPage`) nach `port/driving` heben; `application.SrtHealthQueryService`
implementiert das Port-Interface.

| Datei / Komponente | Änderungs-Art |
|---|---|
| `hexagon/domain/` (Errors + `EventAppendedFrame`) | neu/erweitert |
| `hexagon/port/driving/` (SSE-Subscribe + SrtHealth-Interface + Result-Typen) | neu/erweitert |
| `hexagon/application/{analyze_manifest,srt_health_query_service,event_broker}.go` | update (Typen/Errors umziehen, Ports erfüllen) |
| `adapters/driving/http/{analyze,sse_stream,srt_health_handlers,router}.go` | update (Port/Domain statt `application`) |
| `cmd/api/main.go` | update (Broker als Port durchreichen) |
| `.a-check.yml` | update (`adapters→application`-Kante raus) |

## 4. Trigger

- **`in-progress`:** slice-003 done (a-check-Gate steht). Erfüllt.
- **Rückführung:** falls der Refactor zu groß pro Cluster wird → in A/B/C-Teil-Slices.

## 5. Closure-Trigger

DoD grün + `make gates` grün + Closure-Notiz; `git mv` nach `done/`.

## 6. Risiken und offene Punkte

- **Semantik-Erhalt:** reines Umziehen von Typen/Errors + Port-Extraktion, kein
  Verhaltens-Edit; `make test` ist der Wächter.
- **Channel-im-Port:** der SSE-Subscribe-Port gibt einen `<-chan` zurück (Streaming)
  — bewusst, das ist die Read-Streaming-Schnittstelle.
- **Cursor-Typ:** die SrtHealth-History nutzt `driven.SrtHealthCursor` (Pagination);
  bleibt als querschnittlicher Typ, der Driving-Port darf ihn referenzieren.

## 7. Closure-Notiz

**Ergebnis:** Die Driving-HTTP-Adapter importieren `hexagon/application` nicht mehr
(nicht-Test). Drei Cluster wie geplant: (A) `ErrAnalyzeManifestEmpty` +
`ErrSrtHealthStreamUnknown` nach `domain` gehoben; (B) `EventAppendedFrame` nach
`domain`, neuer Driving-Port `EventStreamInbound` (Compile-Check auf `*EventBroker`);
(C) `SrtHealthInbound` samt Result-Typen (`SrtHealthSummary`, `SrtHealthHistoryItem`,
`SrtHealthHistoryPage`) aus dem HTTP-Adapter nach `port/driving`, erfüllt vom
`SrtHealthQueryService` (Compile-Check). `.a-check.yml`: `adapters→application`-Kante
entfernt — a-check bleibt bei 0 Findings, die Boundary ist jetzt maschinell erzwungen.
`spec/architecture.md` §3.3 um beide neuen Driving-Ports (rollenbasiert) ergänzt.

**Was anders lief (Steering-Loop):** Ein pauschales `gofmt -w ./hexagon ./adapters
./cmd` reformatierte **105 Dateien** — gofmt-1.26.5 formatiert Doc-Kommentar-Listen
(`//  - x` → `//\t- x`) und Konstanten-Alignment anders als der im Repo etablierte
Stand, den golangci-lint hier nicht erzwingt. Das erzeugte massives Diff-Rauschen und
riss `lint-variante-b` auf (`//  §6` → `//\t§6`). Fix: die 86 rein-formatierten
Dateien auf `HEAD` zurückgesetzt, die 7 logisch berührten Nicht-Test-Dateien ebenfalls
zurückgesetzt und die Logik-Edits **chirurgisch** neu appliziert — Ergebnis: 17 Dateien,
nur logische Zeilen. **Lehre:** nie den ganzen Baum formatieren, nur die tatsächlich
editierten Dateien; und Werkzeug-Versionsdrift (`gofmt` vs. golangci-Formatter) vor
einem Blanket-Lauf prüfen. Zweiter Durchgang: revive verlangt Godoc-Form
(`// Name …`, kein `// Name: …`) für die zwei neuen Error-Vars — korrigiert.

**Folge-Slice:** `slice-005` (a-check auf die TS-Pakete + rollenlose Layer-Checks für
`analyzer-service`/`dashboard`/`stream-analyzer`).

## 8. Sub-Area-Modus-Begründung

### Sub-Area: Hexagon-Boundary (apps/api)

- **Modus:** Brownfield (Bestandscode mit erodierter Boundary).
- **Konventionen-Dichte:** hoch — `spec/architecture.md` §3.4 ist die Regel; a-check erzwingt sie.
- **Phase-Reife:** reifer Bestand; die Verdrahtung (DI via Ports) steht schon.
- **Evidenz-/Diskrepanz-Risiko:** niedrig-mittel — Typ-/Error-Umzug ist mechanisch,
  aber berührt die `application`-Public-API; `make test` + a-check verifizieren.
- **Reconciliation-Aufwand:** dieser Slice; danach ist die Boundary Greenfield.
