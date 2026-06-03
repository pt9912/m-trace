// Package mediaserver enthält die externen Media-Server-Adapter
// (R-15). Aktuell nur MediaMTX (HTTP-
// `/v3/config/`-API); SRS bleibt Folge-Item nach `0.12.6`.
package mediaserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// MediaMTXProvisioner implementiert `driven.MediaServerProvisioner`
// gegen die MediaMTX-HTTP-Konfigurations-API
// (`POST /v3/config/paths/add/<name>`, `DELETE /v3/config/paths/
// delete/<name>`). MediaMTX 1.0+ erwartet idempotent PUT/POST mit
// JSON-Body; ein 200/201 signalisiert Erfolg.
//
// Sicherheitsprofil:
//  - Optionaler Bearer-Token für die MediaMTX-`authInternalUsers`-
//  `api`-Action (siehe `examples/srt/mediamtx.yml`); ohne Token
//  ist der Adapter nur in einem Lab-Compose mit IP-Allowlist
//  sicher.
//  - Kein Klartext-Stream-Key im Server-Request — der MediaMTX-
//  Path bekommt nur den SHA-256-Hash als `source`-Marker oder
//  leeren Body, je nach Protocol. Konkrete Wire-Form bleibt
//  bewusst minimal; eine Production-Anbindung verfeinert das
//  pro Operator-Setup.
//
// **Skelett-Charakter**: T9 liefert den Adapter-Pfad und die
// Wire-Form (`provision=true` → `media_server_state`); die
// MediaMTX-Path-Layout-Verfeinerung (Protocols, Auth-Schemas,
// Path-Templates) ist Operator-Detail und wird über `Config`-
// Felder konfigurierbar.
type MediaMTXProvisioner struct {
	httpClient *http.Client
	endpoint   string
	authToken  string
	logger     *slog.Logger
}

// Config bündelt die Boot-Time-Parameter.
type Config struct {
	// Endpoint ist die MediaMTX-API-Basis (z. B. `http://mediamtx:9997`).
	// Pflicht.
	Endpoint string
	// AuthToken wird als `Authorization: Bearer <token>`-Header
	// geschickt. Leer = kein Auth-Header (Lab-Setup).
	AuthToken string
	// HTTPClient erlaubt Test-Injection; Default ist ein eigener
	// Client mit 10-s-Timeout.
	HTTPClient *http.Client
}

// New konstruiert den Adapter. Endpoint ist Pflicht.
func New(cfg Config, logger *slog.Logger) (*MediaMTXProvisioner, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("mediamtx provisioner: endpoint is required")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &MediaMTXProvisioner{
		httpClient: client,
		endpoint:   strings.TrimRight(cfg.Endpoint, "/"),
		authToken:  cfg.AuthToken,
		logger:     logger,
	}, nil
}

// Compile-time check.
var _ driven.MediaServerProvisioner = (*MediaMTXProvisioner)(nil)

// Apply legt den MediaMTX-Path für den Stream an. Wire-Form analog
// MediaMTX-1.x:
//
//	PUT /v3/config/paths/add/<path>
//	{
//	 "source": "publisher",
//	 "sourceProtocol": "<srt|rtmp>"
//	}
//
// MediaMTX antwortet 200/201 auf Erfolg. `409 Conflict` (Path
// existiert) wird als `applied` behandelt (Idempotenz). Andere
// 4xx/5xx → `failed`.
func (p *MediaMTXProvisioner) Apply(ctx context.Context, in driven.MediaServerApplyInput) (driven.MediaServerApplyResult, error) {
	pathName := in.Stream.ID
	url := fmt.Sprintf("%s/v3/config/paths/add/%s", p.endpoint, pathName)
	body := map[string]string{
		"source":         "publisher",
		"sourceProtocol": string(in.Stream.Protocol),
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return driven.MediaServerApplyResult{
			State: driven.MediaServerStateFailed, ErrorCode: "build_request", Detail: err.Error(),
		}, nil
	}
	req.Header.Set("Content-Type", "application/json")
	if p.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.authToken)
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return driven.MediaServerApplyResult{
			State: driven.MediaServerStateFailed, ErrorCode: "unreachable", Detail: trimError(err.Error()),
		}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return driven.MediaServerApplyResult{State: driven.MediaServerStateApplied}, nil
	case resp.StatusCode == http.StatusConflict:
		// Idempotenz: Path war schon angelegt.
		return driven.MediaServerApplyResult{
			State: driven.MediaServerStateApplied, Detail: "path already configured (idempotent)",
		}, nil
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return driven.MediaServerApplyResult{
			State: driven.MediaServerStateFailed, ErrorCode: "auth_failure",
			Detail: trimError(readBodyShort(resp.Body)),
		}, nil
	default:
		return driven.MediaServerApplyResult{
			State: driven.MediaServerStateFailed,
			ErrorCode: fmt.Sprintf("server_status_%d", resp.StatusCode),
			Detail:    trimError(readBodyShort(resp.Body)),
		}, nil
	}
}

// Rollback entfernt den MediaMTX-Path. Best-Effort; Server-Fehler
// loggt der Adapter nur — der Aufrufer sieht den Fehler im Return.
func (p *MediaMTXProvisioner) Rollback(ctx context.Context, projectID, streamID string) error {
	url := fmt.Sprintf("%s/v3/config/paths/delete/%s", p.endpoint, streamID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("mediamtx rollback: build request: %w", err)
	}
	if p.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.authToken)
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mediamtx rollback: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusNotFound {
		body := readBodyShort(resp.Body)
		return fmt.Errorf("mediamtx rollback: status %d, body=%s", resp.StatusCode, body)
	}
	if p.logger != nil {
		p.logger.Info("mediamtx rollback ok",
			"project_id", projectID, "stream_id", streamID, "status", resp.StatusCode)
	}
	return nil
}

// readBodyShort liest die ersten 512 Bytes des Response-Bodies, damit
// Operator-Logs keine ganzen MediaMTX-Stack-Traces enthalten.
func readBodyShort(r io.Reader) string {
	if r == nil {
		return ""
	}
	raw, _ := io.ReadAll(io.LimitReader(r, 512))
	return strings.TrimSpace(string(raw))
}

// trimError schneidet zu lange Detail-Strings ab.
func trimError(s string) string {
	const limit = 256
	if len(s) <= limit {
		return s
	}
	return s[:limit]
}
