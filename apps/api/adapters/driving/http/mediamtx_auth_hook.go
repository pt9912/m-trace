package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// MediaMTXAuthHookHandler bedient `POST /api/ingest/auth-hook` als
// Bridge zwischen MediaMTX-`externalAuth` und dem m-trace-Stream-
// Key-Validate-Pfad (`0.12.5`/RAK-81, R-14). Skelett-Adapter:
// MediaMTX schickt seine Form-Body-Felder, der Handler mappt auf
// die existierende `IngestControlInbound.ValidateKey`-Use-Case.
//
// MediaMTX-Vertrag (https://github.com/bluenviron/mediamtx → `externalAuth`):
//   - Methode: `POST`, `application/x-www-form-urlencoded`
//   - Felder: `user`, `password`, `action`, `path`, optional `ip`,
//     `protocol`, `id`, `query`
//   - Response: HTTP `200` = allow, alles andere = deny. Body wird
//     ignoriert.
//
// Mapping (Operator-Konvention, dokumentiert in `auth.md` §5.7):
//   - `user`     = Project-ID (m-trace).
//   - `password` = Stream-Key (Klartext, einmaliger Wert aus
//     Stream-Anlage; danach nur als `key_hash` persistiert).
//   - `path`     = Stream-ID (m-trace).
//   - `action`   = `publish` (erlaubt). `read`/`api`/`metrics`
//     liefert `403`; ein produktiver Read-/API-Auth-Pfad ist
//     Folge-Item nach `0.12.5`.
//
// Sicherheitsprofil:
//   - Der Endpoint ist eine Trust-Boundary MediaMTX→m-trace und
//     **hat selbst keine Project-Token-Auth**. Operator-Setup
//     muss ihn netzwerkseitig isolieren (Compose-internal-Netz,
//     K8s-`ClusterIP`, Reverse-Proxy mit IP-Allowlist). Doku
//     weist darauf hin.
//   - Kein Klartext-Material wird geloggt — nur Stream-ID und
//     Wire-Outcome (`allow`/`deny`).
//   - Idempotent gegenüber Replays: MediaMTX selbst sendet pro
//     Publish-Versuch einen Hook; die ValidateKey-Operation ist
//     side-effect-frei.
type MediaMTXAuthHookHandler struct {
	UseCase driving.IngestControlInbound
	Logger  *slog.Logger
}

// ServeHTTP implementiert die MediaMTX-`externalAuth`-Antwort.
// `Content-Type: application/x-www-form-urlencoded`-Pflicht; alles
// andere bekommt `400 invalid_content_type`.
func (h *MediaMTXAuthHookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ct := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	// MediaMTX schickt `application/x-www-form-urlencoded` ohne
	// Charset-Suffix; tolerieren wir aber.
	if ct != "application/x-www-form-urlencoded" &&
		!strings.HasPrefix(ct, "application/x-www-form-urlencoded;") {
		http.Error(w, "expected application/x-www-form-urlencoded", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form parse error", http.StatusBadRequest)
		return
	}
	user := strings.TrimSpace(r.PostFormValue("user"))
	pass := r.PostFormValue("password")
	action := strings.TrimSpace(r.PostFormValue("action"))
	path := strings.TrimSpace(r.PostFormValue("path"))

	// Nur `publish` ist erlaubt; Read-Auth bleibt Folge-Item.
	if action != "publish" {
		h.logDeny(r, user, path, "action_not_supported", action)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if user == "" || pass == "" || path == "" {
		h.logDeny(r, user, path, "missing_field", "")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	result, err := h.UseCase.ValidateKey(r.Context(), user, path, pass)
	if err != nil {
		h.logDeny(r, user, path, "validate_error", err.Error())
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !result.Valid {
		h.logDeny(r, user, path, "invalid_key", "")
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if h.Logger != nil {
		h.Logger.Info("auth-hook allowed",
			"project", user,
			"stream", path,
			"action", action,
		)
	}
	w.WriteHeader(http.StatusOK)
}

// logDeny zentralisiert den Audit-Log-Output für jeden Deny-Pfad.
// Wir loggen nur Project- und Stream-ID — niemals den Klartext-Key.
func (h *MediaMTXAuthHookHandler) logDeny(_ *http.Request, project, stream, reason, detail string) {
	if h.Logger == nil {
		return
	}
	args := []any{"project", project, "stream", stream, "reason", reason}
	if detail != "" {
		args = append(args, "detail", detail)
	}
	h.Logger.Info("auth-hook denied", args...)
}
