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
	// errCursorInvalidLegacy: Token decodiert; `v`-Feld fehlt oder ist
	// 1, oder `pid`-Feld vorhanden (Hinweis auf 0.1.x/0.2.x/0.3.x).
	// Dauerhaft abgewiesen — kein One-Shot-Grace-Pfad.
	errCursorInvalidLegacy = errors.New("cursor_invalid_legacy")
	// errCursorInvalidMalformed: Base64-/JSON-Decode schlägt fehl;
	// `v`-Feld unbekannt; Pflichtfeld fehlt oder hat ungültiges
	// Format; oder unbekannte Zusatzfelder vorhanden.
	errCursorInvalidMalformed = errors.New("cursor_invalid_malformed")
	// errCursorExpired: Token decodiert valide, aber Storage-Position
	// existiert nicht mehr (Retention/Wipe). In 0.4.0 ohne TTL nur via
	// `make wipe` erreichbar — Code-Pfad ist dennoch vorgesehen, damit
	// Retention-Folge-Arbeit ohne Wire-Format-Bruch möglich bleibt.
	errCursorExpired = errors.New("cursor_expired")
)

// cursorVersion ist die einzige unterstützte Cursor-Version in 0.4.0.
// Tokens mit fehlendem oder anderem `v`-Wert werden als Legacy oder
// Malformed klassifiziert (siehe ADR-0004 §6).
const cursorVersion = 2

// wireListSessionsCursor ist die JSON-Form von
// driving.ListSessionsCursor. Felder bewusst kurz, damit der Cursor
// im URL-Query klein bleibt. `V` ist Pflicht und steuert die
// Versions-Erkennung (ADR-0004 §5).
type wireListSessionsCursor struct {
	V   int    `json:"v"`
	SA  string `json:"sa"`  // started_at, RFC3339Nano
	SID string `json:"sid"` // session_id
}

// wireSessionEventsCursor ist die JSON-Form von
// driving.SessionEventsCursor.
type wireSessionEventsCursor struct {
	V   int    `json:"v"`
	RCV string `json:"rcv"` // server_received_at, RFC3339Nano
	SEQ *int64 `json:"seq,omitempty"`
	ING int64  `json:"ing"` // ingest_sequence
}

// listSessionsRawCursor ist eine Decode-Zwischenform mit allen
// historisch bekannten Feldern (inkl. `pid` aus 0.1.x). Damit erkennt
// der Decoder Legacy-Cursor deterministisch.
type listSessionsRawCursor struct {
	V   *int    `json:"v,omitempty"`
	PID *string `json:"pid,omitempty"`
	SA  *string `json:"sa,omitempty"`
	SID *string `json:"sid,omitempty"`
}

type sessionEventsRawCursor struct {
	V   *int    `json:"v,omitempty"`
	PID *string `json:"pid,omitempty"`
	RCV *string `json:"rcv,omitempty"`
	SEQ *int64  `json:"seq,omitempty"`
	ING *int64  `json:"ing,omitempty"`
}

// encodeListSessionsCursor liefert den base64-url-encodierten Cursor
// für die Sessions-Listenpage (cursor_version 2). Verwendet
// base64.RawURLEncoding (kein Padding, URL-sicher).
func encodeListSessionsCursor(c *driving.ListSessionsCursor) (string, error) {
	if c == nil {
		return "", nil
	}
	wire := wireListSessionsCursor{
		V:   cursorVersion,
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
// Erkennungsregel aus ADR-0004 §6.
func decodeListSessionsCursor(s string) (*driving.ListSessionsCursor, error) {
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
		// Unknown-fields-Fehler ist ebenfalls "malformed" — andere
		// Cursor-Versionen mit zusätzlichen Feldern werden bewusst
		// nicht still toleriert (ADR-0004 §5).
		return nil, errCursorInvalidMalformed
	}

	// Legacy-Erkennung: PID vorhanden ODER v fehlt/=1.
	if probe.PID != nil {
		return nil, errCursorInvalidLegacy
	}
	if probe.V == nil || *probe.V == 1 {
		return nil, errCursorInvalidLegacy
	}
	if *probe.V != cursorVersion {
		return nil, errCursorInvalidMalformed
	}

	if probe.SA == nil || probe.SID == nil || *probe.SID == "" {
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
// Event-Cursor (cursor_version 2).
func encodeSessionEventsCursor(c *driving.SessionEventsCursor) (string, error) {
	if c == nil {
		return "", nil
	}
	wire := wireSessionEventsCursor{
		V:   cursorVersion,
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
func decodeSessionEventsCursor(s string) (*driving.SessionEventsCursor, error) {
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

	if probe.PID != nil {
		return nil, errCursorInvalidLegacy
	}
	if probe.V == nil || *probe.V == 1 {
		return nil, errCursorInvalidLegacy
	}
	if *probe.V != cursorVersion {
		return nil, errCursorInvalidMalformed
	}

	if probe.RCV == nil || probe.ING == nil {
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
