package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// HMACSessionTokenSigner implementiert `driven.SessionTokenSigner`
// mit HMAC-SHA-256-Signaturen (`alg=HS256`). Der Wire-Vertrag aus
// `spec/backend-api-contract.md` §3.9 spricht von Tokens der Form
// `mtr_st_<base64url(headerJSON)>.<base64url(claimsJSON)>.<base64url(sig)>`.
//
// Sicherheitsprofil:
//  - Signing benutzt ausschließlich den `Active`-Key.
//  - Verify schlägt jedem Key im Ring nach (aktiv + retired), damit
//  ein vor Key-Switch ausgestellter Token nach einem Rollover
//  weiterhin verifiziert.
//  - Unbekannter `kid` → `domain.ErrAuthTokenInvalid`.
//  - Unbekannter `alg` → `domain.ErrAuthTokenInvalid`. Wir
//  validieren `alg` strikt gegen die Domain-Allowlist
//  (`SigningKeyAlgorithmHS256`); ein „none"-Header ist damit
//  ausgeschlossen.
//  - Signature-Vergleich ist konstantzeitnah.
//
// Algorithmus-/`alg`-Wert ist Wire-konstant `HS256`; spätere
// Algorithmen werden additiv ergänzt. Header-/Claims-Encoding ist
// JSON ohne Whitespace, damit das Signaturmaterial deterministisch
// reproduzierbar ist (dieselbe Claim-Struktur produziert byte-genau
// dieselbe Signatur).
type HMACSessionTokenSigner struct {
	Active driven.SigningKeyResolver
}

// SigningKeyResolver ist die kleine Abstraktion über den Key-Ring
// (Definition im driven-Port-Paket, damit der Application-Service
// keinen direkten Zugriff auf das Schlüsselmaterial braucht). Der
// Adapter konsumiert den Resolver hier und in `Verify`.
//
// Der Resolver lebt im driven-Port-Paket; wir halten hier nur den
// Adapter-Code.

// NewHMACSessionTokenSigner konstruiert den Signer mit einem
// vorhandenen Key-Ring-Resolver. Der Resolver muss mindestens einen
// aktiven Key liefern; ohne aktiven Key panict `Sign` nicht, sondern
// liefert `domain.ErrAuthTokenInvalid` (wir behandeln den Fehlfall
// als „Auth-Pfad ist nicht konfiguriert").
func NewHMACSessionTokenSigner(active driven.SigningKeyResolver) *HMACSessionTokenSigner {
	return &HMACSessionTokenSigner{Active: active}
}

// Compile-time check.
var _ driven.SessionTokenSigner = (*HMACSessionTokenSigner)(nil)

// tokenHeader spiegelt den signierten JWS-Header.
type tokenHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

// claimSet ist die Wire-Form des Claim-Sets — JSON-feldname-stable,
// damit Sign und Verify byte-identische Strings produzieren.
type claimSet struct {
	Iss       string `json:"iss"`
	Sub       string `json:"sub"`
	Aud       string `json:"aud"`
	Iat       int64  `json:"iat"`
	Nbf       int64  `json:"nbf"`
	Exp       int64  `json:"exp"`
	Jti       string `json:"jti"`
	SessionID string `json:"session_id,omitempty"`
	Origin    string `json:"origin,omitempty"`
}

// Sign serialisiert die Claims, hängt den HS256-Header an und
// produziert die kompakte Wire-Form. Der Adapter wählt den aktuell
// aktiven Signing-Key; eine spätere Rotation ändert `Active.KID`,
// alte Verify-Keys bleiben über `Active.AllVerify` weiterhin
// auflösbar.
func (s *HMACSessionTokenSigner) Sign(claims domain.SessionTokenClaims) (string, error) {
	if s == nil || s.Active == nil {
		return "", domain.ErrAuthTokenInvalid
	}
	signingKey, err := s.Active.ActiveSigningKey()
	if err != nil {
		return "", err
	}
	if !signingKey.Algorithm.IsKnown() {
		return "", domain.ErrAuthTokenInvalid
	}
	hdr := tokenHeader{Alg: string(signingKey.Algorithm), Kid: string(signingKey.KID), Typ: "JWT"}
	hdrJSON, err := json.Marshal(hdr)
	if err != nil {
		return "", err
	}
	body := claimSet{
		Iss: claims.Iss,
		Sub: claims.Sub,
		Aud: string(claims.Aud),
		Iat: claims.Iat.Unix(),
		Nbf: claims.Nbf.Unix(),
		Exp: claims.Exp.Unix(),
		Jti: claims.JTI,
	}
	if claims.SessionID != nil {
		body.SessionID = *claims.SessionID
	}
	if claims.Origin != nil {
		body.Origin = *claims.Origin
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	signingInput := encode(hdrJSON) + "." + encode(bodyJSON)
	mac := hmac.New(sha256.New, signingKey.Secret)
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return "", err
	}
	sig := mac.Sum(nil)
	return domain.SessionTokenPrefix + signingInput + "." + encode(sig), nil
}

// Verify dekodiert das Wire-Format, prüft Header, sucht den Verify-Key
// im Ring und vergleicht die Signatur in konstanter Zeit. Liefert
// einen vollständigen `domain.SessionTokenClaims`. Zeit-, Audience-,
// Project- und Origin-Bindung prüft der Application-Service über die
// Domain-Funktionen aus `auth_session_token.go`.
func (s *HMACSessionTokenSigner) Verify(token string) (domain.SessionTokenClaims, error) {
	if s == nil || s.Active == nil {
		return domain.SessionTokenClaims{}, domain.ErrAuthTokenInvalid
	}
	parts, err := splitSessionToken(token)
	if err != nil {
		return domain.SessionTokenClaims{}, err
	}
	hdr, err := parseTokenHeader(parts[0])
	if err != nil {
		return domain.SessionTokenClaims{}, err
	}
	key, err := s.lookupKeyForHeader(hdr)
	if err != nil {
		return domain.SessionTokenClaims{}, err
	}
	if err := verifyHMACSignature(key.Secret, parts[0], parts[1], parts[2]); err != nil {
		return domain.SessionTokenClaims{}, err
	}
	return decodeClaimSet(parts[1])
}

// splitSessionToken validiert Prefix und Drei-Segment-Form und
// liefert die JOSE-typischen `[header, claims, sig]`-Strings zurück.
func splitSessionToken(token string) ([3]string, error) {
	var out [3]string
	if !domain.HasSessionTokenPrefix(token) {
		return out, domain.ErrAuthTokenInvalid
	}
	stripped := strings.TrimPrefix(token, domain.SessionTokenPrefix)
	parts := strings.Split(stripped, ".")
	if len(parts) != 3 {
		return out, domain.ErrAuthTokenInvalid
	}
	out[0], out[1], out[2] = parts[0], parts[1], parts[2]
	return out, nil
}

// parseTokenHeader dekodiert den Base64-Header und validiert
// `alg`-Allowlist sowie Pflicht-`kid`.
func parseTokenHeader(encoded string) (tokenHeader, error) {
	hdrBytes, err := decode(encoded)
	if err != nil {
		return tokenHeader{}, domain.ErrAuthTokenInvalid
	}
	var hdr tokenHeader
	if err := json.Unmarshal(hdrBytes, &hdr); err != nil {
		return tokenHeader{}, domain.ErrAuthTokenInvalid
	}
	if !domain.SigningKeyAlgorithm(hdr.Alg).IsKnown() {
		return tokenHeader{}, domain.ErrAuthTokenInvalid
	}
	if hdr.Kid == "" {
		return tokenHeader{}, domain.ErrAuthTokenInvalid
	}
	return hdr, nil
}

// lookupKeyForHeader sucht den Verify-Key über die `kid` und stellt
// sicher, dass sein Algorithmus mit dem Header übereinstimmt.
func (s *HMACSessionTokenSigner) lookupKeyForHeader(hdr tokenHeader) (domain.SessionSigningKey, error) {
	verifyKeys, err := s.Active.AllVerifyKeys()
	if err != nil {
		return domain.SessionSigningKey{}, err
	}
	key, err := domain.LookupSigningKey(verifyKeys, domain.SigningKeyID(hdr.Kid))
	if err != nil {
		return domain.SessionSigningKey{}, err
	}
	if key.Algorithm != domain.SigningKeyAlgorithm(hdr.Alg) {
		return domain.SessionSigningKey{}, domain.ErrAuthTokenInvalid
	}
	return key, nil
}

// verifyHMACSignature berechnet `HMAC-SHA-256(secret, "header.claims")`
// und vergleicht in konstanter Zeit gegen die präsentierte Signatur.
func verifyHMACSignature(secret []byte, headerSeg, claimsSeg, sigSeg string) error {
	signingInput := headerSeg + "." + claimsSeg
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return err
	}
	expected := mac.Sum(nil)
	provided, err := decode(sigSeg)
	if err != nil {
		return domain.ErrAuthTokenInvalid
	}
	if !domain.ConstantTimeEqualSignature(expected, provided) {
		return domain.ErrAuthTokenInvalid
	}
	return nil
}

// decodeClaimSet baut den `domain.SessionTokenClaims` aus dem
// signierten Body. Format-Verstöße liefern `ErrAuthTokenInvalid`.
func decodeClaimSet(encoded string) (domain.SessionTokenClaims, error) {
	bodyBytes, err := decode(encoded)
	if err != nil {
		return domain.SessionTokenClaims{}, domain.ErrAuthTokenInvalid
	}
	var body claimSet
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return domain.SessionTokenClaims{}, domain.ErrAuthTokenInvalid
	}
	out := domain.SessionTokenClaims{
		Iss: body.Iss,
		Sub: body.Sub,
		Aud: domain.SessionTokenAudience(body.Aud),
		Iat: time.Unix(body.Iat, 0).UTC(),
		Nbf: time.Unix(body.Nbf, 0).UTC(),
		Exp: time.Unix(body.Exp, 0).UTC(),
		JTI: body.Jti,
	}
	if body.SessionID != "" {
		v := body.SessionID
		out.SessionID = &v
	}
	if body.Origin != "" {
		v := body.Origin
		out.Origin = &v
	}
	return out, nil
}

// encode wraps base64.RawURLEncoding für die JWT-typische URL-sichere
// Base64-Variante ohne Padding.
func encode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// decode ist das Pendant zu `encode`. Gibt einen wrap'-fähigen Fehler
// zurück, damit der Aufrufer ihn auf `domain.ErrAuthTokenInvalid`
// mappen kann.
func decode(s string) ([]byte, error) {
	if s == "" {
		return nil, errors.New("empty base64 segment")
	}
	out, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	return out, nil
}
