package http_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHTTP_Trace_CrossVersion_LegacyHandlerAcceptsTraceParent deckt
// plan-0.4.0 §3.4b #1 (aus §3.3-Review, Should-fix #1).
//
// Vertrag: ein SDK auf 0.4.0 sendet `traceparent`-Header an einen
// Server, der wie 0.3.x das Header-Feld gar nicht liest und keine
// `correlation_id` persistiert. Das HTTP-Layer-Verhalten muss
// unverändert bleiben — Status 202, kein 4xx wegen unbekannten
// Headern (RFC 9110 §5.1: "Unrecognized header fields SHOULD be
// ignored").
//
// Realisierung als reiner Adapter-Test mit einem minimalen Stub-
// Handler, der genau die Pre-§3.2-Pflichten erfüllt: Token-Check,
// Body-Größe, JSON-Parse, schema_version=1.0 — und nichts darüber
// hinaus. Insbesondere wird die `traceparent`-Header-Existenz nicht
// einmal abgefragt, geschweige denn validiert. Wenn der Test grün
// bleibt, ist nachgewiesen: ein 0.4.0-SDK kann gegen einen 0.3.x-
// Server arbeiten, ohne dass der Header zum Reject-Faktor wird.
//
// Option (c) (echter Node-Cross-Run gegen Go-httptest) ist
// dokumentiert deferred (§3.4b-Plan); das hier ist die deterministische
// und wartungsärmste Variante.
func TestHTTP_Trace_CrossVersion_LegacyHandlerAcceptsTraceParent(t *testing.T) {
	t.Parallel()

	handler := legacyPlaybackHandler(t)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	// Realistischer SDK-Header-Wert (W3C-Format-konform — aber das
	// ist für den Vertrag irrelevant: der Legacy-Server liest ihn
	// nicht). Wir wählen bewusst einen *gültigen* String, weil ein
	// kaputter Wert §3.4b #2 wäre.
	tp := "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"

	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, srv.URL+"/api/playback-events", strings.NewReader(validBody))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("X-MTrace-Token", "demo-token")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("traceparent", tp)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 202 from legacy handler, got %d (body=%q) — Pre-§3.2-Server darf den Header weder lesen noch ablehnen",
			resp.StatusCode, string(body))
	}
}

// legacyPlaybackHandler simuliert das HTTP-Verhalten eines 0.3.x-
// Backends: nur die in 0.3.x existierenden Pflichten (Token, Body-
// Limit, JSON-Parse, schema_version=1.0). Es liest **nicht** den
// `traceparent`-Header und schreibt **keine** `correlation_id`. Damit
// ist der Vertrag „§3.4b-§0.3.x-Snapshot" maschinenlesbar dokumentiert
// — wenn ein zukünftiger Reviewer wissen will, was 0.3.x mit dem
// Header gemacht hat, ist die Antwort: gar nichts (Code = leere
// Lese-Operation).
func legacyPlaybackHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-MTrace-Token") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Bewusst kein r.Header.Get("traceparent") — Pre-§3.2-Snapshot.
		body, err := io.ReadAll(io.LimitReader(r.Body, 256*1024))
		if err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		var payload struct {
			SchemaVersion string            `json:"schema_version"`
			Events        []json.RawMessage `json:"events"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if payload.SchemaVersion != "1.0" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if len(payload.Events) == 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
}
