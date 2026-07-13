// Package redisutil bündelt die zwischen den Redis-Limiter-Adaptern
// (auth: Issuance R-17 / Origin R-22; ratelimit: Ingest R-26) geteilten
// Low-Level-Helfer. Vor diesem Paket existierten die EVALSHA→EVAL-
// Fallback-Logik und die NOSCRIPT-Erkennung als handgepflegte Kopien
// pro Adapter — ein Drift (z. B. geänderter go-redis-Fehlertext) hätte
// die Kopien still auseinanderlaufen lassen.
package redisutil

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

// IsNoScriptError erkennt die Redis-Antwort auf ein EVALSHA, dessen
// Script nicht (mehr) im Script-Cache liegt (z. B. nach SCRIPT FLUSH
// oder Server-Neustart).
func IsNoScriptError(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT")
}

// Eval führt das Script bevorzugt über EVALSHA (sha) aus und fällt bei
// NOSCRIPT auf ein inline EVAL des Volltexts zurück. sha == "" (SCRIPT
// LOAD beim Konstruieren fehlgeschlagen) geht direkt auf EVAL.
func Eval(ctx context.Context, client redis.UniversalClient, sha, script string, keys []string, args []any) (any, error) {
	if sha != "" {
		res, err := client.EvalSha(ctx, sha, keys, args...).Result()
		if err == nil {
			return res, nil
		}
		if !IsNoScriptError(err) {
			return nil, err
		}
	}
	return client.Eval(ctx, script, keys, args...).Result()
}

// FailModeLabel benennt den Fail-Mode einheitlich für Logs über alle
// Redis-Limiter (Operatoren greppen nach genau diesen Wortlauten).
func FailModeLabel(failOpen bool) string {
	if failOpen {
		return "fail-open (local memory fallback)"
	}
	return "fail-closed (deny on outage)"
}
