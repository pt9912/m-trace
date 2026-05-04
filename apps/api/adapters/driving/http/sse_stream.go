package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// SSE-Backfill-Limit aus spec/backend-api-contract.md §10a. Bei
// größerer Lücke öffnet der Server den Stream mit einem
// `backfill_truncated`-Frame und liefert nur die jüngsten 1000
// Events.
const sseBackfillLimit = 1000

// SSE-Heartbeat-Intervall aus spec §10a. Der Server pushed alle 15s
// einen `: heartbeat`-Comment, damit Proxies den Stream nicht als
// idle-Timeout schließen.
const sseHeartbeatInterval = 15 * time.Second

// SseStreamHandler bedient `GET /api/stream-sessions/stream`. Auth
// ist tokenpflichtig (Spec §10a); Stream und Backfill sind auf das
// aufgelöste Project skopiert.
type SseStreamHandler struct {
	Resolver  driven.ProjectResolver
	Events    driven.EventRepository
	Broker    *application.EventBroker
	Logger    *slog.Logger
	Heartbeat time.Duration
	// BackfillLimit überschreibt den Default
	// `sseBackfillLimit` (1000 Events). Tests setzen das niedriger,
	// um den Truncation-Pfad ohne 1000-Events-Setup zu erreichen.
	BackfillLimit int
}

// ServeHTTP implementiert net/http.Handler.
func (h *SseStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeStatus(w, http.StatusMethodNotAllowed)
		return
	}

	projectID, ok := resolveProjectFromToken(r.Context(), w, r, h.Resolver)
	if !ok {
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		// `httptest.NewRecorder` liefert keinen Flusher; Production
		// (`net/http.ResponseWriter`) tut. 500 ist defensiv.
		h.logError("ResponseWriter does not implement http.Flusher")
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	subCtx, cancel := context.WithCancel(r.Context())
	defer cancel()
	frames := h.Broker.Subscribe(subCtx, projectID)

	// Backfill aus EventRepository, falls Last-Event-ID gesetzt.
	if err := h.streamBackfill(subCtx, w, flusher, projectID, r.Header.Get("Last-Event-ID")); err != nil {
		h.logError("backfill failed", "error", err)
		return
	}

	heartbeat := h.Heartbeat
	if heartbeat <= 0 {
		heartbeat = sseHeartbeatInterval
	}
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case frame, ok := <-frames:
			if !ok {
				return
			}
			if err := writeAppendedFrame(w, flusher, frame); err != nil {
				h.logError("write append frame failed", "error", err)
				return
			}
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// streamBackfill liest alle Events des Projects mit
// `ingest_sequence > Last-Event-ID` aus der Persistenz und schreibt
// sie als SSE-Frames; der Konsument empfängt sie vor dem ersten
// Live-Frame. Ohne Header → leer; mit ungültigem Wert → leer (kein
// 4xx, weil EventSource-Browsern das Header automatisch setzen).
func (h *SseStreamHandler) streamBackfill(
	ctx context.Context, w http.ResponseWriter, flusher http.Flusher,
	projectID, lastEventIDRaw string,
) error {
	if lastEventIDRaw == "" {
		return nil
	}
	lastEventID, err := strconv.ParseInt(lastEventIDRaw, 10, 64)
	if err != nil || lastEventID < 0 {
		// Ungültiger Header → ignorieren; Spec sagt, EventSource
		// setzt den Header automatisch beim Reconnect. Ein Bug im
		// Konsumenten darf den Stream nicht 4xx-en.
		return nil
	}
	limit := h.BackfillLimit
	if limit <= 0 {
		limit = sseBackfillLimit
	}
	events, err := h.Events.ListAfterIngestSequence(ctx, projectID, lastEventID, limit)
	if err != nil {
		return err
	}
	// Truncation-Marker, wenn das Backfill-Limit voll ausgeschöpft ist.
	// Konsumenten müssen dann den Detail-Snapshot neu laden.
	if len(events) == limit {
		if err := writeTruncatedFrame(w, flusher, events[0].IngestSequence); err != nil {
			return err
		}
	}
	for _, e := range events {
		frame := application.EventAppendedFrame{
			ProjectID:      e.ProjectID,
			SessionID:      e.SessionID,
			EventName:      e.EventName,
			IngestSequence: e.IngestSequence,
		}
		if err := writeAppendedFrame(w, flusher, frame); err != nil {
			return err
		}
	}
	return nil
}

func (h *SseStreamHandler) logError(msg string, args ...any) {
	if h.Logger == nil {
		return
	}
	h.Logger.Error(msg, args...)
}

// writeAppendedFrame serialisiert einen `event_appended`-Frame in
// EventSource-Format. Spec §10a verlangt exakt
// `id`/`event`/`data`-Feldsatz mit Project-/Session-/Sequence-/
// Event-Name-Daten.
func writeAppendedFrame(w http.ResponseWriter, flusher http.Flusher, frame application.EventAppendedFrame) error {
	body, err := json.Marshal(frameWire{
		ProjectID:      frame.ProjectID,
		SessionID:      frame.SessionID,
		IngestSequence: frame.IngestSequence,
		EventName:      frame.EventName,
	})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "id: %d\nevent: event_appended\ndata: %s\n\n",
		frame.IngestSequence, body); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

// writeTruncatedFrame markiert die Backfill-Truncation: Konsument
// muss den Detail-Snapshot per REST neu laden, der älteste
// gelieferte `ingest_sequence` ist im Frame.
func writeTruncatedFrame(w http.ResponseWriter, flusher http.Flusher, oldestSeq int64) error {
	body, err := json.Marshal(map[string]int64{"oldest_ingest_sequence": oldestSeq})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: backfill_truncated\ndata: %s\n\n", body); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

type frameWire struct {
	ProjectID      string `json:"project_id"`
	SessionID      string `json:"session_id"`
	IngestSequence int64  `json:"ingest_sequence"`
	EventName      string `json:"event_name"`
}
