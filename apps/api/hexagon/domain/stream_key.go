package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// Stream-Key-Domäne (`0.11.0`, RAK-66).
//
// Sicherheitsprofil:
//   - CSPRNG: `crypto/rand.Read` mit 32 Byte (256 Bit Entropie).
//     Liefert genug Entropie für eine globale Eindeutigkeit pro
//     Project-Scope; eine geringe Kollisionswahrscheinlichkeit auf
//     Datenbankebene wird zusätzlich durch den Unique-Constraint auf
//     `key_hash` abgefangen (T2).
//   - Output-Format: URL-sicheres Base64 (`base64.RawURLEncoding`,
//     keine Padding-Zeichen) mit dem Prefix `mtr_ing_`. So sind
//     Keys in URLs, JSON-Bodies und CLI-Aufrufen kopierbar.
//   - Hash: SHA-256 über den Klartext, hex-kodiert. SHA-256 reicht
//     hier, weil der Klartext bereits 256 Bit Entropie trägt — ein
//     teurer Password-Hash (`scrypt`/`argon2`) wäre Overkill und
//     würde nur den Validate-Endpoint langsam machen, ohne
//     Sicherheitsgewinn (kein Wörterbuch-/Brute-Force-Risiko).
//   - Fingerprint: erste 8 + letzte 4 Klartext-Zeichen mit `...`
//     dazwischen, plus Prefix. Reicht zum Wiedererkennen, lässt
//     aber nicht den ganzen Key reconstruieren.
//   - Validate: konstantzeitvergleich auf den vollständigen Hash.
//     Der Fingerprint ist **nicht** verifier-tauglich.
//   - Klartext-Keys leben nur in `StreamKeyMaterial` und werden vom
//     Caller transient an die Create-/Rotate-Antwort weitergereicht.

// streamKeyPrefix ist der projektweite Marker für Ingest-Stream-
// Keys. Der Prefix ist Teil des Klartexts und damit auch im Hash;
// das ist beabsichtigt, weil ein Cross-Project-Bypass mit fremdem
// Key-Format dadurch zusätzlich erschwert wird.
const streamKeyPrefix = "mtr_ing_"

const streamKeyEntropyBytes = 32 // 256 Bit

// StreamKeyMaterial bündelt die Werte aus einem `GenerateStreamKey`-
// Aufruf. **`Value` ist der Klartext** und darf nicht persistiert,
// geloggt oder in Lifecycle-Events übernommen werden — er gehört
// ausschließlich in die Create-/Rotate-Wire-Antwort.
type StreamKeyMaterial struct {
	Value       string
	Hash        string
	Fingerprint string
	CreatedAt   time.Time
}

// StreamKey ist die persistente, log-/audit-/event-taugliche Sicht.
// Der Klartext-`Value` taucht hier bewusst nicht auf.
type StreamKey struct {
	Hash        string
	Fingerprint string
	CreatedAt   time.Time
}

// ToPersistable extrahiert die persistenz-taugliche Sicht aus dem
// Material. Das Klartext-`Value` wird im Aufrufer ausschließlich an
// die Create-/Rotate-Antwort weitergereicht; alles andere bekommt
// nur diese Form zu sehen.
func (m StreamKeyMaterial) ToPersistable() StreamKey {
	return StreamKey{
		Hash:        m.Hash,
		Fingerprint: m.Fingerprint,
		CreatedAt:   m.CreatedAt,
	}
}

// ErrStreamKeyMalformed signalisiert, dass ein vom Aufrufer
// übergebener Klartext-Key nicht das erwartete Format hat (Prefix
// fehlt, Base64-Decoding scheitert, Längenanforderung verfehlt).
// Der Validate-Endpoint mappt das auf `valid:false` ohne weitere
// Detailinformation, damit Cross-Project-Leak-Risiken minimiert
// bleiben.
var ErrStreamKeyMalformed = errors.New("stream key has invalid format")

// GenerateStreamKey erzeugt einen neuen Stream-Key. Der Aufrufer ist
// dafür verantwortlich, das Klartext-`Value` ausschließlich in die
// Create-/Rotate-Wire-Antwort zu reichen — alle anderen
// Persistenz-, Log- und Eventpfade bekommen nur die persistente
// Sicht über `ToPersistable`.
//
// `now` macht die Funktion testbar, ohne `time.Now()` zu mocken; in
// Produktion setzt der Aufrufer `time.Now().UTC()` ein.
func GenerateStreamKey(now time.Time) (StreamKeyMaterial, error) {
	raw := make([]byte, streamKeyEntropyBytes)
	if _, err := rand.Read(raw); err != nil {
		return StreamKeyMaterial{}, err
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	value := streamKeyPrefix + encoded
	return StreamKeyMaterial{
		Value:       value,
		Hash:        hashStreamKey(value),
		Fingerprint: fingerprintStreamKey(value),
		CreatedAt:   now,
	}, nil
}

// ValidateStreamKey vergleicht den vom Aufrufer übergebenen
// Klartext-Key in konstanter Zeit gegen einen vorhandenen
// `StreamKey.Hash`. Liefert `false` bei Format-Verstoß oder
// Hash-Mismatch; `error` ist nicht-nil nur bei Format-Verstoß
// (Aufrufer kann das wahlweise ignorieren — der Validate-Endpoint
// gibt `{ "valid": false }` ohne Detailcode zurück).
func ValidateStreamKey(provided string, persisted StreamKey) (bool, error) {
	if !strings.HasPrefix(provided, streamKeyPrefix) {
		return false, ErrStreamKeyMalformed
	}
	candidate := hashStreamKey(provided)
	if subtle.ConstantTimeCompare([]byte(candidate), []byte(persisted.Hash)) != 1 {
		return false, nil
	}
	return true, nil
}

// hashStreamKey berechnet den SHA-256-Hex-Digest des vollständigen
// Klartext-Keys. Wird von `GenerateStreamKey` und `ValidateStreamKey`
// gleich genutzt; getrennt von `fingerprintStreamKey`, damit der
// Fingerprint nicht versehentlich als verifier verwendet werden
// kann.
func hashStreamKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// fingerprintStreamKey baut die redigierte Audit-Form. Verwendet die
// ersten 8 und letzten 4 Zeichen des **Klartexts** (inklusive
// Prefix), damit das Format wiedererkennbar bleibt. Bei sehr kurzen
// Werten (theoretisch unmöglich, wenn `GenerateStreamKey` benutzt
// wurde) fällt der Builder auf den Prefix plus `...` zurück, ohne
// den Klartext zu verraten.
func fingerprintStreamKey(value string) string {
	const headLen = 8
	const tailLen = 4
	if len(value) <= headLen+tailLen {
		return streamKeyPrefix + "..."
	}
	return value[:headLen] + "..." + value[len(value)-tailLen:]
}
