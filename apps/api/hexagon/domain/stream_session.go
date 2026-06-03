package domain

import (
	"fmt"
	"math"
	"time"
)

// SampleRateFull ist der Integer-ppm-Wert für „Session ist voll
// gesampelt" (entspricht Float `1.0`). Wird als Default in der
// SQLite-Migration V7 (`stream_sessions.sample_rate_ppm`) verwendet
// und ist gleichzeitig der Sentinel-Wert für die
// Immutability-Bedingung im Ingest-Pfad (siehe
//  §6 / R-10).
const SampleRateFull = 1_000_000

// SampleRatePPMFromFloat normalisiert einen vom Player-SDK gelieferten
// `sampleRate`-Float auf den Integer-ppm-Wert für die Persistenz.
// Bereich-Check: `(0, 1]` — Werte außerhalb (≤ 0 oder > 1) liefern
// einen Fehler; Aufrufer mappt das auf einen Drift-Counter und nutzt
// `SampleRateFull` als Fallback (siehe ).
//
// Rundung via `math.Round` (`round-half-away-from-zero`); der zurück-
// gegebene Integer liegt im Bereich `[1, SampleRateFull]`. Float-
// Rundungsartefakte des SDK (z. B. `0.499999…`) werden so deterministisch
// auf den nächstgelegenen ppm-Wert gerundet.
func SampleRatePPMFromFloat(x float64) (int, error) {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, fmt.Errorf("sample_rate %v is not a finite number", x)
	}
	if x <= 0 || x > 1 {
		return 0, fmt.Errorf("sample_rate %v is outside (0, 1]", x)
	}
	ppm := int(math.Round(x * float64(SampleRateFull)))
	// Min-Clamp: sehr kleine Floats (z. B. 1e-7) passieren die
	// `x > 0`-Range-Prüfung, runden aber auf 0; der DB-Range ist
	// `[1, SampleRateFull]`, daher heben wir auf 1 an. Max-Clamp
	// ist nicht nötig: `math.Round(x * 1_000_000)` für `x ≤ 1` ist
	// per Definition `≤ 1_000_000`.
	if ppm < 1 {
		ppm = 1
	}
	return ppm, nil
}

// SessionState ist der grobe Lifecycle einer Player-Session
// .
//
//  - Active: letztes Event innerhalb des Stalled-Schwellwerts.
//  - Stalled: keine Events innerhalb des Schwellwerts (z. B. 60 s),
//  aber noch kein Ended.
//  - Ended: explizites End-Event aus dem SDK oder Inaktivität jenseits
//  des Stalled-Fensters.
//
// Stalled/Ended-Übergänge übernimmt der Lifecycle-Sweeper aus
//  Sub-Item 8 (in 0.1.0 noch ⬜); §5.1 Sub-Item 3
// liefert nur den Zustand „Active" plus die Felder, gegen die der
// Sweeper später entscheidet.
type SessionState string

// Session-Lifecycle-Zustände aus Sub-Item 8.
// `Active` ist der Eintrittszustand beim ersten Event; `Stalled` und
// `Ended` werden vom Sweeper gesetzt (siehe SessionsSweeper).
const (
	SessionStateActive  SessionState = "active"
	SessionStateStalled SessionState = "stalled"
	SessionStateEnded   SessionState = "ended"
)

// StreamSession aggregiert Events mit gleicher session_id
// gesetzt; LastEventAt und EventCount tracken folgende Events derselben
// Session und sind die Grundlage für Lifecycle-Übergänge (Sub-Item 8).
//
// EndedAt wird nur gesetzt, wenn State==Ended; bis dahin nil.
type StreamSession struct {
	ID          string
	ProjectID   string
	State       SessionState
	StartedAt   time.Time
	LastEventAt time.Time
	EndedAt     *time.Time
	EventCount  int64
	// CorrelationID ist die Server-generierte, durable Source-of-Truth
	// für die Tempo-unabhängige Dashboard-Korrelation der Session. Wird
	// beim allerersten Event der Session erzeugt (UUIDv4) und über alle
	// Folge-Events konstant gehalten. Source spec/telemetry-model.md
	CorrelationID string
	// EndSource benennt den Auslöser des Endzustands:
	//  - SessionEndSourceClient bei explizitem `session_ended`-Event
	//  - SessionEndSourceSweeper bei zeitbasiertem Sweeper-Ende
	//  - "" (Leerwert) wenn State != ended, oder bei Legacy-Sessions
	//  vor dem V4-Migration-Closeout
	// Read-Pfad mappt den Leerwert auf JSON `null` (siehe API-Kontrakt
	// .
	EndSource SessionEndSource
	// SampleRatePPM ist die normalisierte Sampling-Rate der Session in
	// Integer-ppm (parts per million). `SampleRateFull` = voll gesampelt
	// (Default seit Migration V7). Immutable nach erstem Sub-`SampleRateFull`-
	// Wert; spätere Drift wird in `mtrace_sample_rate_drift_total`
	// gezählt, überschreibt aber nicht. Siehe / R-10
	// und spec/telemetry-model.md
	SampleRatePPM int
}

// SessionEndSource klassifiziert den Auslöser des Endzustands einer
// Session ( H1).
type SessionEndSource string

// Session-EndSource-Werte; siehe spec/backend-api-contract.md
const (
	// SessionEndSourceClient: explizites `session_ended`-Event vom SDK.
	SessionEndSourceClient SessionEndSource = "client"
	// SessionEndSourceSweeper: zeitbasiertes Ende durch SessionsSweeper.
	SessionEndSourceSweeper SessionEndSource = "sweeper"
)
