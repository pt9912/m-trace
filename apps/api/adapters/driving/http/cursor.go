package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// Cursor-Fehlerklassen aus ADR-0004 §6 / API-Kontrakt §10.3. Der
// HTTP-Adapter mappt sie auf die in §10.3 definierten Bodies und
// HTTP-Status. Domain- und Application-Layer kennen diese Klassen
// nicht — sie sind Wire-Format-Detail.
var (
	// errCursorInvalidLegacy: Token decodiert; `v`-Feld fehlt, ist
	// 1 oder 2 (Pre-§4.3-Format ohne Project-/Session-Scope) — oder
	// das gemeinsame Erkennungsmerkmal aus 0.1.x ist da. Dauerhaft
	// abgewiesen — kein One-Shot-Grace-Pfad.
	errCursorInvalidLegacy = errors.New("cursor_invalid_legacy")
	// errCursorInvalidMalformed: Base64-/JSON-Decode schlägt fehl;
	// `v`-Feld unbekannt; Pflichtfeld fehlt oder hat ungültiges
	// Format; v3-Cursor mit fremdem Project- oder Session-Scope
	// (siehe ADR-0004 §6 / API-Kontrakt §10.3); oder unbekannte
	// Zusatzfelder vorhanden.
	errCursorInvalidMalformed = errors.New("cursor_invalid_malformed")
	// errCursorExpired: Token decodiert valide, aber Storage-Position
	// existiert nicht mehr (Retention/Wipe). In 0.4.0 ohne TTL nur via
	// `make wipe` erreichbar — Code-Pfad ist dennoch vorgesehen, damit
	// Retention-Folge-Arbeit ohne Wire-Format-Bruch möglich bleibt.
	errCursorExpired = errors.New("cursor_expired")
)

// cursorVersion ist die einzige unterstützte Cursor-Version ab
// plan-0.4.0 §4.3. Tokens mit `v=1` oder `v=2` werden als Legacy
// abgewiesen; v3 trägt zusätzlich `pid` (Project-Scope) und für
// Event-Cursor `sid` (Collection-Scope), damit ein Cursor aus
// Project A nicht im Request-Kontext von Project B (oder ein Event-
// Cursor aus Session X nicht im Request-Kontext von Session Y) als
// gültig akzeptiert wird (`cursor_invalid_malformed`).
const cursorVersion = 3

// wireListSessionsCursor ist die JSON-Form von
// driving.ListSessionsCursor. Felder bewusst kurz, damit der Cursor
// im URL-Query klein bleibt. `V` und `PID` sind ab v3 Pflicht.
type wireListSessionsCursor struct {
	V   int    `json:"v"`
	PID string `json:"pid"` // project_id (Project-Scope ab v3)
	SA  string `json:"sa"`  // started_at, RFC3339Nano
	SID string `json:"sid"` // session_id
}

// wireSessionEventsCursor ist die JSON-Form von
// driving.SessionEventsCursor. `V`, `PID` und `SID` sind ab v3
// Pflicht und tragen den Collection-Scope `(project_id, session_id)`.
type wireSessionEventsCursor struct {
	V   int    `json:"v"`
	PID string `json:"pid"` // project_id (Collection-Scope ab v3)
	SID string `json:"sid"` // session_id (Collection-Scope ab v3)
	RCV string `json:"rcv"` // server_received_at, RFC3339Nano
	SEQ *int64 `json:"seq,omitempty"`
	ING int64  `json:"ing"` // ingest_sequence
}

// listSessionsRawCursor ist eine Decode-Zwischenform mit allen
// jemals belegten Feldnamen. `pid` hat in 0.1.x die Bedeutung
// `process_instance_id` (Legacy) und ab v3 die Bedeutung `project_id`
// (Project-Scope) — daher ist die Versions-Erkennung über `v` Pflicht
// vor jeder Interpretation von `pid`.
type listSessionsRawCursor struct {
	V   *int    `json:"v,omitempty"`
	PID *string `json:"pid,omitempty"`
	SA  *string `json:"sa,omitempty"`
	SID *string `json:"sid,omitempty"`
}

type sessionEventsRawCursor struct {
	V   *int    `json:"v,omitempty"`
	PID *string `json:"pid,omitempty"`
	SID *string `json:"sid,omitempty"`
	RCV *string `json:"rcv,omitempty"`
	SEQ *int64  `json:"seq,omitempty"`
	ING *int64  `json:"ing,omitempty"`
}

// encodeListSessionsCursor liefert den base64-url-encodierten Cursor
// für die Sessions-Listenpage (cursor_version 3, mit Project-Scope).
// `projectID` muss der aufgelöste Project-Kontext des Requests sein,
// damit ein zurückgegebener `next_cursor` nur im selben Project
// wieder akzeptiert wird.
func encodeListSessionsCursor(c *driving.ListSessionsCursor, projectID string) (string, error) {
	if c == nil {
		return "", nil
	}
	wire := wireListSessionsCursor{
		V:   cursorVersion,
		PID: projectID,
		SA:  c.StartedAt.UTC().Format(time.RFC3339Nano),
		SID: c.SessionID,
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// decodeListSessionsCursor parst einen URL-Cursor zurück in den
// typisierten Cursor. Leerer String → nil. Defekter Wert →
// errCursorInvalidLegacy / errCursorInvalidMalformed je nach
// Erkennungsregel aus ADR-0004 §6 / API-Kontrakt §10.3:
//   - v fehlt oder v ∈ {1, 2}: legacy (Pre-§4.3-Format).
//   - v = 3 ohne `pid`/`sid` oder mit anderem `pid` als
//     `requestProjectID`: malformed.
//   - v ∉ {1, 2, 3}: malformed.
func decodeListSessionsCursor(s, requestProjectID string) (*driving.ListSessionsCursor, error) {
	if s == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, errCursorInvalidMalformed
	}
	var probe listSessionsRawCursor
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&probe); err != nil {
		return nil, errCursorInvalidMalformed
	}

	// Version-Klassifikation. v fehlt → 0.1.x-Format → legacy.
	if probe.V == nil {
		return nil, errCursorInvalidLegacy
	}
	switch *probe.V {
	case 1, 2:
		// v=1: 0.1.x mit process_instance_id (kein Project-Scope).
		// v=2: 0.2.0–0.3.x mit durable Sortierung, aber ohne Project-
		// Scope. Beide ab §4.3 dauerhaft abgewiesen.
		return nil, errCursorInvalidLegacy
	case cursorVersion:
		// v3 — weiter unten validieren.
	default:
		return nil, errCursorInvalidMalformed
	}

	if probe.PID == nil || probe.SA == nil || probe.SID == nil || *probe.SID == "" {
		return nil, errCursorInvalidMalformed
	}
	// Project-Scope: ein v3-Cursor aus Project A darf nicht in einem
	// Request mit Project-Kontext B akzeptiert werden.
	if *probe.PID != requestProjectID {
		return nil, errCursorInvalidMalformed
	}
	startedAt, err := time.Parse(time.RFC3339Nano, *probe.SA)
	if err != nil {
		return nil, errCursorInvalidMalformed
	}
	return &driving.ListSessionsCursor{
		StartedAt: startedAt,
		SessionID: *probe.SID,
	}, nil
}

// encodeSessionEventsCursor liefert den base64-url-encodierten
// Event-Cursor (cursor_version 3, mit Collection-Scope
// `(project_id, session_id)`).
func encodeSessionEventsCursor(c *driving.SessionEventsCursor, projectID, sessionID string) (string, error) {
	if c == nil {
		return "", nil
	}
	wire := wireSessionEventsCursor{
		V:   cursorVersion,
		PID: projectID,
		SID: sessionID,
		RCV: c.ServerReceivedAt.UTC().Format(time.RFC3339Nano),
		SEQ: c.SequenceNumber,
		ING: c.IngestSequence,
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// decodeSessionEventsCursor ist das Pendant zu encodeSessionEventsCursor.
// Ein v3-Cursor mit nicht-passendem Project- oder Session-Scope liefert
// `cursor_invalid_malformed` — damit kann ein Event-Cursor aus Session
// A nicht für Session B im selben Project nutzbar sein
// (API-Kontrakt §10.3 / ADR-0004 §6).
func decodeSessionEventsCursor(s, requestProjectID, requestSessionID string) (*driving.SessionEventsCursor, error) {
	if s == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, errCursorInvalidMalformed
	}
	var probe sessionEventsRawCursor
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&probe); err != nil {
		return nil, errCursorInvalidMalformed
	}

	if probe.V == nil {
		return nil, errCursorInvalidLegacy
	}
	switch *probe.V {
	case 1, 2:
		return nil, errCursorInvalidLegacy
	case cursorVersion:
	default:
		return nil, errCursorInvalidMalformed
	}

	if probe.PID == nil || probe.SID == nil || probe.RCV == nil || probe.ING == nil {
		return nil, errCursorInvalidMalformed
	}
	if *probe.PID != requestProjectID {
		return nil, errCursorInvalidMalformed
	}
	if *probe.SID != requestSessionID {
		return nil, errCursorInvalidMalformed
	}
	receivedAt, err := time.Parse(time.RFC3339Nano, *probe.RCV)
	if err != nil {
		return nil, errCursorInvalidMalformed
	}
	return &driving.SessionEventsCursor{
		ServerReceivedAt: receivedAt,
		SequenceNumber:   probe.SEQ,
		IngestSequence:   *probe.ING,
	}, nil
}
