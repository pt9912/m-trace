package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// webrtcMetrics bündelt die `mtrace_webrtc_*`-Counter aus
// spec/telemetry-model.md plus den Sample-State für die
// serverseitige Delta-Berechnung.
//
// State-Counter (`connection_state`, `ice_state`, `dtls_state`)
// zählen angenommene Samples — nicht Gauges. Counter-Felder
// (`packets_lost`, `bytes_received`, `bytes_sent`) sind label-frei
// (außer Target-Metadaten) und werden serverseitig deltadiffenziert:
// erster Sample setzt nur die Baseline, Folge-Samples inkrementieren
// den Counter um max(0, current - last).
//
// Idempotenz: Samples mit `webrtc.sample_id ≤ last_sample_id`
// inkrementieren keinen Counter und überschreiben den Last-Sample-
// State nicht. Reconnect bekommt eine neue `peer_connection_run_id`
// und damit einen neuen Eintrag im Sample-State.
type webrtcMetrics struct {
	connectionStateTotal *prometheus.CounterVec
	iceStateTotal        *prometheus.CounterVec
	dtlsStateTotal       *prometheus.CounterVec

	packetsLostTotal   prometheus.Counter
	bytesReceivedTotal prometheus.Counter
	bytesSentTotal     prometheus.Counter

	mu    sync.Mutex
	state map[string]*webrtcRunState
}

type webrtcRunState struct {
	lastSampleID  int64
	packetsLost   int64
	bytesReceived int64
	bytesSent     int64
}

func newWebRTCMetrics() *webrtcMetrics {
	return &webrtcMetrics{
		connectionStateTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mtrace_webrtc_connection_state_total",
				Help: "Total accepted WebRTC samples grouped by RTCPeerConnectionState.",
			},
			[]string{"connection_state"},
		),
		iceStateTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mtrace_webrtc_ice_state_total",
				Help: "Total accepted WebRTC samples grouped by RTCIceConnectionState.",
			},
			[]string{"ice_state"},
		),
		dtlsStateTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mtrace_webrtc_dtls_state_total",
				Help: "Total accepted WebRTC samples grouped by RTCDtlsTransportState.",
			},
			[]string{"dtls_state"},
		),
		packetsLostTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_webrtc_packets_lost_total",
			Help: "Total non-negative deltas of getStats() packetsLost across all PeerConnection runs.",
		}),
		bytesReceivedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_webrtc_bytes_received_total",
			Help: "Total non-negative deltas of getStats() bytesReceived across all PeerConnection runs.",
		}),
		bytesSentTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_webrtc_bytes_sent_total",
			Help: "Total non-negative deltas of getStats() bytesSent across all PeerConnection runs.",
		}),
		state: make(map[string]*webrtcRunState),
	}
}

func (m *webrtcMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.connectionStateTotal,
		m.iceStateTotal,
		m.dtlsStateTotal,
		m.packetsLostTotal,
		m.bytesReceivedTotal,
		m.bytesSentTotal,
	}
}

func (m *webrtcMetrics) record(s driven.WebRTCSampleSnapshot) {
	// State-Counter immer inkrementieren — sie zählen angenommene
	// Samples, unabhängig vom Sample-State (§3.5.1).
	m.connectionStateTotal.WithLabelValues(s.ConnectionState).Inc()
	m.iceStateTotal.WithLabelValues(s.IceState).Inc()
	m.dtlsStateTotal.WithLabelValues(s.DtlsState).Inc()

	key := s.ProjectID + "|" + s.SessionID + "|" + s.RunID
	m.mu.Lock()
	defer m.mu.Unlock()
	prev, ok := m.state[key]
	if !ok {
		// Erster Sample für diese Run-ID → Baseline; kein Delta-Increment.
		m.state[key] = &webrtcRunState{
			lastSampleID:  s.SampleID,
			packetsLost:   s.PacketsLost,
			bytesReceived: s.BytesReceived,
			bytesSent:     s.BytesSent,
		}
		return
	}
	// Idempotenz: Duplicate/Retry mit gleichem oder älterem sample_id
	// inkrementieren keinen Counter und aktualisieren den State nicht.
	if s.SampleID <= prev.lastSampleID {
		return
	}
	addDelta(m.packetsLostTotal, s.PacketsLost, prev.packetsLost)
	addDelta(m.bytesReceivedTotal, s.BytesReceived, prev.bytesReceived)
	addDelta(m.bytesSentTotal, s.BytesSent, prev.bytesSent)
	prev.lastSampleID = s.SampleID
	prev.packetsLost = s.PacketsLost
	prev.bytesReceived = s.BytesReceived
	prev.bytesSent = s.BytesSent
}

func addDelta(c prometheus.Counter, current, last int64) {
	delta := current - last
	if delta <= 0 {
		// Negativer/Null-Delta inkrementiert nicht; State wird durch
		// Caller auf den neuen Wert aktualisiert (Counter-Reset-Modell).
		return
	}
	c.Add(float64(delta))
}
