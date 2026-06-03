package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// OutboundWebhookEvent ist das Wire-Modell, das ein
// `OutboundWebhookDispatcher` an externe Konsumenten zustellt
// (`0.12.5`/RAK-82, R-16). Bewusst minimal — keine Klartext-Keys,
// keine PII; nur die Lifecycle-Felder plus Audit-Metadaten.
//
// Wire-Felder werden vom Adapter im JSON-Body serialisiert:
//
//	{
//	 "event_id": "evt_…",
//	 "type": "stream_started" | "stream_ended",
//	 "project_id": "<id>",
//	 "stream_id": "<id>",
//	 "observed_at": "<RFC3339Nano>",
//	 "source": "local-smoke" | "mediamtx-hook",
//	 "connection_id":"…",
//	 "reason": "…"
//	}
type OutboundWebhookEvent struct {
	EventID      string
	Kind         domain.StreamLifecycleEventKind
	ProjectID    string
	StreamID     string
	ObservedAt   string
	Source       domain.StreamLifecycleEventSource
	ConnectionID string
	Reason       string
}

// OutboundWebhookDispatcher liefert Stream-Lifecycle-Events an einen
// externen Konsumenten (`0.12.5`/RAK-82, R-16). Adapter-Vertrag:
//
//  - `Dispatch` ist eine Best-Effort-Operation. Adapter mit
//  Retry-Schema sollten den Retry intern abwickeln und erst nach
//  Erschöpfung der Versuche einen Fehler zurückliefern (Dead-
//  Letter-Pfad). Der Caller sollte den Fehler loggen, aber den
//  primären Lifecycle-Pfad nicht failen lassen.
//  - `Dispatch` ist idempotent gegenüber Retries seitens des
//  Konsumenten: `event_id` ist opak und eindeutig pro Event,
//  damit der Konsument Replays deduplizieren kann.
//  - Adapter darf den Klartext-Stream-Key **niemals** in der
//  Payload mitschicken — nur Hash/Fingerprint sind erlaubt.
//
// `nil`-Implementierung ist erlaubt und bedeutet „Outbound-Webhook
// deaktiviert" — Caller muss den nil-Check vor dem Aufruf machen.
type OutboundWebhookDispatcher interface {
	Dispatch(ctx context.Context, event OutboundWebhookEvent) error
}
