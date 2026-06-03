package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"
)

// Project-Token-Generationen-Domäne (RAK-73).
//
// Das Modell pinnt:
//  - Klartext-Token werden nicht persistiert; nur Hash + Fingerprint
//  plus Lifecycle-Felder (`not_before`, `grace_until`, `expires_at`,
//  `revoked_at`) leben in `ProjectTokenGeneration`.
//  - `grace_until` ist die Restart-stabile Source of Truth für die
//  Grace-Validierung; wird nicht aus Prozesszustand oder
//  `RotatedFrom` rekonstruiert.
//  - `revoked_at` beendet Grace sofort.
//  - Status wird pro Validierungsaufruf aus den Lifecycle-Feldern
//  plus aktueller Clock evaluiert; es gibt kein gespeichertes
//  Status-Feld, das mit den Zeit-Feldern in Konflikt geraten
//  könnte.
//  - Hash-Vergleich läuft in konstanter Zeit über `crypto/subtle`.
//
// Wire-Prefix `mtr_pt_` (Project Token) ist bewusst unterschiedlich
// zu `mtr_st_` (Session Token, §3.9) und `mtr_ing_` (Stream Key,
// §3.8) — Prefix-Verwechslungen schlagen direkt im Adapter durch.

// projectTokenPrefix ist der projektweite Marker für rotierbare
// Project Tokens. Der Prefix ist Teil des Klartexts und damit auch im
// Hash; ein Cross-Token-Type-Bypass mit fremdem Token-Format wird
// dadurch zusätzlich erschwert.
const projectTokenPrefix = "mtr_pt_"

const projectTokenEntropyBytes = 32 // 256 Bit

// ProjectTokenGenerationStatus klassifiziert den Lifecycle-Zustand
// einer einzelnen Generation (RAK-73, `spec/backend-api-contract.md`
// §3.9). Der Wert wird **nicht** persistiert — er wird pro
// Validierungsaufruf aus den Lifecycle-Feldern plus aktueller Clock
// berechnet, damit kein Drift zwischen Persistenz und Zeitvergleich
// entstehen kann.
type ProjectTokenGenerationStatus string

// ProjectTokenGenerationStatus-Werte aus dem Muss-Scope.
const (
	ProjectTokenStatusActive     ProjectTokenGenerationStatus = "active"
	ProjectTokenStatusGrace      ProjectTokenGenerationStatus = "grace"
	ProjectTokenStatusNotYetValid ProjectTokenGenerationStatus = "not_yet_valid"
	ProjectTokenStatusExpired    ProjectTokenGenerationStatus = "expired"
	ProjectTokenStatusRevoked    ProjectTokenGenerationStatus = "revoked"
)

// CanAuthenticate gibt true zurück, wenn eine Generation in diesem
// Status einen Request authentifizieren darf. nutzt das im
// Multi-Generation-Lookup, um Active- und Grace-Tokens als gültig
// und alle anderen als ungültig zu klassifizieren.
func (s ProjectTokenGenerationStatus) CanAuthenticate() bool {
	switch s {
	case ProjectTokenStatusActive, ProjectTokenStatusGrace:
		return true
	default:
		return false
	}
}

// ProjectTokenGeneration ist die persistierbare Sicht auf eine
// einzelne Token-Generation. **Klartext-Felder gibt es hier nicht** —
// nur Hash, Fingerprint und Lifecycle-Metadaten.
//
// `RotatedFrom` zeigt optional auf die Vorgänger-Generation. Der Wert
// ist rein dokumentarisch — die Grace-Entscheidung läuft
// ausschließlich über `GraceUntil`, **nicht** über
// `RotatedFrom`-Lookups (Plan §0.6 Threat Model).
type ProjectTokenGeneration struct {
	TokenID     string
	ProjectID   string
	KeyHash     string
	Fingerprint string
	NotBefore   time.Time
	GraceUntil  *time.Time
	ExpiresAt   *time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time
	RotatedFrom *string
}

// ProjectTokenMaterial bündelt die Werte einer
// `GenerateProjectToken`-Antwort. **`Value` ist der Klartext** und
// darf nur in der Issuance-/Rotate-Wire-Antwort erscheinen — nicht in
// Persistenz, Logs, Metriken, Traces oder Fixtures.
type ProjectTokenMaterial struct {
	Value      string
	Generation ProjectTokenGeneration
}

// GenerateProjectToken erzeugt eine neue Token-Generation mit
// CSPRNG-Klartext, SHA-256-Hash und redigiertem Fingerprint. Der
// Aufrufer bekommt die persistente Sicht über
// `Material.Generation`; das Klartext-`Material.Value` darf
// ausschließlich an die Wire-Antwort der Issuance-/Rotate-Endpoints
// weitergereicht werden.
//
// `tokenID`, `projectID`, `notBefore`, `graceUntil`, `expiresAt` und
// `rotatedFrom` werden vom Aufrufer gesetzt — typischerweise im
// Application-Service auf Basis der Repository-State (z. B. ULID,
// `now`, optionaler Grace-Dauer aus Project Policy).
func GenerateProjectToken(
	tokenID string,
	projectID string,
	notBefore time.Time,
	graceUntil *time.Time,
	expiresAt *time.Time,
	rotatedFrom *string,
	createdAt time.Time,
) (ProjectTokenMaterial, error) {
	raw := make([]byte, projectTokenEntropyBytes)
	if _, err := rand.Read(raw); err != nil {
		return ProjectTokenMaterial{}, err
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	value := projectTokenPrefix + encoded
	gen := ProjectTokenGeneration{
		TokenID:     tokenID,
		ProjectID:   projectID,
		KeyHash:     hashProjectToken(value),
		Fingerprint: fingerprintProjectToken(value),
		NotBefore:   notBefore,
		GraceUntil:  cloneTimePointer(graceUntil),
		ExpiresAt:   cloneTimePointer(expiresAt),
		RevokedAt:   nil,
		CreatedAt:   createdAt,
		RotatedFrom: cloneStringPointer(rotatedFrom),
	}
	return ProjectTokenMaterial{Value: value, Generation: gen}, nil
}

// EvaluateProjectTokenStatus berechnet den effektiven Lifecycle-
// Status einer Generation gegen die übergebene Clock. Reihenfolge
// (höchste Priorität zuerst):
//
//  1. `RevokedAt` gesetzt und `now >= RevokedAt` → revoked. `RevokedAt`
//  beendet Grace sofort, auch wenn `GraceUntil` noch in der
//  Zukunft läge.
//  2. `ExpiresAt` gesetzt und `now >= ExpiresAt` → expired.
//  3. `now < NotBefore` → not_yet_valid.
//  4. `GraceUntil` gesetzt und `now >= GraceUntil`-Beginn-Indikator:
//  Grace ist semantisch das Fenster zwischen `now < GraceUntil`
//  einer alten Generation, in dem sie noch authentifizieren darf.
//  Wir behandeln eine Generation mit gesetztem `GraceUntil` als
//  `grace`, solange `now < GraceUntil` und keine der oberen
//  Bedingungen greift; nach Ablauf von `GraceUntil` ohne
//  `ExpiresAt` läuft sie auf `expired`.
//  5. Sonst → active.
//
// Diese Reihenfolge spiegelt §3.9 (Fehlerpräzedenz revoked → expired
// → not_yet_valid) und Plan §0.6 (Grace ist persistiert, `RevokedAt`
// beendet Grace sofort) deterministisch wider.
func EvaluateProjectTokenStatus(g ProjectTokenGeneration, now time.Time) ProjectTokenGenerationStatus {
	if g.RevokedAt != nil && !now.Before(*g.RevokedAt) {
		return ProjectTokenStatusRevoked
	}
	if g.ExpiresAt != nil && !now.Before(*g.ExpiresAt) {
		return ProjectTokenStatusExpired
	}
	if now.Before(g.NotBefore) {
		return ProjectTokenStatusNotYetValid
	}
	if g.GraceUntil != nil {
		if now.Before(*g.GraceUntil) {
			return ProjectTokenStatusGrace
		}
		// `GraceUntil` gesetzt aber abgelaufen, ohne separates
		// `ExpiresAt`: die Generation ist nach dem Grace-Fenster
		// effektiv abgelaufen.
		if g.ExpiresAt == nil {
			return ProjectTokenStatusExpired
		}
	}
	return ProjectTokenStatusActive
}

// StatusToAuthError mapt einen evaluierten Generation-Status auf den
// passenden Auth-Domainfehler aus `errors.go`. `active` und `grace`
// liefern `nil`, weil sie authentifizieren dürfen. Die Funktion ist
// die zentrale Quelle für die Status→Fehler-Zuordnung; HTTP-Adapter
// und Application-Service nutzen sie statt Inline-Switch-Statements.
func StatusToAuthError(status ProjectTokenGenerationStatus) error {
	switch status {
	case ProjectTokenStatusActive, ProjectTokenStatusGrace:
		return nil
	case ProjectTokenStatusRevoked:
		return ErrAuthTokenRevoked
	case ProjectTokenStatusExpired:
		return ErrAuthTokenExpired
	case ProjectTokenStatusNotYetValid:
		return ErrAuthTokenNotYetValid
	default:
		return ErrAuthTokenInvalid
	}
}

// ValidateProjectTokenString prüft Format und Hash eines vom Aufrufer
// präsentierten Klartext-Project-Tokens gegen eine Liste von
// Generationen. Liefert die passende Generation und ihren effektiven
// Status zurück, oder einen stabilen Auth-Fehler.
//
// Reihenfolge:
//  1. Format/Prefix → `ErrAuthTokenInvalid`.
//  2. Hash-Vergleich in konstanter Zeit gegen alle Generationen des
//  Aufrufer-Scopes. Bei Treffer wird die Generation übernommen.
//  3. `EvaluateProjectTokenStatus` plus `StatusToAuthError` mappen
//  den Status auf einen Fehler oder `nil`.
//
// Es gibt **keinen** Cross-Project-Lookup: der Aufrufer übergibt
// ausschließlich Generationen des relevanten Scopes (typischerweise
// alle Generationen pro `project_id` aus dem Repository-Lookup oder
// alle Generationen über die Map des Static-Resolvers).
//
// Bei keinem Hash-Treffer liefert die Funktion `ErrAuthTokenInvalid`
// — der HTTP-Adapter mappt das auf `401 auth_token_invalid`. So
// leakt die Funktion nicht, ob der Token zu einem fremden Project
// gehört oder nur am Hash-Vergleich gescheitert ist.
func ValidateProjectTokenString(provided string, generations []ProjectTokenGeneration, now time.Time) (ProjectTokenGeneration, ProjectTokenGenerationStatus, error) {
	if !strings.HasPrefix(provided, projectTokenPrefix) {
		return ProjectTokenGeneration{}, "", ErrAuthTokenInvalid
	}
	candidate := hashProjectToken(provided)
	candidateBytes := []byte(candidate)
	for _, g := range generations {
		if subtle.ConstantTimeCompare(candidateBytes, []byte(g.KeyHash)) == 1 {
			status := EvaluateProjectTokenStatus(g, now)
			if err := StatusToAuthError(status); err != nil {
				return g, status, err
			}
			return g, status, nil
		}
	}
	return ProjectTokenGeneration{}, "", ErrAuthTokenInvalid
}

// HasProjectTokenPrefix gibt true zurück, wenn `token` mit dem
// `mtr_pt_`-Wire-Prefix beginnt. Adapter nutzen das, um neue
// rotierbare Project Tokens vom Legacy-Statisch-Token-Pfad
// (`demo-token` ohne Prefix) zu unterscheiden, ohne den Klartext
// auszuwerten.
func HasProjectTokenPrefix(token string) bool {
	return strings.HasPrefix(token, projectTokenPrefix)
}

// FingerprintProjectTokenValue baut die redigierte Audit-Form aus
// einem Klartext-Token. Der Adapter nutzt das, wenn er einen frisch
// präsentierten Token kurzzeitig im Speicher hält und einen
// Log-/Audit-Identifier braucht, ohne den Hash mit zu schreiben.
// Persistenz nimmt ausschließlich `ProjectTokenGeneration.Fingerprint`,
// nie ein on-the-fly aus dem Klartext gerechnetes Ergebnis.
func FingerprintProjectTokenValue(value string) string {
	return fingerprintProjectToken(value)
}

// hashProjectToken berechnet den SHA-256-Hex-Digest des vollständigen
// Klartext-Tokens. Identisches Verfahren wie `stream_key.go`, weil
// der Klartext bereits 256 Bit Entropie trägt — kein Password-Hash
// nötig.
func hashProjectToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// fingerprintProjectToken baut die redigierte Audit-Form: erste 8
// Klartext-Zeichen (inklusive Prefix) + `...` + letzte 4 Zeichen.
// Bei sehr kurzen Werten fällt der Builder auf den Prefix plus `...`
// zurück, ohne den Klartext zu verraten.
func fingerprintProjectToken(value string) string {
	const headLen = 8
	const tailLen = 4
	if len(value) <= headLen+tailLen {
		return projectTokenPrefix + "..."
	}
	return value[:headLen] + "..." + value[len(value)-tailLen:]
}

// cloneTimePointer kopiert einen optionalen Time-Pointer. Wichtig,
// damit eine Generation, die in den Repository-Cache zurückgegeben
// wurde, nicht versehentlich denselben Pointer wie der Aufrufer
// teilt — ein nachträgliches Mutieren am Aufrufer-Side darf den
// gespeicherten Lifecycle-Wert nicht ändern.
func cloneTimePointer(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := *t
	return &v
}
