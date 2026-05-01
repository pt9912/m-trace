package domain

// StreamAnalysisRequest beschreibt einen Manifestanalyse-Auftrag,
// unabhängig vom konkreten Manifestformat. 0.3.0 deckt HLS ab; weitere
// Formate (DASH/CMAF, F-73) werden additiv ergänzt, ohne den Vertrag
// zu brechen (plan-0.3.0 §2 Tranche 1).
//
// Genau eines von ManifestText oder ManifestURL muss gesetzt sein. Die
// konkrete Eingabevalidierung erfolgt im Adapter, weil Lade-Politik
// (Timeout, Größe, SSRF) dort implementiert wird (plan-0.3.0 §3
// Tranche 2).
type StreamAnalysisRequest struct {
	// ManifestText ist der Manifestinhalt selbst.
	ManifestText string
	// ManifestURL benennt die Quelle, falls der Adapter selbst lädt.
	ManifestURL string
	// BaseURL wird für relative URIs im Manifest genutzt; optional.
	BaseURL string
}

// PlaylistType klassifiziert das Analyseergebnis grob. Weitere Werte
// (z. B. DASH-MPD-Varianten) sind additiv erlaubt, ohne bestehende
// Konsumenten zu brechen.
type PlaylistType string

const (
	// PlaylistTypeUnknown signalisiert, dass weder Master- noch
	// Media-Playlist sicher erkannt wurde.
	PlaylistTypeUnknown PlaylistType = "unknown"
	// PlaylistTypeMaster steht für eine HLS Master Playlist.
	PlaylistTypeMaster PlaylistType = "master"
	// PlaylistTypeMedia steht für eine HLS Media Playlist.
	PlaylistTypeMedia PlaylistType = "media"
)

// FindingLevel folgt der Drei-Stufen-Skala aus plan-0.3.0 Tranche 4.
type FindingLevel string

const (
	// FindingLevelInfo dokumentiert beobachtete Eigenschaften ohne
	// Handlungsbedarf.
	FindingLevelInfo FindingLevel = "info"
	// FindingLevelWarning markiert Auffälligkeiten, die das Manifest
	// noch interpretierbar lassen.
	FindingLevelWarning FindingLevel = "warning"
	// FindingLevelError markiert Verstöße, die das Manifest nicht mehr
	// als gültig auswerten lassen.
	FindingLevelError FindingLevel = "error"
)

// StreamAnalysisFinding bildet eine einzelne Auffälligkeit ab.
type StreamAnalysisFinding struct {
	Code    string
	Level   FindingLevel
	Message string
}

// StreamAnalysisErrorCode klassifiziert von analyzer-Implementierungen
// gemeldete Fehler (RFC-7807-ähnliche Code-Domäne aus
// `@npm9912/stream-analyzer.AnalysisErrorResult.code`). Driving-
// Adapter mappen die Codes auf HTTP-Statuscodes (plan-0.3.0 §7).
type StreamAnalysisErrorCode string

const (
	// StreamAnalysisInvalidInput: der Aufrufer hat eine fehlerhafte
	// Eingabe geliefert (kind/text/url-Form). HTTP 400.
	StreamAnalysisInvalidInput StreamAnalysisErrorCode = "invalid_input"
	// StreamAnalysisManifestNotHLS: das Manifest ist kein HLS-Inhalt.
	// HTTP 422 — semantisch wohlgeformt, aber inhaltlich nicht
	// verarbeitbar.
	StreamAnalysisManifestNotHLS StreamAnalysisErrorCode = "manifest_not_hls"
	// StreamAnalysisFetchBlocked: die übergebene URL wurde vom
	// SSRF-Schutz abgelehnt (privat/loopback/credentials/scheme).
	// HTTP 400 — der Aufrufer hat eine unsichere URL geliefert.
	StreamAnalysisFetchBlocked StreamAnalysisErrorCode = "fetch_blocked"
	// StreamAnalysisFetchFailed: das URL-Laden ist netzwerk-/status-
	// /content-type-bedingt fehlgeschlagen. HTTP 502 — wir agieren
	// als Gateway, Upstream hat versagt.
	StreamAnalysisFetchFailed StreamAnalysisErrorCode = "fetch_failed"
	// StreamAnalysisManifestTooLarge: geladenes Manifest überschritt
	// das maxBytes-Limit. HTTP 502.
	StreamAnalysisManifestTooLarge StreamAnalysisErrorCode = "manifest_too_large"
	// StreamAnalysisInternalError: unerwarteter Fehler im Analyzer
	// selbst. HTTP 502 (Gateway-Sicht).
	StreamAnalysisInternalError StreamAnalysisErrorCode = "internal_error"
)

// StreamAnalysisDomainError signalisiert, dass der Analyzer den
// Aufruf bewusst abgelehnt hat (Code aus einer abgeschlossenen
// Domäne). Das ist KEIN Transport-/Verfügbarkeitsproblem — Driving-
// Adapter müssen die beiden Klassen unterscheiden, damit Konsumenten
// nicht "analyzer unavailable" lesen, wenn sie tatsächlich eine
// fehlerhafte Eingabe gesendet haben.
type StreamAnalysisDomainError struct {
	Code    StreamAnalysisErrorCode
	Message string
	// Details enthält strukturierte Zusatzinfos aus dem analyzer-
	// Result (z. B. {host, address, family} bei fetch_blocked). Das
	// Mapping auf HTTP-Body-Felder ist Sache des Driving-Adapters.
	Details map[string]any
}

// Error erfüllt das error-Interface.
func (e *StreamAnalysisDomainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return string(e.Code) + ": " + e.Message
}

// StreamAnalysisSummary aggregiert die Manifestauswertung in eine
// kleine Top-Level-Struktur (RAK-26-konformes Wire-Format). `ItemCount`
// zählt bei Master-Playlists Variants+Renditions, bei Media-Playlists
// Segmente und ist 0 für `unknown`. Tranche 5 in der TS-Library hat
// `AnalysisSummary` als gleichnamigen, additiv erweiterbaren Typ
// dokumentiert (`docs/user/stream-analyzer.md` §2.2); diese Domain-
// Spiegelung erlaubt dem API-Adapter, das Feld vom analyzer-service
// in die `POST /api/analyze`-Antwort durchzureichen, statt es zu
// verwerfen.
type StreamAnalysisSummary struct {
	ItemCount int
}

// StreamAnalysisInputMetadata spiegelt das `input`-Feld aus dem
// AnalysisResult-Wire-Format (Tranche 5). `Source` ist `"text"` oder
// `"url"`; `URL` ist nur bei `Source == "url"` gesetzt; `BaseURL`
// trägt die finale URL nach Redirects (URL-Input) bzw. die optionale
// Base-URL aus dem Text-Input.
type StreamAnalysisInputMetadata struct {
	Source  string
	URL     string
	BaseURL string
}

// StreamAnalysisResult ist das Domain-Modell der Analyseausgabe. Das
// stabile JSON-Schema entsteht in plan-0.3.0 Tranche 5; bis dahin
// reicht die Domain die analyzer-spezifischen Detail-Strukturen als
// vorcodiertes JSON weiter, ohne sich auf eine HLS-spezifische
// Struktur festzulegen.
type StreamAnalysisResult struct {
	// AnalyzerVersion stammt aus packages/stream-analyzer/package.json
	// (plan-0.3.0 §2 Tranche 1: Versionssynchronizität).
	AnalyzerVersion string
	// Input spiegelt die ursprüngliche Eingabeform; vom analyzer-
	// service durchgereicht, damit Konsumenten Text-/URL-Pfade
	// unterscheiden können (RAK-26-Pflichtfeld).
	Input StreamAnalysisInputMetadata
	// PlaylistType klassifiziert das Manifest.
	PlaylistType PlaylistType
	// Summary ist das Aggregat (RAK-26-Pflichtfeld im Wire-Format).
	Summary StreamAnalysisSummary
	// Findings sind die Drei-Stufen-Befunde aus Tranche 4.
	Findings []StreamAnalysisFinding
	// EncodedDetails ist die typspezifische Detail-Sektion, vorcodiert
	// als UTF-8-JSON-Objekt (Tranche 5 fixiert das Schema). Leere
	// Slices signalisieren "kein Detail geliefert"; ungültiges JSON
	// ist ein Adapter-Bug und kein Domain-Zustand.
	EncodedDetails []byte
}
