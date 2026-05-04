package domain

import "time"

// SessionBoundary ist ein durabler, session-skopierter Marker, der
// Lücken in der Netzwerk-Beobachtung benennt — z. B. wenn der Browser
// gar kein Manifest-/Segment-Signal liefert (Native HLS, CORS-Block,
// Resource-Timing-Lücke). Boundaries sind keine Events: sie haben
// kein `event_name`, zählen nicht in `accepted` und ändern die
// Batch-`schema_version` nicht. Source spec/telemetry-model.md §1.4
// und plan-0.4.0 §4.4.
//
// In Tranche 3 ist nur Kind="network_signal_absent" definiert; weitere
// Boundary-Kinds erweitern die Domäne additiv.
type SessionBoundary struct {
	// Kind ist der Boundary-Typ — in 0.4.0 ausschließlich
	// "network_signal_absent". Die Validation lehnt andere Werte mit
	// 422 ab; das Feld bleibt String, damit zukünftige Kinds additiv
	// dazukommen können.
	Kind string
	// ProjectID und SessionID müssen eine `(project_id, session_id)`-
	// Partition referenzieren, für die im selben Batch mindestens ein
	// Event vorhanden ist (siehe API-Kontrakt §3.4). ProjectID muss
	// zum aufgelösten Token passen.
	ProjectID string
	SessionID string
	// NetworkKind ist "manifest" oder "segment".
	NetworkKind string
	// Adapter ist "hls.js", "native_hls" oder "unknown".
	Adapter string
	// Reason verwendet denselben normativen Reason-Enum wie
	// `meta["network.unavailable_reason"]` (spec/telemetry-model.md
	// §1.4) plus Pattern `^[a-z0-9_]{1,64}$`.
	Reason string
	// ClientTimestamp ist die SDK-Uhr beim Capability-Erkennen; nicht
	// autoritativ, dient der Timeline-Anzeige.
	ClientTimestamp time.Time
	// ServerReceivedAt setzt der Use-Case beim Akzeptieren des
	// Batches. Durable Sortier-Anker im Read-Pfad.
	ServerReceivedAt time.Time
}

// BoundaryKindNetworkSignalAbsent ist der einzige Boundary-Kind in
// Tranche 3. Wert ist normativ aus
// contracts/event-schema.json#session_boundaries.kinds.
const BoundaryKindNetworkSignalAbsent = "network_signal_absent"
