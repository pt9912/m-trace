package auth_test

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// b64 wraps `base64.RawURLEncoding.EncodeToString` so the table data
// stays readable as plain ASCII secret strings.
func b64(t *testing.T, s string) string {
	t.Helper()
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func TestParseSigningKeysEnv_MultiKeyHappyPath(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	secretA := b64(t, "secret-a-32-bytes-1234567890abcd")
	secretB := b64(t, "secret-b-32-bytes-1234567890abcd")
	keys, active, noConfig, err := auth.ParseSigningKeysEnv(
		"kid_a:"+secretA+",kid_b:"+secretB,
		"kid_b",
		"", "",
		now,
	)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if noConfig {
		t.Fatalf("noKeysConfigured must be false when multi-key set")
	}
	if active != "kid_b" {
		t.Errorf("active KID: want kid_b, got %s", active)
	}
	if len(keys) != 2 {
		t.Fatalf("want 2 keys, got %d", len(keys))
	}
	if keys[0].KID != "kid_a" || keys[1].KID != "kid_b" {
		t.Errorf("order not preserved: %q %q", keys[0].KID, keys[1].KID)
	}
	if string(keys[0].Secret) != "secret-a-32-bytes-1234567890abcd" {
		t.Errorf("secret_a decoded mismatch")
	}
}

func TestParseSigningKeysEnv_SingleKeyBackwardsCompat(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	secret := b64(t, "single-key-secret-32-bytes-1234")
	keys, active, noConfig, err := auth.ParseSigningKeysEnv(
		"", "",
		secret, "kid_legacy",
		now,
	)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if noConfig {
		t.Fatalf("noKeysConfigured must be false when fallback set")
	}
	if active != "kid_legacy" {
		t.Errorf("active KID: want kid_legacy, got %s", active)
	}
	if len(keys) != 1 {
		t.Fatalf("want 1 key, got %d", len(keys))
	}
	if keys[0].KID != "kid_legacy" {
		t.Errorf("KID mismatch: got %s", keys[0].KID)
	}
}

func TestParseSigningKeysEnv_NoKeysConfigured(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	keys, active, noConfig, err := auth.ParseSigningKeysEnv(
		"", "",
		"", "lab-default",
		now,
	)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !noConfig {
		t.Fatalf("noKeysConfigured must be true when nothing set")
	}
	if keys != nil {
		t.Errorf("keys must be nil, got %d entries", len(keys))
	}
	if active != "lab-default" {
		t.Errorf("active KID should pass through fallback hint, got %s", active)
	}
}

func TestParseSigningKeysEnv_MultiKeyRequiresActiveKID(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	secret := b64(t, "secret-a-32-bytes-1234567890abcd")
	_, _, _, err := auth.ParseSigningKeysEnv(
		"kid_a:"+secret,
		"",
		"", "",
		now,
	)
	if err == nil {
		t.Fatalf("expected error when MTRACE_AUTH_SIGNING_ACTIVE_KID is missing")
	}
	if !strings.Contains(err.Error(), "MTRACE_AUTH_SIGNING_ACTIVE_KID") {
		t.Errorf("error must mention missing active kid env: %v", err)
	}
}

func TestParseSigningKeysEnv_ActiveKIDNotInList(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	secret := b64(t, "secret-a-32-bytes-1234567890abcd")
	_, _, _, err := auth.ParseSigningKeysEnv(
		"kid_a:"+secret,
		"kid_z",
		"", "",
		now,
	)
	if err == nil {
		t.Fatalf("expected error when active kid not in list")
	}
	if !strings.Contains(err.Error(), "kid_z") {
		t.Errorf("error must name the missing active kid: %v", err)
	}
}

func TestParseSigningKeysEnv_DuplicateKID(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	secret := b64(t, "secret-a-32-bytes-1234567890abcd")
	_, _, _, err := auth.ParseSigningKeysEnv(
		"kid_a:"+secret+",kid_a:"+secret,
		"kid_a",
		"", "",
		now,
	)
	if err == nil {
		t.Fatalf("expected error on duplicate kid")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error must mention duplicate: %v", err)
	}
}

func TestParseSigningKeysEnv_EmptyKID(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	secret := b64(t, "secret-a-32-bytes-1234567890abcd")
	_, _, _, err := auth.ParseSigningKeysEnv(
		":"+secret,
		"any",
		"", "",
		now,
	)
	if err == nil {
		t.Fatalf("expected error on empty kid")
	}
}

func TestParseSigningKeysEnv_EmptySecret(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	_, _, _, err := auth.ParseSigningKeysEnv(
		"kid_a:",
		"kid_a",
		"", "",
		now,
	)
	if err == nil {
		t.Fatalf("expected error on empty secret")
	}
}

func TestParseSigningKeysEnv_InvalidBase64(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	_, _, _, err := auth.ParseSigningKeysEnv(
		"kid_a:not!valid!base64!*&^",
		"kid_a",
		"", "",
		now,
	)
	if err == nil {
		t.Fatalf("expected error on invalid base64")
	}
}

func TestParseSigningKeysEnv_MalformedEntry(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	_, _, _, err := auth.ParseSigningKeysEnv(
		"kid_a_no_colon",
		"kid_a_no_colon",
		"", "",
		now,
	)
	if err == nil {
		t.Fatalf("expected error on missing colon")
	}
}

func TestParseSigningKeysEnv_WhitespaceAndEmptyEntries(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	secret := b64(t, "secret-a-32-bytes-1234567890abcd")
	// Mit umgebenden Spaces, doppeltem Komma und führendem/trailing
	// Komma — der Parser muss alles tolerieren.
	keys, active, _, err := auth.ParseSigningKeysEnv(
		" ,kid_a : "+secret+" , ,kid_b: "+secret+" ,",
		" kid_a ",
		"", "",
		now,
	)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if active != "kid_a" {
		t.Errorf("active KID: want kid_a (after trim), got %q", active)
	}
	if len(keys) != 2 {
		t.Fatalf("want 2 keys, got %d", len(keys))
	}
}

func TestParseSigningKeysEnv_SingleKeyMissingKID(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	secret := b64(t, "single-key-secret-32-bytes-1234")
	_, _, _, err := auth.ParseSigningKeysEnv(
		"", "",
		secret, "",
		now,
	)
	if err == nil {
		t.Fatalf("expected error when SIGNING_KEY set but SIGNING_KID missing")
	}
	if !strings.Contains(err.Error(), "MTRACE_AUTH_SIGNING_KID") {
		t.Errorf("error must mention missing kid env: %v", err)
	}
}

func TestParseSigningKeysEnv_SingleKeyInvalidBase64(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	_, _, _, err := auth.ParseSigningKeysEnv(
		"", "",
		"not!valid!base64!*&^", "kid",
		now,
	)
	if err == nil {
		t.Fatalf("expected error on invalid base64 for single-key fallback")
	}
}

// TestParseSigningKeysEnv_RotationEndToEnd verifiziert das Rotation-
// Szenario aus `auth.md` §5.3.1 als Code-Pfad (RAK-78): ein Token,
// das mit `kid_a` signiert wurde, muss nach ACTIVE-Umschaltung auf
// `kid_b` weiterhin verifizieren, solange `kid_a` im Key-Ring bleibt.
func TestParseSigningKeysEnv_RotationEndToEnd(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	secretA := b64(t, "rotation-secret-a-32bytes-abcdef")
	secretB := b64(t, "rotation-secret-b-32bytes-abcdef")
	keysEnv := "kid_a:" + secretA + ",kid_b:" + secretB

	// Phase 1: ACTIVE=kid_a. Token unter kid_a signieren.
	keys1, active1, _, err := auth.ParseSigningKeysEnv(keysEnv, "kid_a", "", "", now)
	if err != nil {
		t.Fatalf("phase1 parse: %v", err)
	}
	resolver1, err := auth.NewMultiKeySigningResolver(active1, keys1...)
	if err != nil {
		t.Fatalf("phase1 resolver: %v", err)
	}
	signer1 := auth.NewHMACSessionTokenSigner(resolver1)
	claims := domain.SessionTokenClaims{
		Iss: "m-trace", Sub: "p1",
		Aud: domain.SessionTokenAudiencePlaybackEvents,
		Iat: now, Nbf: now, Exp: now.Add(15 * time.Minute),
		JTI: "st_rotation_test",
	}
	tok, err := signer1.Sign(claims)
	if err != nil {
		t.Fatalf("phase1 sign: %v", err)
	}

	// Phase 2: ACTIVE=kid_b (Rotation), kid_a bleibt im Verify-Set.
	keys2, active2, _, err := auth.ParseSigningKeysEnv(keysEnv, "kid_b", "", "", now)
	if err != nil {
		t.Fatalf("phase2 parse: %v", err)
	}
	resolver2, err := auth.NewMultiKeySigningResolver(active2, keys2...)
	if err != nil {
		t.Fatalf("phase2 resolver: %v", err)
	}
	signer2 := auth.NewHMACSessionTokenSigner(resolver2)

	// Alt-Token muss weiterhin verifizieren.
	if _, err := signer2.Verify(tok); err != nil {
		t.Fatalf("rotated resolver must verify pre-rotation token: %v", err)
	}

	// Neue Tokens werden mit kid_b signiert.
	newTok, err := signer2.Sign(claims)
	if err != nil {
		t.Fatalf("phase2 sign: %v", err)
	}
	if newTok == tok {
		t.Errorf("rotated signer must produce different token bytes (different kid header)")
	}
}
