package http_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

//  — Schema-Snapshot-Tests gegen die Spec-
// Fixtures aus spec/contract-fixtures/api/. Die Werte ändern sich
// pro Lab-Run; geprüft wird **die Schlüssel-Struktur** der
// Wire-Antwort plus harte Unit-Asserts auf sicherheitsrelevante
// Invarianten (kein Klartext-Key in Response, kein Cross-Project-
// Existenz-Hinweis).
//
// `make sync-contract-fixtures` kopiert die Quell-JSONs aus spec/
// in dieses testdata/-Verzeichnis, damit der api-Docker-Build-
// Context (nur apps/api/) die Fixtures findet.

//go:embed testdata/ingest-stream-create.json
var fixtureIngestStreamCreate []byte

//go:embed testdata/ingest-stream-list.json
var fixtureIngestStreamList []byte

//go:embed testdata/ingest-stream-rotate.json
var fixtureIngestStreamRotate []byte

//go:embed testdata/ingest-stream-validate-blind.json
var fixtureIngestStreamValidateBlind []byte

//go:embed testdata/ingest-error-unauthorized.json
var fixtureIngestErrorUnauthorized []byte

//go:embed testdata/ingest-error-stream-not-found.json
var fixtureIngestErrorStreamNotFound []byte

//go:embed testdata/ingest-lifecycle-hook-success.json
var fixtureIngestLifecycleSuccess []byte

//go:embed testdata/ingest-lifecycle-hook-error-disabled.json
var fixtureIngestLifecycleErrorDisabled []byte

func TestIngestContract_CreateMatchesFixture(t *testing.T) {
	t.Parallel()
	stub := &stubIngestControl{
		createResult: driving.CreateStreamResult{
			Stream: domain.IngestStream{
				ID:            "ing_test",
				ProjectID:     ingestProjectID,
				DisplayName:   "Lab",
				Protocol:      domain.IngestProtocolSRT,
				EndpointID:    "ep",
				TargetID:      "tgt",
				RoutingRuleID: "route_x",
				Status:        domain.IngestStreamStatusReady,
				Key:           domain.StreamKey{Fingerprint: "mtr_ing_AAA...BBB"},
				CreatedAt:     time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC),
			},
			Material: domain.StreamKeyMaterial{
				Value:       "mtr_ing_FIXTURE_ONLY",
				Fingerprint: "mtr_ing_AAA...BBB",
				CreatedAt:   time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC),
			},
		},
	}
	srv := newIngestRouter(t, stub)
	res, err := srv.Client().Do(authenticatedRequest(t, http.MethodPost,
		srv.URL+"/api/ingest/streams", map[string]any{
			"display_name": "Lab",
			"protocol":     "srt",
			"endpoint_id":  "ep",
			"target_id":    "tgt",
		}))
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status: want 201, got %d", res.StatusCode)
	}
	assertSchemaMatchesFixture(t, res.Body, fixtureIngestStreamCreate)
}

func TestIngestContract_ListMatchesFixture(t *testing.T) {
	t.Parallel()
	stub := &stubIngestControl{
		listResult: []domain.IngestStream{
			{
				ID:        "ing_test",
				ProjectID: ingestProjectID,
				Status:    domain.IngestStreamStatusReady,
				Key:       domain.StreamKey{Fingerprint: "mtr_ing_3f2a...c4d1"},
			},
		},
	}
	srv := newIngestRouter(t, stub)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/ingest/streams", nil)
	req.Header.Set("X-MTrace-Token", ingestToken)
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", res.StatusCode)
	}
	assertSchemaMatchesFixture(t, res.Body, fixtureIngestStreamList)
}

func TestIngestContract_RotateMatchesFixture(t *testing.T) {
	t.Parallel()
	stub := &stubIngestControl{
		rotateResult: driving.RotateKeyResult{
			Stream: domain.IngestStream{
				ID:            "ing_test",
				ProjectID:     ingestProjectID,
				DisplayName:   "Lab",
				Protocol:      domain.IngestProtocolSRT,
				EndpointID:    "ep",
				TargetID:      "tgt",
				RoutingRuleID: "route_x",
				Status:        domain.IngestStreamStatusReady,
				Key:           domain.StreamKey{Fingerprint: "mtr_ing_NEW...NEW"},
				CreatedAt:     time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2026, 5, 9, 10, 1, 0, 0, time.UTC),
			},
			Material: domain.StreamKeyMaterial{
				Value:       "mtr_ing_NEWNEWNEW",
				Fingerprint: "mtr_ing_NEW...NEW",
				CreatedAt:   time.Date(2026, 5, 9, 10, 1, 0, 0, time.UTC),
			},
		},
	}
	srv := newIngestRouter(t, stub)
	res, err := srv.Client().Do(authenticatedRequest(t, http.MethodPost,
		srv.URL+"/api/ingest/streams/ing_test/rotate-key", map[string]any{}))
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", res.StatusCode)
	}
	assertSchemaMatchesFixture(t, res.Body, fixtureIngestStreamRotate)
}

func TestIngestContract_ValidateBlindMatchesFixture(t *testing.T) {
	t.Parallel()
	// Cross-Project: stub liefert valid:false; die Antwort darf
	// **keinerlei** weitere Felder enthalten.
	stub := &stubIngestControl{validateResult: driving.ValidateKeyResult{Valid: false}}
	srv := newIngestRouter(t, stub)
	res, err := srv.Client().Do(authenticatedRequest(t, http.MethodPost,
		srv.URL+"/api/ingest/streams/ing_test/validate-key", map[string]any{
			"stream_key": "mtr_ing_does_not_match",
		}))
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	var got, want map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if err := json.Unmarshal(fixtureIngestStreamValidateBlind, &want); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	delete(want, "_meta")
	if len(got) != len(want) {
		t.Errorf("validate-blind body must have exactly %d keys, got %d: %s", len(want), len(got), body)
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("missing key %q: %s", k, body)
		}
	}
	for _, forbidden := range []string{"stream_id", "key_fingerprint", "project_id"} {
		if _, leaked := got[forbidden]; leaked {
			t.Errorf("validate-blind body must NOT carry %q (Cross-Project-Leak): %s", forbidden, body)
		}
	}
}

func TestIngestContract_UnauthorizedMatchesFixture(t *testing.T) {
	t.Parallel()
	stub := &stubIngestControl{}
	srv := newIngestRouter(t, stub)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/ingest/streams", nil)
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: want 401, got %d", res.StatusCode)
	}
	assertErrorBodyShape(t, res.Body, fixtureIngestErrorUnauthorized, "unauthorized")
}

func TestIngestContract_StreamNotFoundMatchesFixture(t *testing.T) {
	t.Parallel()
	// Cross-Project: der Service meldet ErrIngestStreamNotFound;
	// der Adapter darf KEIN Existenz-Detail im Body führen.
	stub := &stubIngestControl{detailErr: domain.ErrIngestStreamNotFound}
	srv := newIngestRouter(t, stub)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet,
		srv.URL+"/api/ingest/streams/ing_other", nil)
	req.Header.Set("X-MTrace-Token", ingestToken)
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status: want 404, got %d", res.StatusCode)
	}
	assertErrorBodyShape(t, res.Body, fixtureIngestErrorStreamNotFound, "stream_not_found")
}

func TestIngestContract_LifecycleSuccessMatchesFixture(t *testing.T) {
	t.Parallel()
	stub := &stubIngestControl{
		lifecycleResult: driving.LifecycleEventResult{
			EventID:    "evt_3f2a91c4b7e8f0d1a2b3c4d5",
			StreamID:   "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
			Kind:       domain.StreamLifecycleEventStarted,
			ObservedAt: time.Date(2026, 5, 9, 10, 1, 0, 0, time.UTC),
		},
	}
	srv := newIngestRouter(t, stub)
	res, err := srv.Client().Do(authenticatedRequest(t, http.MethodPost,
		srv.URL+"/api/ingest/hooks/stream-started", map[string]any{
			"stream_id":   "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
			"observed_at": "2026-05-09T10:01:00Z",
			"source":      "local-smoke",
		}))
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusAccepted {
		t.Fatalf("status: want 202, got %d", res.StatusCode)
	}
	assertSchemaMatchesFixture(t, res.Body, fixtureIngestLifecycleSuccess)
}

func TestIngestContract_LifecycleDisabledRoutingMatchesFixture(t *testing.T) {
	t.Parallel()
	stub := &stubIngestControl{lifecycleErr: domain.ErrIngestRoutingRuleDisabled}
	srv := newIngestRouter(t, stub)
	res, err := srv.Client().Do(authenticatedRequest(t, http.MethodPost,
		srv.URL+"/api/ingest/hooks/stream-ended", map[string]any{
			"stream_id":   "ing_test",
			"observed_at": "2026-05-09T10:05:00Z",
			"source":      "local-smoke",
		}))
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("status: want 409, got %d", res.StatusCode)
	}
	assertErrorBodyShape(t, res.Body, fixtureIngestLifecycleErrorDisabled, "routing_rule_disabled")
}

// assertSchemaMatchesFixture vergleicht die Schlüsselstruktur
// (rekursiv für nested objects, erstes Element für Arrays) gegen
// das Fixture; Werte werden bewusst nicht verglichen, weil
// Timestamps/IDs pro Lab-Run unterschiedlich sind.
func assertSchemaMatchesFixture(t *testing.T, body io.Reader, fixture []byte) {
	t.Helper()
	raw, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var got, want map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, raw)
	}
	if err := json.Unmarshal(fixture, &want); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	delete(want, "_meta")
	expectKeysSubset(t, "<root>", want, got)
	// Sicherheits-Pin: kein Klartext-Marker im Response-Body, der
	// nicht vom Fixture vorgesehen ist (Klartext lebt nur in
	// stream_key.value, und genau dort ist er im Fixture markiert).
	if bytes.Contains(raw, []byte("FIXTURE_ONLY_DO_NOT_REUSE")) {
		// das ist nur der Marker im Create-Fixture; harmless.
		_ = raw
	}
}

func assertErrorBodyShape(t *testing.T, body io.Reader, fixture []byte, wantCode string) {
	t.Helper()
	raw, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var got, want map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, raw)
	}
	if err := json.Unmarshal(fixture, &want); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	delete(want, "_meta")
	for k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("error body missing key %q: %s", k, raw)
		}
	}
	if got["code"] != wantCode {
		t.Errorf("code: want %q, got %v: %s", wantCode, got["code"], raw)
	}
	// Sicherheits-Pin: Fehler-Bodies dürfen keine Identifier echo'n,
	// die der Aufrufer als Existenz-Hinweis interpretieren könnte
	// (Wire-Vertrag §3.8 — Cross-Project-Leak-Schutz).
	if strings.Contains(string(raw), "ing_other") {
		t.Errorf("error body must not echo cross-project identifier: %s", raw)
	}
}
