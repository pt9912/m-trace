package auth

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// ErrNoSecretConfigured signalisiert, dass das Backend zwar verfügbar
// ist, aber keine Signing-Keys konfiguriert sind — der `EnvSecretBackend`
// nutzt das, um den Caller (`main.go`) das Lab-Default-Fallback-
// Verhalten entscheiden zu lassen (ENV-Adapter mit
// `MTRACE_AUTH_LAB_DEFAULT=1` springt ein, alle anderen Backends
// failen).
//
// Wichtig: dieser Fehler ist explizit ein „kein Material gefunden"-
// Signal, nicht ein „Backend nicht erreichbar"-Signal. Backends mit
// I/O-Fehler liefern dedizierte Fehler (z. B. Vault-Outage).
var ErrNoSecretConfigured = errors.New("auth: no signing key material configured")

// EnvSecretBackend implementiert `driven.AuthSecretBackend` über die
// ENV-Variablen `MTRACE_AUTH_SIGNING_KEYS`/`_ACTIVE_KID` mit Single-
// Key-Backwards-Compat zu `0.12.0` (`MTRACE_AUTH_SIGNING_KEY`/`_KID`).
// Default-Selektion in `main.go#buildAuthSecretBackend`.
//
// Lookup-Quelle ist `os.Getenv`; Tests können `LookupFn` injizieren,
// um deterministische Werte zu liefern, ohne den globalen Prozess-
// ENV-Zustand zu mutieren.
type EnvSecretBackend struct {
	LookupFn func(key string) string
	Now      func() time.Time
}

// NewEnvSecretBackend konstruiert den Default-Adapter mit
// `os.Getenv` und `time.Now`.
func NewEnvSecretBackend() *EnvSecretBackend {
	return &EnvSecretBackend{LookupFn: os.Getenv, Now: time.Now}
}

// Compile-time check.
var _ driven.AuthSecretBackend = (*EnvSecretBackend)(nil)

// LoadSigningKeys ruft `ParseSigningKeysEnv` mit den vier ENV-Werten
// und mappt den `noKeysConfigured`-Pfad auf den `ErrNoSecretConfigured`-
// Sentinel.
func (b *EnvSecretBackend) LoadSigningKeys(_ context.Context) ([]domain.SessionSigningKey, domain.SigningKeyID, error) {
	lookup := b.LookupFn
	if lookup == nil {
		lookup = os.Getenv
	}
	now := time.Now
	if b.Now != nil {
		now = b.Now
	}
	keys, activeKID, noKeys, err := ParseSigningKeysEnv(
		lookup("MTRACE_AUTH_SIGNING_KEYS"),
		lookup("MTRACE_AUTH_SIGNING_ACTIVE_KID"),
		lookup("MTRACE_AUTH_SIGNING_KEY"),
		lookup("MTRACE_AUTH_SIGNING_KID"),
		now().UTC(),
	)
	if err != nil {
		return nil, "", err
	}
	if noKeys {
		// Der hint-KID (z. B. authDefaultLabSigningKID) ist für den
		// Caller wichtig, falls er den Lab-Default-Pfad aktiviert.
		// Wir tunneln ihn über den activeKID-Rückgabewert, der bei
		// fehlendem Material sonst leer wäre.
		return nil, activeKID, ErrNoSecretConfigured
	}
	return keys, activeKID, nil
}
