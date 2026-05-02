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

// legacyPlaybackHandler ist der minimale Server-Stub für den
// §3.4b-#1-Vertrag. Er snapshotted **eine einzige** Eigenschaft des
// 0.3.x-Backends: dass `traceparent` im HTTP-Layer nicht gelesen wird.
// Alle anderen 0.3.x-Pflichten (Body-Größe, Detail-Validierung,
// schema_version-Check, Domain-Pipeline, Origin-Bindung, Rate-Limit)
// werden **nicht** gespiegelt — andernfalls würde der Stub zur
// halb-treuen Kopie und ein zukünftiger Reviewer könnte aus einem
// Stub-Verhalten falsche Schlüsse über reales 0.3.x ziehen. Was hier
// fehlt, ist absichtlich nicht da; was hier ist, ist Vertrag.
//
// Der Stub akzeptiert daher jede POST-Anfrage mit Token-Header und
// JSON-parsefähigem Body als 202 — egal welche Header sonst gesetzt
// sind. Damit ist der „unbekannter Header bricht den Server nicht"-
// Vertrag (RFC 9110 §5.1) maschinenlesbar, ohne sich auf weitere
// 0.3.x-Verhaltensdetails zu committen.
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
		// Bewusst kein r.Header.Get("traceparent") — exakt das ist der
		// §3.4b-#1-Snapshot des 0.3.x-Verhaltens.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
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
		w.WriteHeader(http.StatusAccepted)
	})
}
