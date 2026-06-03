package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// RAK-66: Stream-Key-Erzeugung mit CSPRNG,
// SHA-256-Hash, redigiertem Fingerprint und Konstantzeit-Vergleich.
// Tests laufen ohne HTTP, Storage oder Docker.

func mustGenerate(t *testing.T) domain.StreamKeyMaterial {
	t.Helper()
	material, err := domain.GenerateStreamKey(time.Unix(1715235600, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateStreamKey: %v", err)
	}
	return material
}

func TestGenerateStreamKey_FormatAndPrefix(t *testing.T) {
	t.Parallel()
	m := mustGenerate(t)
	if !strings.HasPrefix(m.Value, "mtr_ing_") {
		t.Errorf("Value missing prefix: %q", m.Value)
	}
	// Klartext-Länge: Prefix (8) + Base64(32 Bytes Entropie) ohne
	// Padding = 8 + 43 = 51.
	if len(m.Value) != 51 {
		t.Errorf("Value length: want 51, got %d", len(m.Value))
	}
	// SHA-256 hex = 64 Zeichen.
	if len(m.Hash) != 64 {
		t.Errorf("Hash length: want 64 hex chars, got %d", len(m.Hash))
	}
	// Fingerprint: 8 head + 3 dots + 4 tail = 15 Zeichen.
	if len(m.Fingerprint) != 15 {
		t.Errorf("Fingerprint length: want 15, got %d (%q)", len(m.Fingerprint), m.Fingerprint)
	}
	if !strings.HasPrefix(m.Fingerprint, "mtr_ing_") {
		t.Errorf("Fingerprint missing prefix: %q", m.Fingerprint)
	}
	if !strings.Contains(m.Fingerprint, "...") {
		t.Errorf("Fingerprint missing ellipsis: %q", m.Fingerprint)
	}
	if m.CreatedAt.IsZero() {
		t.Errorf("CreatedAt is zero")
	}
}

func TestGenerateStreamKey_Uniqueness(t *testing.T) {
	t.Parallel()
	const N = 1000
	seen := make(map[string]struct{}, N)
	hashes := make(map[string]struct{}, N)
	for i := 0; i < N; i++ {
		m, err := domain.GenerateStreamKey(time.Now().UTC())
		if err != nil {
			t.Fatalf("GenerateStreamKey [%d]: %v", i, err)
		}
		if _, dup := seen[m.Value]; dup {
			t.Fatalf("duplicate Value at iteration %d: %q", i, m.Value)
		}
		seen[m.Value] = struct{}{}
		if _, dup := hashes[m.Hash]; dup {
			t.Fatalf("duplicate Hash at iteration %d", i)
		}
		hashes[m.Hash] = struct{}{}
	}
}

func TestGenerateStreamKey_HashIsStableForSameValue(t *testing.T) {
	t.Parallel()
	// Wir können den Klartext nicht direkt steuern, aber ein
	// Round-Trip Validate(value, ToPersistable) muss true liefern.
	m := mustGenerate(t)
	persisted := m.ToPersistable()
	ok, err := domain.ValidateStreamKey(m.Value, persisted)
	if err != nil {
		t.Fatalf("ValidateStreamKey: %v", err)
	}
	if !ok {
		t.Errorf("Validate must accept the just-generated key")
	}
}

func TestValidateStreamKey_RejectsWrongValue(t *testing.T) {
	t.Parallel()
	m := mustGenerate(t)
	persisted := m.ToPersistable()
	wrong := "mtr_ing_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	ok, err := domain.ValidateStreamKey(wrong, persisted)
	if err != nil {
		t.Errorf("Validate must not error on well-formed wrong key: %v", err)
	}
	if ok {
		t.Errorf("Validate must reject a wrong but well-formed key")
	}
}

func TestValidateStreamKey_RejectsMissingPrefix(t *testing.T) {
	t.Parallel()
	m := mustGenerate(t)
	persisted := m.ToPersistable()
	ok, err := domain.ValidateStreamKey("not-a-stream-key", persisted)
	if !errors.Is(err, domain.ErrStreamKeyMalformed) {
		t.Errorf("expected ErrStreamKeyMalformed, got %v", err)
	}
	if ok {
		t.Errorf("must not be valid")
	}
}

func TestValidateStreamKey_RejectsEmptyValue(t *testing.T) {
	t.Parallel()
	m := mustGenerate(t)
	persisted := m.ToPersistable()
	ok, err := domain.ValidateStreamKey("", persisted)
	if !errors.Is(err, domain.ErrStreamKeyMalformed) {
		t.Errorf("expected ErrStreamKeyMalformed, got %v", err)
	}
	if ok {
		t.Errorf("must not be valid")
	}
}

func TestStreamKeyMaterial_ToPersistableExcludesValue(t *testing.T) {
	t.Parallel()
	m := mustGenerate(t)
	p := m.ToPersistable()
	if p.Hash != m.Hash {
		t.Errorf("Hash mismatch")
	}
	if p.Fingerprint != m.Fingerprint {
		t.Errorf("Fingerprint mismatch")
	}
	// Plan T1 DoD: kein Domain-Test snapshotet echte Klartext-Keys.
	// Persistente Sicht hat per Konstruktion kein Value-Feld; dieser
	// Test pinnt das anhand des Reflection-frei zugänglichen
	// Felder-Sets. Wenn jemand StreamKey um ein Klartext-Value-Feld
	// erweitert, fällt das auf — bricht aber den Build, nicht diesen
	// Test. Hier verifizieren wir nur, dass die offensichtliche
	// Zuordnung (Hash/Fingerprint) korrekt durchgereicht wird.
	if !p.CreatedAt.Equal(m.CreatedAt) {
		t.Errorf("CreatedAt mismatch")
	}
}
