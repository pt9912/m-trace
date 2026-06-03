package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// LabPassThroughKMSDecrypter ist ein Lab-Mock-Decrypter: er gibt
// den eingegebenen Ciphertext **unverändert** als Plaintext zurück.
// **NICHT FÜR PRODUKTION** — der Boot-Wiring aktiviert ihn nur unter
// `MTRACE_AUTH_KMS_LAB_MODE=1`. Erlaubt einen
// `make smoke-kms-skeleton`-Pfad ohne echtes AWS-KMS-Konto und
// einen unit-test-freien Boot-Smoke des KMS-Adapter-Pfads.
type LabPassThroughKMSDecrypter struct{}

// Decrypt gibt `ciphertext` 1:1 als `plaintext` zurück.
func (LabPassThroughKMSDecrypter) Decrypt(_ context.Context, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

// KMSDecrypter abstrahiert den AWS-KMS-`Decrypt`-Aufruf vom Adapter.
// Boot-Time-Wiring injiziert eine Production-Implementation (heute
// per Operator-Provider in `cmd/api/main.go`, Folge-Item nach
// `0.12.6` ist ein eingebauter AWS-SDK-v2-Adapter); Tests setzen einen
// In-Process-Stub, der den Mock-Plaintext liefert.
//
// Sicherheitsprofil:
//  - Plaintext (= das `keys`-Feldformat aus `ParseSigningKeysEnv`)
//  fließt nur durch den Decrypter zurück; nichts wird persistiert.
//  - Ciphertext kommt aus dem ENV `MTRACE_AUTH_KMS_ENCRYPTED_KEYS`
//  (Base64-URL-encoded KMS-Ciphertext-Blob). Bei Build-Hardening
//  kann er auch aus einem File-Path geliefert werden — siehe
//  `MTRACE_AUTH_KMS_ENCRYPTED_KEYS_PATH`.
//  - Der Adapter macht **kein** Re-Decrypt pro Request — der Plaintext
//  wird beim Boot in `LoadSigningKeys` eingelesen und an den
//  `MultiKeySigningResolver` übergeben. Keys leben dann nur im
//  Prozess-RAM des API-Containers.
type KMSDecrypter interface {
	Decrypt(ctx context.Context, ciphertext []byte) (plaintext []byte, err error)
}

// KMSSecretBackend implementiert `driven.AuthSecretBackend` über
// AWS-KMS (oder einen kompatiblen `Decrypt`-Pfad — der Adapter ist
// vendor-neutral durch die `KMSDecrypter`-Indirektion).
//
// **Adapter-Charakter** (`0.12.6` T8): produktive AWS-SDK-Anbindung
// ist NICHT Teil des Scopes. Operatoren, die KMS bereits
// einsetzen, können den Adapter durch einen eigenen Decrypter
// hinter dem `KMSDecrypter`-Interface aktivieren. Skelett-Adapter
// im Repo nutzt die `KMSDecrypter`-Abstraktion + Mock-Decrypter
// in Tests; ein produktiver Smoke gegen ein echtes KMS-Konto ist
// Folge-Item.
type KMSSecretBackend struct {
	Decrypter  KMSDecrypter
	Ciphertext []byte
	ActiveKID  domain.SigningKeyID
	Now        func() time.Time
}

// KMSBackendConfig fasst die Operator-konfigurierten Felder zusammen.
type KMSBackendConfig struct {
	Decrypter  KMSDecrypter
	Ciphertext []byte
	ActiveKID  domain.SigningKeyID
}

// NewKMSSecretBackend baut den Adapter aus den `MTRACE_AUTH_KMS_*`-
// ENV-Variablen plus einem injizierten `Decrypter`. Heute liest der
// Konstruktor:
//
//  - `MTRACE_AUTH_KMS_ENCRYPTED_KEYS` (Base64-URL-Ciphertext) ODER
//  `MTRACE_AUTH_KMS_ENCRYPTED_KEYS_PATH` (File mit Ciphertext).
//  - `MTRACE_AUTH_KMS_ACTIVE_KID` — die aktive KID, die nach Decrypt
//  in der gleichen Form wie bei `MTRACE_AUTH_SIGNING_ACTIVE_KID`
//  ausgewertet wird.
//
// Der Constructor failt fail-closed, wenn weder Ciphertext-ENV noch
// Path-ENV gesetzt sind, oder wenn `ActiveKID` fehlt.
func NewKMSSecretBackend(lookup func(key string) string, decrypter KMSDecrypter) (*KMSSecretBackend, error) {
	if lookup == nil {
		lookup = os.Getenv
	}
	if decrypter == nil {
		return nil, errors.New("auth kms backend: KMSDecrypter is nil (operator must wire AWS-SDK or compatible adapter)")
	}
	activeRaw := strings.TrimSpace(lookup("MTRACE_AUTH_KMS_ACTIVE_KID"))
	if activeRaw == "" {
		return nil, errors.New("auth kms backend: MTRACE_AUTH_KMS_ACTIVE_KID is required")
	}
	ciphertext, err := loadKMSCiphertext(lookup)
	if err != nil {
		return nil, err
	}
	return &KMSSecretBackend{
		Decrypter:  decrypter,
		Ciphertext: ciphertext,
		ActiveKID:  domain.SigningKeyID(activeRaw),
		Now:        time.Now,
	}, nil
}

// loadKMSCiphertext liest den KMS-Ciphertext entweder aus
// `MTRACE_AUTH_KMS_ENCRYPTED_KEYS` (Base64-URL) oder
// `MTRACE_AUTH_KMS_ENCRYPTED_KEYS_PATH` (File-Path). Beide sind
// alternativ zueinander; ist keiner gesetzt, fails der Constructor.
func loadKMSCiphertext(lookup func(key string) string) ([]byte, error) {
	rawB64 := strings.TrimSpace(lookup("MTRACE_AUTH_KMS_ENCRYPTED_KEYS"))
	rawPath := strings.TrimSpace(lookup("MTRACE_AUTH_KMS_ENCRYPTED_KEYS_PATH"))
	if rawB64 == "" && rawPath == "" {
		return nil, errors.New("auth kms backend: either MTRACE_AUTH_KMS_ENCRYPTED_KEYS or _ENCRYPTED_KEYS_PATH must be set")
	}
	if rawB64 != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(rawB64)
		if err != nil {
			// `RawURLEncoding` lehnt Padding ab; fallback auf StdEncoding.
			decoded, err = base64.StdEncoding.DecodeString(rawB64)
			if err != nil {
				return nil, fmt.Errorf("auth kms backend: MTRACE_AUTH_KMS_ENCRYPTED_KEYS not valid base64: %w", err)
			}
		}
		return decoded, nil
	}
	data, err := os.ReadFile(rawPath)
	if err != nil {
		return nil, fmt.Errorf("auth kms backend: read MTRACE_AUTH_KMS_ENCRYPTED_KEYS_PATH %q: %w", rawPath, err)
	}
	return data, nil
}

// Compile-time check.
var _ driven.AuthSecretBackend = (*KMSSecretBackend)(nil)

// LoadSigningKeys decryptet den Ciphertext via `Decrypter` und reicht
// den Plaintext (im `ParseSigningKeysEnv`-Format) an den gemeinsamen
// Validator weiter.
func (b *KMSSecretBackend) LoadSigningKeys(ctx context.Context) ([]domain.SessionSigningKey, domain.SigningKeyID, error) {
	if b == nil {
		return nil, "", errors.New("auth kms backend: nil receiver")
	}
	if b.Decrypter == nil {
		return nil, "", errors.New("auth kms backend: nil decrypter")
	}
	plaintext, err := b.Decrypter.Decrypt(ctx, b.Ciphertext)
	if err != nil {
		return nil, "", fmt.Errorf("auth kms backend: decrypt: %w", err)
	}
	now := time.Now
	if b.Now != nil {
		now = b.Now
	}
	keysRaw := strings.TrimSpace(string(plaintext))
	if keysRaw == "" {
		return nil, "", errors.New("auth kms backend: decrypted plaintext is empty")
	}
	keys, activeKID, noKeys, parseErr := ParseSigningKeysEnv(keysRaw, string(b.ActiveKID), "", "", now().UTC())
	if parseErr != nil {
		return nil, "", fmt.Errorf("auth kms backend: parse: %w", parseErr)
	}
	if noKeys {
		return nil, "", errors.New("auth kms backend: decrypted plaintext yielded no signing keys")
	}
	return keys, activeKID, nil
}
