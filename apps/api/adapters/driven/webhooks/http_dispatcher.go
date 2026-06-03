// Package webhooks implementiert ausgehende Webhook-Adapter
// (`0.12.5`/RAK-82, R-16). Aktuell nur HTTP-POST mit HMAC-Signatur;
// alternative Transport-Adapter (z. B. AWS-SNS, Pub/Sub) bleiben
// Folge-Item nach `0.12.5`.
package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// HTTPDispatcher ist der HTTP-Adapter für
// `driven.OutboundWebhookDispatcher` (`0.12.5`/RAK-82, R-16).
// Adressiert R-16 (lokales Lifecycle-Eventmodell ohne Outbound).
//
// Wire-Vertrag (Doku in `docs/user/auth.md` §5.8):
//  - `POST` an `Endpoint` mit `Content-Type: application/json`.
//  - Body: `OutboundWebhookEvent` serialisiert als JSON (siehe
//  `driven.OutboundWebhookEvent`-Doc).
//  - Signatur: `X-MTrace-Signature: sha256=<hex>` über den
//  gesamten Body, HMAC-SHA-256 mit `Secret`. Der Konsument
//  muss exakt den Body-Hash prüfen; Whitespace-Veränderungen
//  ändern die Signatur.
//  - Timestamp im Header `X-MTrace-Timestamp` (RFC3339Nano),
//  damit der Konsument Replay-Schutz machen kann.
//
// Retry-Schema (statisch, Folge-Item: ENV-konfigurierbar):
//  - bis zu `MaxAttempts` Versuche (Default 3).
//  - Exponential-Backoff mit Basis `BaseBackoff` (Default 100ms),
//  verdoppelt nach jedem Versuch: 100ms, 200ms, 400ms.
//  - Per-Versuch-Timeout `RequestTimeout` (Default 10s).
//  - Erfolg ist `200 ≤ status < 300`; alles andere zählt als
//  fehlgeschlagen und wird nach Backoff erneut versucht.
//  - Nach Erschöpfung der Versuche: `Dispatch` liefert einen
//  `Dead-Letter`-Fehler. Der Caller loggt das, lässt aber den
//  Lifecycle-Pfad nicht failen.
//
// Disabled-Modus: wenn `Endpoint == ""`, ist der Dispatcher ein
// No-Op — `Dispatch` liefert ohne Roundtrip `nil`. Bootstrap kann
// damit den Dispatcher unkonfiguriert lassen und der Caller spart
// sich den nil-Check.
type HTTPDispatcher struct {
	Endpoint       string
	Secret         []byte
	HTTPClient     *http.Client
	Logger         *slog.Logger
	Now            func() time.Time
	MaxAttempts    int
	BaseBackoff    time.Duration
	RequestTimeout time.Duration
}

// NewHTTPDispatcher konstruiert einen einsatzbereiten Dispatcher
// mit Default-Werten für Retry und Timeout. `endpoint == ""`
// markiert den Adapter als disabled (No-Op).
func NewHTTPDispatcher(endpoint string, secret []byte, logger *slog.Logger) *HTTPDispatcher {
	return &HTTPDispatcher{
		Endpoint:       strings.TrimSpace(endpoint),
		Secret:         secret,
		HTTPClient:     newDefaultHTTPClient(),
		Logger:         logger,
		Now:            time.Now,
		MaxAttempts:    3,
		BaseBackoff:    100 * time.Millisecond,
		RequestTimeout: 10 * time.Second,
	}
}

// Compile-time check.
var _ driven.OutboundWebhookDispatcher = (*HTTPDispatcher)(nil)

// ErrOutboundWebhookExhausted markiert den Dead-Letter-Pfad: alle
// `MaxAttempts` Versuche sind ohne Erfolg ausgelaufen. Der Caller
// kann das Sentinel via `errors.Is` erkennen und für Metriken
// auswerten.
var ErrOutboundWebhookExhausted = errors.New("outbound webhook: max attempts exhausted")

// Dispatch implementiert den Driven-Port. Disabled (Endpoint leer) →
// no-op. Sonst: bis zu `MaxAttempts` Versuche mit Exponential-Backoff.
func (d *HTTPDispatcher) Dispatch(ctx context.Context, event driven.OutboundWebhookEvent) error {
	if d == nil || d.Endpoint == "" {
		return nil
	}
	payload, err := marshalEvent(event)
	if err != nil {
		return fmt.Errorf("outbound webhook: marshal event: %w", err)
	}
	signature := computeSignature(d.Secret, payload)
	now := d.now()

	var lastErr error
	maxAttempts := d.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if attempt > 1 {
			backoff := d.backoffFor(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
		err := d.sendOnce(ctx, payload, signature, now)
		if err == nil {
			return nil
		}
		lastErr = err
		if d.Logger != nil {
			d.Logger.Warn("outbound webhook attempt failed",
				"event_id", event.EventID,
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"error", err.Error(),
			)
		}
	}
	if d.Logger != nil {
		d.Logger.Error("outbound webhook dead-letter",
			"event_id", event.EventID,
			"endpoint", d.Endpoint,
			"last_error", lastErr.Error(),
		)
	}
	return fmt.Errorf("%w: last_error=%v", ErrOutboundWebhookExhausted, lastErr)
}

func (d *HTTPDispatcher) sendOnce(ctx context.Context, payload []byte, signature string, now time.Time) error {
	reqCtx, cancel := context.WithTimeout(ctx, d.requestTimeout())
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, d.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MTrace-Signature", "sha256="+signature)
	req.Header.Set("X-MTrace-Timestamp", now.UTC().Format(time.RFC3339Nano))
	resp, err := d.client().Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Body abräumen, damit der Connection-Pool wiederverwendet
		// werden kann.
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("status %d body=%q", resp.StatusCode, strings.TrimSpace(string(body)))
}

func (d *HTTPDispatcher) backoffFor(attempt int) time.Duration {
	base := d.BaseBackoff
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	mult := 1 << (attempt - 2)
	if mult < 1 {
		mult = 1
	}
	return base * time.Duration(mult)
}

func (d *HTTPDispatcher) client() *http.Client {
	if d.HTTPClient != nil {
		return d.HTTPClient
	}
	return newDefaultHTTPClient()
}

func newDefaultHTTPClient() *http.Client {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Client{Timeout: 10 * time.Second}
	}
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport.Clone(),
	}
}

func (d *HTTPDispatcher) requestTimeout() time.Duration {
	if d.RequestTimeout > 0 {
		return d.RequestTimeout
	}
	return 10 * time.Second
}

func (d *HTTPDispatcher) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now()
}

// marshalEvent serialisiert das Event ins Wire-Format (siehe
// Doc-Comment am `HTTPDispatcher`).
func marshalEvent(e driven.OutboundWebhookEvent) ([]byte, error) {
	body := map[string]any{
		"event_id":      e.EventID,
		"type":          string(e.Kind),
		"project_id":    e.ProjectID,
		"stream_id":     e.StreamID,
		"observed_at":   e.ObservedAt,
		"source":        string(e.Source),
		"connection_id": e.ConnectionID,
		"reason":        e.Reason,
	}
	return json.Marshal(body)
}

// computeSignature liefert den HMAC-SHA-256-Hex über `payload`,
// signiert mit `secret`. Caller setzt den `sha256=…`-Prefix.
func computeSignature(secret, payload []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
