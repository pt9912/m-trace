package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// BrowserIngestPolicyLookup liest die `domain.BrowserIngestPolicy`
// für eine `project_id`. Wird vom Browser-Ingest-Enforcement-Pfad
// genutzt, um Origin-Pin und CSRF-Pflicht pro Project aufzulösen
// (`0.12.5`/RAK-80). Der `auth.InMemoryProjectPolicyResolver` erfüllt
// das Interface natürlicher Weise — er hat schon eine
// `ResolvePolicy`-Methode auf dem `ProjectPolicyResolver`-Port.
type BrowserIngestPolicyLookup interface {
	ResolvePolicy(ctx context.Context, projectID string) (domain.ProjectPolicy, error)
}

// BrowserIngestEnforcementConfig bündelt die Dependencies für die
// Browser-Ingest-POST-Middleware aus (RAK-80).
//
// Verhalten pro Request:
//  - `Origin`-Header fehlt → Operator-/CLI-Pfad. Bei aktivierter
//  `BrowserIngestPolicy` und gesetztem `OriginPin` wird der
//  Request mit `403 ingest_browser_origin_pin_mismatch` abgelehnt,
//  weil ein Pin nur Sinn macht, wenn der Origin auch gemeldet
//  wird. Sonst (kein Pin) gilt der heutige Operator-Pfad
//  unverändert.
//  - `X-MTrace-Token` fehlt → keine Project-Identifikation möglich
//  → Middleware tut nichts; der bestehende Handler liefert sein
//  `auth_token_missing`/-`invalid`-Verhalten.
//  - Token resolved + `BrowserIngestPolicy.Enabled=false` → Pfad
//  wie heute (RAK-74-Scope-Cut bleibt strikt).
//  - Token resolved + `Enabled=true`:
//  1. Origin **muss** in `CORSAllowlist` stehen — sonst `403
//  ingest_browser_origin_not_allowed`.
//  2. Falls `OriginPin != ""` → Origin muss exakt dem Pin
//  entsprechen — sonst `403 ingest_browser_origin_pin_mismatch`.
//  3. Falls `CSRFRequired` → `X-MTrace-CSRF`-Header muss
//  nicht-leer sein. Der konkrete Token-Vergleich ist Folge-
//  Item (Production-Anti-CSRF-Bibliothek); das Skelett
//  verhindert mindestens fehlende CSRF-Header — `403
//  ingest_browser_csrf_missing`.
type BrowserIngestEnforcementConfig struct {
	Projects driven.ProjectResolver
	Policies BrowserIngestPolicyLookup
	Logger   *slog.Logger
}

// browserIngestEnforcement gibt eine Middleware zurück, die die
// Browser-Ingest-Constraints aus `BrowserIngestPolicy` auf
// `/api/ingest/*`-POSTs erzwingt. Wenn `Projects` oder `Policies`
// nil sind, wird die Middleware zum No-Op — der heutige Pfad bleibt
// erhalten (Default RAK-74-Scope-Cut).
func browserIngestEnforcement(cfg BrowserIngestEnforcementConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !shouldEnforceBrowserIngest(cfg, r) {
				next.ServeHTTP(w, r)
				return
			}
			policy, ok := resolveActiveBrowserPolicy(cfg, r)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			if code, msg := checkBrowserIngestPolicy(policy.BrowserIngest, r); code != "" {
				writeBrowserIngestForbidden(w, code, msg)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// shouldEnforceBrowserIngest filtert die billigsten Skip-Pfade:
// Middleware deaktiviert (cfg nil), kein Project-Token-Header (also
// Operator-/CLI-Pfad oder ein bereits ungültiger Header — der
// Handler liefert das passende Mapping selbst).
func shouldEnforceBrowserIngest(cfg BrowserIngestEnforcementConfig, r *http.Request) bool {
	if cfg.Projects == nil || cfg.Policies == nil {
		return false
	}
	return strings.TrimSpace(r.Header.Get("X-MTrace-Token")) != ""
}

// resolveActiveBrowserPolicy resolved das Project aus dem Token und
// holt die Policy. Bei ungültigem Token oder fehlender/inaktiver
// Browser-Ingest-Policy liefert die Funktion `ok=false` und der
// Caller routet den Request unverändert weiter.
func resolveActiveBrowserPolicy(cfg BrowserIngestEnforcementConfig, r *http.Request) (domain.ProjectPolicy, bool) {
	token := strings.TrimSpace(r.Header.Get("X-MTrace-Token"))
	project, err := cfg.Projects.ResolveByToken(r.Context(), token)
	if err != nil {
		return domain.ProjectPolicy{}, false
	}
	policy, err := cfg.Policies.ResolvePolicy(r.Context(), project.ID)
	if err != nil || policy.BrowserIngest.IsZero() || !policy.BrowserIngest.Enabled {
		return domain.ProjectPolicy{}, false
	}
	return policy, true
}

// checkBrowserIngestPolicy liefert leere Strings, wenn der Request
// die aktivierte `BrowserIngestPolicy` erfüllt. Sonst liefert die
// Funktion den Wire-Fehler-Code und eine operator-lesbare Message.
func checkBrowserIngestPolicy(b domain.BrowserIngestPolicy, r *http.Request) (code, message string) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		if b.OriginPin == "" {
			return "", ""
		}
		return "ingest_browser_origin_pin_mismatch",
			"browser-ingest policy requires a pinned Origin header for this project"
	}
	if !b.AllowsBrowserOrigin(origin) {
		return "ingest_browser_origin_not_allowed",
			"Origin is not in the project browser-ingest allowlist"
	}
	if !b.MatchesOriginPin(origin) {
		return "ingest_browser_origin_pin_mismatch",
			"Origin does not match the configured browser-ingest pin"
	}
	if b.CSRFRequired && strings.TrimSpace(r.Header.Get("X-MTrace-CSRF")) == "" {
		return "ingest_browser_csrf_missing",
			"X-MTrace-CSRF header is required when browser-ingest policy demands CSRF"
	}
	return "", ""
}

// writeBrowserIngestForbidden schreibt ein `application/json`-Body mit
// `code`/`message`-Feldern für `403`-Antworten der Browser-Ingest-
// Enforcement, analog zu den anderen API-Fehler-Codes (`auth_*`
// etc.). Status ist immer `403` — die Funktion ist deshalb status-
// frei (unparam-Lint).
func writeBrowserIngestForbidden(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusForbidden)
	body := `{"code":"` + code + `","message":"` + message + `"}`
	_, _ = w.Write([]byte(body))
}
