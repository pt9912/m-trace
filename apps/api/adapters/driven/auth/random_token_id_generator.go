package auth

import (
	"crypto/rand"
	"encoding/base32"
	"strings"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// RandomTokenIDGenerator implementiert `driven.TokenIDGenerator`
// (`0.12.0`, RAK-72) mit `crypto/rand`-Entropie und `st_`-Prefix
// analog zur Wire-Skizze in `spec/backend-api-contract.md` §3.9.
//
// Format: `st_<base32(random16Bytes, ohne Padding, lowercase)>`. 16
// Bytes Entropie reichen für eine `jti`-Eindeutigkeit über die
// erwartete Token-Lebensdauer; eine spätere ULID-Implementierung ist
// drop-in austauschbar (gleicher Port, gleicher Prefix bleibt
// erhalten).
type RandomTokenIDGenerator struct{}

// NewRandomTokenIDGenerator konstruiert den ID-Generator.
func NewRandomTokenIDGenerator() *RandomTokenIDGenerator {
	return &RandomTokenIDGenerator{}
}

// Compile-time check.
var _ driven.TokenIDGenerator = (*RandomTokenIDGenerator)(nil)

// NewTokenID erzeugt eine neue opake `st_`-ID. Crockford-Base32-
// Lowercase, ohne Padding — URL-sicher, gut tippbar, deutlich
// kürzer als Hex.
func (g *RandomTokenIDGenerator) NewTokenID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
	return "st_" + strings.ToLower(encoded), nil
}
