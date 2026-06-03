package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// maxAuthRequestBytes ist die Defense-in-Depth-Grenze für den
// Body von `POST /api/auth/session-tokens` (RAK-72).
// 4 KiB reichen weit für die spec-definierten Issuance-Felder; alles
// darüber wird mit `413` abgewiesen, bevor der JSON-Parser läuft.
const maxAuthRequestBytes = 4 * 1024

// AuthSessionTokensHandler implementiert
// `POST /api/auth/session-tokens` aus `spec/backend-api-contract.md`
// §3.9. Validierungsreihenfolge:
//
//  1. Content-Type → `415`
//  2. Body-Größe → `413`
//  3. JSON-Parsing → `400 invalid_json`
//  4. `X-MTrace-Token` (Project Token) auflösen → `401 auth_token_*`
//  5. `project_id`-Konsistenzcheck → `401 auth_project_mismatch`
//  6. Audience gegen Project-Policy → `403 auth_session_scope_denied`
//  7. TTL-Auflösung → `422 auth_token_ttl_too_large`
//  8. Issuance-Rate-Limit → `429 auth_issuance_rate_limited`
//  9. Use-Case → `201` mit Klartext-Token (genau einmal)
type AuthSessionTokensHandler struct {
	UseCase  driving.AuthSessionInbound
	Resolver driven.ProjectResolver
	Logger   *slog.Logger
}

// ServeHTTP implementiert net/http.Handler.
func (h *AuthSessionTokensHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAuthProblem(w, http.StatusMethodNotAllowed, "method_not_allowed",
			"POST /api/auth/session-tokens erwartet POST.")
		return
	}
	if !ensureAuthContentType(w, r) {
		return
	}
	body, err := readAuthBody(w, r)
	if err != nil {
		return
	}
	// §3.9-Validierungsreihenfolge: Content-Type → Body-Size → JSON →
	// Auth → Konsistenz → Audience → TTL → Rate-Limit. JSON-Parse
	// muss VOR der Auth-Resolution laufen, damit ein Request ohne
	// Token plus kaputtem JSON `400 invalid_json` liefert (nicht
	// `401 auth_token_missing`).
	var req struct {
		ProjectID  string `json:"project_id,omitempty"`
		SessionID  string `json:"session_id,omitempty"`
		Origin     string `json:"origin,omitempty"`
		Audience   string `json:"audience"`
		TTLSeconds int    `json:"ttl_seconds,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeAuthProblem(w, http.StatusBadRequest, "invalid_json", "Body ist kein gültiges JSON.")
		return
	}
	resolvedProjectID, ok := resolveProjectFromMTraceToken(w, r, h.Resolver)
	if !ok {
		return
	}
	result, err := h.UseCase.IssueSessionToken(r.Context(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   resolvedProjectID,
		RequestProjectID:    req.ProjectID,
		Audience:            req.Audience,
		RequestedTTLSeconds: req.TTLSeconds,
		SessionID:           req.SessionID,
		Origin:              req.Origin,
	})
	if err != nil {
		writeAuthError(w, h.Logger, err)
		return
	}
	tokenPayload := map[string]any{
		"value":      result.Value,
		"token_id":   result.TokenID,
		"project_id": result.ProjectID,
		"audience":   string(result.Audience),
		"expires_at": result.ExpiresAt.UTC().Format(time.RFC3339),
	}
	if result.SessionID != "" {
		tokenPayload["session_id"] = result.SessionID
	}
	writeJSON(w, http.StatusCreated, map[string]any{"session_token": tokenPayload})
}

// resolveProjectFromMTraceToken liest `X-MTrace-Token` und löst den
// Project-Token gegen den Static-Resolver auf.
// nutzt für die Issuance ausschließlich den Legacy-Header — Bearer-/
// Session-Token-Pfade sind nur für Konsum-Endpoints relevant
// (RAK-72: ein Session Token darf keinen weiteren Session Token
// minten). Fehlt der Header oder ist ungültig, antwortet die Funktion
// mit `401 auth_token_missing` bzw. `401 auth_token_invalid`.
func resolveProjectFromMTraceToken(w http.ResponseWriter, r *http.Request, resolver driven.ProjectResolver) (string, bool) {
	token := strings.TrimSpace(r.Header.Get("X-MTrace-Token"))
	if token == "" || resolver == nil {
		writeAuthProblem(w, http.StatusUnauthorized, "auth_token_missing",
			"`X-MTrace-Token` fehlt.")
		return "", false
	}
	project, err := resolver.ResolveByToken(r.Context(), token)
	if err != nil {
		writeAuthProblem(w, http.StatusUnauthorized, "auth_token_invalid",
			"`X-MTrace-Token` ist ungültig.")
		return "", false
	}
	return project.ID, true
}

// ensureAuthContentType erzwingt `application/json`. Spiegelt das
// bestehende Verhalten von `ensureContentType` aus dem Ingest-
// Handler.
func ensureAuthContentType(w http.ResponseWriter, r *http.Request) bool {
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	mainType := strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	if mainType != "application/json" {
		writeAuthProblem(w, http.StatusUnsupportedMediaType, "unsupported_media_type",
			"Content-Type muss application/json sein.")
		return false
	}
	return true
}

// readAuthBody enforced die `4 KiB`-Body-Grenze. Bodies darüber
// liefern `413 payload_too_large` ohne Body-Read.
func readAuthBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	limited := io.LimitReader(r.Body, maxAuthRequestBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		writeAuthProblem(w, http.StatusBadRequest, "invalid_request", "Request-Body konnte nicht gelesen werden.")
		return nil, err
	}
	if int64(len(raw)) > maxAuthRequestBytes {
		writeAuthProblem(w, http.StatusRequestEntityTooLarge, "payload_too_large",
			"Request-Body überschreitet das 4 KiB-Limit für Auth-Endpunkte.")
		return nil, errors.New("payload too large")
	}
	return raw, nil
}

// writeAuthError mappt Auth-Domainfehler auf die Codes aus
// §3.9. Reihenfolge ist die neunstufige Fehlerpräzedenz aus dem
// Wire-Vertrag — der Application-Service liefert genau einen Fehler
// pro Aufruf, hier wird er auf das HTTP-Schema gemappt.
func writeAuthError(w http.ResponseWriter, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrAuthTokenMissing):
		writeAuthProblem(w, http.StatusUnauthorized, "auth_token_missing", "Pflicht-Auth-Header fehlt.")
	case errors.Is(err, domain.ErrAuthTokenInvalid):
		writeAuthProblem(w, http.StatusUnauthorized, "auth_token_invalid", "Auth-Token ungültig.")
	case errors.Is(err, domain.ErrAuthTokenRevoked):
		writeAuthProblem(w, http.StatusUnauthorized, "auth_token_revoked", "Auth-Token widerrufen.")
	case errors.Is(err, domain.ErrAuthTokenExpired):
		writeAuthProblem(w, http.StatusUnauthorized, "auth_token_expired", "Auth-Token abgelaufen.")
	case errors.Is(err, domain.ErrAuthTokenNotYetValid):
		writeAuthProblem(w, http.StatusUnauthorized, "auth_token_not_yet_valid", "Auth-Token noch nicht gültig.")
	case errors.Is(err, domain.ErrAuthProjectMismatch):
		writeAuthProblem(w, http.StatusUnauthorized, "auth_project_mismatch", "Token-`project_id` widerspricht der Anfrage.")
	case errors.Is(err, domain.ErrAuthSessionScopeDenied):
		writeAuthProblem(w, http.StatusForbidden, "auth_session_scope_denied", "Audience/Session-Bindung nicht erlaubt.")
	case errors.Is(err, domain.ErrAuthPolicyDenied):
		writeAuthProblem(w, http.StatusForbidden, "auth_policy_denied", "Project-Policy lehnt den Request ab.")
	case errors.Is(err, domain.ErrAuthTokenTTLTooLarge):
		writeAuthProblem(w, http.StatusUnprocessableEntity, "auth_token_ttl_too_large",
			"`ttl_seconds` ist <= 0 oder überschreitet die wirksame Project-Grenze (max. 900 s).")
	case errors.Is(err, domain.ErrAuthIssuanceRateLimited):
		writeAuthProblem(w, http.StatusTooManyRequests, "auth_issuance_rate_limited",
			"Issuance-Quote überschritten.")
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		writeAuthProblem(w, http.StatusServiceUnavailable, "service_unavailable", "Anfrage wurde abgebrochen.")
	default:
		if logger != nil {
			logger.Warn("auth issuance handler error", "err", err.Error())
		}
		writeAuthProblem(w, http.StatusInternalServerError, "internal_error", "Issuance fehlgeschlagen.")
	}
}

// writeAuthProblem nutzt das in §3.9 implizierte JSON-Body-Schema
// (`status`, `code`, `message`) — selbe Form wie die anderen
// Adapter, damit der Client nur einen Error-Reader braucht.
func writeAuthProblem(w http.ResponseWriter, status int, code, message string) {
	body := map[string]any{
		"status":  "error",
		"code":    code,
		"message": message,
	}
	writeJSON(w, status, body)
}
