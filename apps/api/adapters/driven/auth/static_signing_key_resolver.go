package auth

import (
	"errors"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// StaticSigningKeyResolver implementiert `driven.SigningKeyResolver`
// gegen eine vom Operator konfigurierte Key-Liste (`0.12.0`,
// RAK-72). Restart-Stabilität wird durch den Operator-Workflow
// garantiert: dieselbe Konfiguration produziert deterministisch
// dieselben Keys.
//
// `activeKID` markiert den aktuell aktiven Signing-Key; alle weiteren
// Keys bleiben ausschließlich Verify-Keys, bis sie ebenfalls aktiviert
// werden.
//
// Sicherheitsprofil:
//   - `Secret`-Material wird beim Loader-Aufruf in den Resolver kopiert
//     und nie aus diesem Adapter heraus geleakt — nur die Domain-Form
//     `SessionSigningKey` wird zurückgegeben.
//   - Tests dürfen mit kurzem Material arbeiten; produktive
//     Konfiguration muss mindestens 256 Bit Entropie pro Key
//     bereitstellen (Wartung in Operator-Doku).
type StaticSigningKeyResolver struct {
	keys      map[domain.SigningKeyID]domain.SessionSigningKey
	order     []domain.SigningKeyID
	activeKID domain.SigningKeyID
}

// NewStaticSigningKeyResolver baut den Resolver. `keys` muss mindestens
// einen Eintrag für `activeKID` haben; andere Keys bleiben Verify-Keys.
// Doppelte `KID`-Werte führen zu einem deterministischen Fehler beim
// Loader, damit Operator-Konfigurationsfehler früh sichtbar werden.
func NewStaticSigningKeyResolver(activeKID domain.SigningKeyID, keys ...domain.SessionSigningKey) (*StaticSigningKeyResolver, error) {
	if activeKID == "" {
		return nil, errors.New("auth: active signing key id must not be empty")
	}
	out := &StaticSigningKeyResolver{
		keys:      make(map[domain.SigningKeyID]domain.SessionSigningKey, len(keys)),
		order:     make([]domain.SigningKeyID, 0, len(keys)),
		activeKID: activeKID,
	}
	for _, k := range keys {
		if k.KID == "" {
			return nil, errors.New("auth: signing key id must not be empty")
		}
		if _, dup := out.keys[k.KID]; dup {
			return nil, errors.New("auth: duplicate signing key id " + string(k.KID))
		}
		// Defensiv kopieren, damit Caller das Material nicht
		// nachträglich mutieren kann.
		secretCopy := append([]byte(nil), k.Secret...)
		copyKey := k
		copyKey.Secret = secretCopy
		out.keys[k.KID] = copyKey
		out.order = append(out.order, k.KID)
	}
	if _, ok := out.keys[activeKID]; !ok {
		return nil, errors.New("auth: active signing key " + string(activeKID) + " not present in ring")
	}
	return out, nil
}

// Compile-time check.
var _ driven.SigningKeyResolver = (*StaticSigningKeyResolver)(nil)

// ActiveSigningKey gibt den aktuell aktiven Signing-Key zurück.
func (r *StaticSigningKeyResolver) ActiveSigningKey() (domain.SessionSigningKey, error) {
	if r == nil {
		return domain.SessionSigningKey{}, domain.ErrAuthTokenInvalid
	}
	k, ok := r.keys[r.activeKID]
	if !ok {
		return domain.SessionSigningKey{}, domain.ErrAuthTokenInvalid
	}
	return k, nil
}

// AllVerifyKeys gibt alle geladenen Keys in stabiler Reihenfolge
// zurück. Verify nutzt das, um nach `kid` aufzulösen — Reihenfolge
// matters nicht funktional, ist aber stabil für Tests.
func (r *StaticSigningKeyResolver) AllVerifyKeys() ([]domain.SessionSigningKey, error) {
	if r == nil {
		return nil, domain.ErrAuthTokenInvalid
	}
	out := make([]domain.SessionSigningKey, 0, len(r.order))
	for _, kid := range r.order {
		out = append(out, r.keys[kid])
	}
	return out, nil
}
