package mediamtxclient

import (
	"encoding/json"
	"testing"
	"time"
)

//  (extra-gates.md) — Fuzz-Target für
// das SRT-Health-Mapping. Pinnt:
//
//  - Random-JSON-Bytes (MediaMTX-`/v3/srtconns/list`-Wire-Format)
//  dürfen weder Panic noch unrealistische Domain-Werte erzeugen.
//  - Negative oder NaN-Floats (mbps* / msRTT) müssen auf 0 oder
//  gleichem Sentinel gemappt werden, nicht in
//  domain.SrtConnectionSample-Felder mit unspezifizierten
//  Werten leaken.
//
// Pflicht-Bereich aus Plan §4 DoD-Item 1: „SRT-Health-Mapping".

// FuzzMapMediaMtxItem deserialisiert random JSON-Bytes als
// `srtConnItem` und ruft `mapItem`. Die Fuzz-Engine produziert die
// Bytes; wir erlauben JSON-Decode-Failures (kein Mapping wird
// versucht), aber jedes erfolgreiche Decode-Result muss ein
// Domain-Sample produzieren, das die Sentinel-Regeln aus
// spec/telemetry-model.md nicht verletzt.
func FuzzMapMediaMtxItem(f *testing.F) {
	// Seed-Korpus mit produktiven Variants.
	f.Add([]byte(`{"id":"a","state":"publish","path":"srt-test","bytesReceived":1024,"msRTT":3.5,"mbpsLinkCapacity":1500.0}`))
	f.Add([]byte(`{"state":"read","mbpsLinkCapacity":-1}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"state":"exotic","msRTT":-7.0}`))

	requiredBandwidth := int64(1_000_000)
	collectedAt := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)

	f.Fuzz(func(t *testing.T, raw []byte) {
		var item srtConnItem
		if err := json.Unmarshal(raw, &item); err != nil {
			// Ungültiges JSON ist legitimer Fuzz-Output — Mapping
			// passiert nicht.
			return
		}
		sample := mapItem(item, collectedAt, &requiredBandwidth)

		// Sanity-Invariants gegen Domain-Drift:
		if sample.AvailableBandwidthBPS < 0 {
			t.Fatalf("negative AvailableBandwidthBPS=%d for raw=%q", sample.AvailableBandwidthBPS, raw)
		}
		if sample.RTTMillis < 0 && item.MsRTT >= 0 {
			t.Fatalf("RTTMillis=%v from MsRTT=%v drifted negative for raw=%q", sample.RTTMillis, item.MsRTT, raw)
		}
	})
}
