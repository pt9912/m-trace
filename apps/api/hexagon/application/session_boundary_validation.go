package application

import (
	"fmt"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// MaxSessionBoundaries ist die obere Grenze für `session_boundaries[]`
// pro Batch aus contracts/event-schema.json#batch.max_session_boundaries
// und API-Kontrakt §3.4.
const MaxSessionBoundaries = 20

// isReservedBoundaryKind erlaubt aktuell nur "network_signal_absent"
// (contracts/event-schema.json#session_boundaries.kinds). Zukünftige
// Kinds erweitern den switch additiv.
func isReservedBoundaryKind(s string) bool {
	switch s {
	case domain.BoundaryKindNetworkSignalAbsent:
		return true
	default:
		return false
	}
}

// isReservedBoundaryAdapter spiegelt
// contracts/event-schema.json#session_boundaries.adapters.
func isReservedBoundaryAdapter(s string) bool {
	switch s {
	case "hls.js", "native_hls", "unknown":
		return true
	default:
		return false
	}
}

// parseAndValidateBoundaries normalisiert das wire-Boundary-Array zu
// domain.SessionBoundary, erzwingt Pflichtfelder, Reason-Enum/Pattern
// und Partition-Match gegen die Event-Sessions des Batches. Verstöße
// liefern domain.ErrInvalidEvent (mappt auf 422); der Aufrufer ist
// dafür zuständig, kein Persist auszuführen, wenn der Validator einen
// Fehler liefert.
//
// `eventSessions` ist die Menge `(project_id, session_id)`-Partitionen,
// für die mindestens ein Event im selben Batch existiert — der
// Use-Case baut sie aus den geparsten Events.
func parseAndValidateBoundaries(
	in []driving.BoundaryInput,
	projectID string,
	eventSessions map[string]struct{},
	now time.Time,
) ([]domain.SessionBoundary, error) {
	if len(in) == 0 {
		return nil, nil
	}
	if len(in) > MaxSessionBoundaries {
		return nil, fmt.Errorf(
			"%w: session_boundaries exceeds %d",
			domain.ErrInvalidEvent, MaxSessionBoundaries,
		)
	}
	out := make([]domain.SessionBoundary, 0, len(in))
	for _, b := range in {
		boundary, err := parseBoundary(b, projectID, eventSessions, now)
		if err != nil {
			return nil, err
		}
		out = append(out, boundary)
	}
	return out, nil
}

// parseBoundary validiert einen einzelnen Boundary-Eintrag.
func parseBoundary(
	b driving.BoundaryInput,
	projectID string,
	eventSessions map[string]struct{},
	now time.Time,
) (domain.SessionBoundary, error) {
	if err := requireBoundaryFields(b); err != nil {
		return domain.SessionBoundary{}, err
	}
	if b.ProjectID != projectID {
		// Boundary-Project-Mismatch ist 422 (Validierungsfehler), nicht
		// 401 — anders als bei Event-Project-Mismatch wird der Token-
		// Project-Bindung-Pfad bereits durch die Event-Schleife
		// geprüft. Hier hat der Block einen inkonsistenten Wert in
		// einem Wrapper-Feld.
		return domain.SessionBoundary{}, fmt.Errorf(
			"%w: boundary.project_id %q must match resolved project %q",
			domain.ErrInvalidEvent, b.ProjectID, projectID,
		)
	}
	if !isReservedBoundaryKind(b.Kind) {
		return domain.SessionBoundary{}, fmt.Errorf(
			"%w: boundary.kind %q not in reserved domain",
			domain.ErrInvalidEvent, b.Kind,
		)
	}
	if !isReservedNetworkKindValue(b.NetworkKind) {
		return domain.SessionBoundary{}, fmt.Errorf(
			"%w: boundary.network_kind %q not in {manifest, segment}",
			domain.ErrInvalidEvent, b.NetworkKind,
		)
	}
	if !isReservedBoundaryAdapter(b.Adapter) {
		return domain.SessionBoundary{}, fmt.Errorf(
			"%w: boundary.adapter %q not in {hls.js, native_hls, unknown}",
			domain.ErrInvalidEvent, b.Adapter,
		)
	}
	if !networkUnavailableReasonPattern.MatchString(b.Reason) {
		return domain.SessionBoundary{}, fmt.Errorf(
			"%w: boundary.reason violates ^[a-z0-9_]{1,64}$",
			domain.ErrInvalidEvent,
		)
	}
	if !isNetworkUnavailableReason(b.Reason) {
		return domain.SessionBoundary{}, fmt.Errorf(
			"%w: boundary.reason %q not in reserved reason enum",
			domain.ErrInvalidEvent, b.Reason,
		)
	}
	if _, ok := eventSessions[b.SessionID]; !ok {
		return domain.SessionBoundary{}, fmt.Errorf(
			"%w: boundary references session %q without event in same batch",
			domain.ErrInvalidEvent, b.SessionID,
		)
	}
	ts, err := time.Parse(time.RFC3339Nano, b.ClientTimestamp)
	if err != nil {
		return domain.SessionBoundary{}, fmt.Errorf(
			"%w: boundary.client_timestamp not RFC3339Nano",
			domain.ErrInvalidEvent,
		)
	}
	return domain.SessionBoundary{
		Kind:             b.Kind,
		ProjectID:        b.ProjectID,
		SessionID:        b.SessionID,
		NetworkKind:      b.NetworkKind,
		Adapter:          b.Adapter,
		Reason:           b.Reason,
		ClientTimestamp:  ts,
		ServerReceivedAt: now,
	}, nil
}

// requireBoundaryFields prüft die Pflichtfelder aus
// contracts/event-schema.json#session_boundaries.required_fields.
func requireBoundaryFields(b driving.BoundaryInput) error {
	if b.Kind == "" || b.ProjectID == "" || b.SessionID == "" ||
		b.NetworkKind == "" || b.Adapter == "" || b.Reason == "" ||
		b.ClientTimestamp == "" {
		return fmt.Errorf("%w: boundary missing required field", domain.ErrInvalidEvent)
	}
	return nil
}
