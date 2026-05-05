package mediamtxclient

import (
	"strconv"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// srtConnsListResponse spiegelt die MediaMTX-Antwort von
// `GET /v3/srtconns/list` so weit wie nötig. Felder sind exakt nach
// Probe-Befund (plan-0.6.0 §2.4) plus Fixture
// `spec/contract-fixtures/srt/mediamtx-srtconns-list.json` benannt.
type srtConnsListResponse struct {
	ItemCount int           `json:"itemCount"`
	PageCount int           `json:"pageCount"`
	Items     []srtConnItem `json:"items"`
}

// srtConnItem trägt alle für RAK-43 plus optional erweiterte Signale
// nötigen Felder. Unbekannte Felder werden vom JSON-Decoder
// stillschweigend ignoriert (Forward-Compatibility gegen MediaMTX-
// Versionen).
type srtConnItem struct {
	ID                     string  `json:"id"`
	Created                string  `json:"created"`
	RemoteAddr             string  `json:"remoteAddr"`
	State                  string  `json:"state"`
	Path                   string  `json:"path"`
	BytesReceived          int64   `json:"bytesReceived"`
	PacketsReceived        int64   `json:"packetsReceived"`
	PacketsReceivedLoss    int64   `json:"packetsReceivedLoss"`
	PacketsReceivedLossRate float64 `json:"packetsReceivedLossRate"`
	PacketsRetrans         int64   `json:"packetsRetrans"`
	PacketsReceivedRetrans int64   `json:"packetsReceivedRetrans"`
	MsRTT                  float64 `json:"msRTT"`
	MbpsReceiveRate        float64 `json:"mbpsReceiveRate"`
	MbpsSendRate           float64 `json:"mbpsSendRate"`
	MbpsLinkCapacity       float64 `json:"mbpsLinkCapacity"`
	MbpsMaxBW              float64 `json:"mbpsMaxBW"`
	MsReceiveBuf           int64   `json:"msReceiveBuf"`
	BytesReceiveBuf        int64   `json:"bytesReceiveBuf"`
	PacketsReceiveBuf      int64   `json:"packetsReceiveBuf"`
}

// mapItem konvertiert einen MediaMTX-Eintrag in einen Domain-Sample.
// Konventionen aus spec/telemetry-model.md §7.1:
//   - mbpsLinkCapacity × 1_000_000 → AvailableBandwidthBPS (bps)
//   - mbpsReceiveRate × 1_000_000 → ThroughputBPS (optional)
//   - bytesReceived (string) → SourceSequence (Surrogat-Sequence)
//   - state ∈ {"publish","read"} → ConnectionStateConnected
//
// Wenn ein Pflichtfeld semantisch fehlt (z. B. mbpsLinkCapacity ≤ 0,
// state leer), markiert der Adapter ConnectionState als `unknown` —
// Evaluate in der Application-Schicht klassifiziert das als
// `partial`.
func mapItem(it srtConnItem, collectedAt time.Time) domain.SrtConnectionSample {
	state := mapState(it.State)
	available := int64(it.MbpsLinkCapacity * 1_000_000)
	if it.MbpsLinkCapacity <= 0 {
		// Markiere die Quelle als partial — Evaluate hat das Mapping.
		state = domain.ConnectionStateUnknown
	}

	throughput := optionalThroughput(it.MbpsReceiveRate)
	lossRate := optionalLossRate(it.PacketsReceivedLossRate)

	sample := domain.SrtConnectionSample{
		StreamID:     it.Path,
		ConnectionID: it.ID,

		// MediaMTX liefert keinen Source-Sample-Timestamp;
		// ConnectedAt (`Created`) ist Verbindungs-Beginn, nicht
		// Sample-Zeit. SourceObservedAt bleibt Zero — das Domain-
		// Modell behandelt das via SourceSequence-Surrogat
		// (spec §7.6).
		SourceSequence: strconv.FormatInt(it.BytesReceived, 10),
		CollectedAt:    collectedAt,

		RTTMillis:             it.MsRTT,
		PacketLossTotal:       it.PacketsReceivedLoss,
		PacketLossRate:        lossRate,
		RetransmissionsTotal:  it.PacketsReceivedRetrans,
		AvailableBandwidthBPS: available,
		ThroughputBPS:         throughput,

		ConnectionState: state,
	}

	return sample
}

// mapState bildet den MediaMTX-`state` auf domain.ConnectionState.
// Unbekannter / leerer State → `unknown` (Evaluate klassifiziert
// das als `partial`).
func mapState(state string) domain.ConnectionState {
	switch state {
	case "publish", "read":
		return domain.ConnectionStateConnected
	case "":
		return domain.ConnectionStateUnknown
	default:
		// MediaMTX kann auch "idle" liefern, wenn eine Verbindung im
		// Aufbau ist. Dass wir das hier als unknown klassifizieren,
		// schützt die Health-Bewertung vor Werten ohne aktiven
		// Stream.
		return domain.ConnectionStateUnknown
	}
}

// optionalThroughput konvertiert mbps → bps und liefert nil, falls
// die Quelle keinen positiven Wert liefert (häufig bei
// Publish-Connections, die nichts senden).
func optionalThroughput(mbps float64) *int64 {
	if mbps <= 0 {
		return nil
	}
	v := int64(mbps * 1_000_000)
	return &v
}

// optionalLossRate liefert nil, falls die Quelle 0 meldet — der
// Use Case darf `0` von `nicht-geliefert` unterscheiden, wir
// behandeln 0 als "kein Loss" und reichen stattdessen den absoluten
// Counter weiter (PacketLossTotal). Wer einen explizit nicht-null
// Rate-Wert sehen will, prüft im Use Case PacketLossRate != nil.
func optionalLossRate(rate float64) *float64 {
	if rate <= 0 {
		return nil
	}
	return &rate
}

