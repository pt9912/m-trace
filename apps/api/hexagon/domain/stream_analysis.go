package domain

// StreamAnalysisRequest beschreibt einen Manifestanalyse-Auftrag,
// unabhängig vom konkreten Manifestformat. 0.3.0 deckt HLS ab; DASH-
// MPD ist (RAK-58 / NF-12) zweiter `AnalyzerKind`-Wert.
// CMAF (NF-13 / RAK-60..RAK-64, ab 0.10.0) ist **kein** neuer
// AnalyzerKind, sondern ein additives `details.cmaf`-Signalmodell
// unter den bestehenden HLS-/DASH-Detail-Objekten — der Adapter
// reicht es als Bestandteil von `EncodedDetails` unverändert durch.
// Weitere Formate werden additiv ergänzt, ohne den Vertrag zu
// brechen.
//
// Genau eines von ManifestText oder ManifestURL muss gesetzt sein. Die
// konkrete Eingabevalidierung erfolgt im Adapter, weil Lade-Politik
// (Timeout, Größe, SSRF) dort implementiert wird
//
// Ab trägt der Request optional Session-Link-Felder
// (`CorrelationID`, `SessionID`) plus den vom HTTP-Adapter aufgelösten
// `ProjectID` (aus `X-MTrace-Token`). Link-Lookups laufen über
// `(ProjectID, CorrelationID)` bzw. `(ProjectID, SessionID)`; ein
// Request ohne Link-Felder bleibt session-los (`detached`).
type StreamAnalysisRequest struct {
	// ManifestText ist der Manifestinhalt selbst.
	ManifestText string
	// ManifestURL benennt die Quelle, falls der Adapter selbst lädt.
	ManifestURL string
	// BaseURL wird für relative URIs im Manifest genutzt; optional.
	BaseURL string
	// ProjectID ist der vom HTTP-Adapter aus `X-MTrace-Token`
	// aufgelöste Project-Kontext. Leer für ungebundene Analyze-
	// Requests (ohne Token, ohne Link-Felder); Pflicht, sobald
	// `CorrelationID` oder `SessionID` gesetzt ist (Vertrag aus
	// API-Kontrakt/§4).
	ProjectID string
	// CorrelationID ist die optionale Server-Korrelations-ID einer
	// bestehenden Session. Hat innerhalb des `ProjectID`-Kontexts
	// Vorrang vor `SessionID`.
	CorrelationID string
	// SessionID ist der optionale Fallback-Link auf eine bestehende
	// Session (siehe API-Kontrakt).
	SessionID string
}

// SessionLinkStatus ist die Statusdomäne der Analyzer-Session-Link-
// Hülle (`{analysis, session_link}`) aus API-Kontrakt und
//  Use-Case entscheidet den Wert pro Request:
//
//  - linked: Session ist bekannt und konsistent.
//  - detached: Request hat keine Link-Felder — bewusst session-los.
//  - not_found_detached: Link-Felder gesetzt, aber Session nicht im
//  Project gefunden.
//  - conflict_detached: beide Link-Felder gesetzt, beide bekannt, aber
//  widersprüchlich (`session_id` zeigt nicht auf die Session der
//  `correlation_id`).
type SessionLinkStatus string

// SessionLink-Statuswerte aus API-Kontrakt
const (
	SessionLinkStatusLinked            SessionLinkStatus = "linked"
	SessionLinkStatusDetached          SessionLinkStatus = "detached"
	SessionLinkStatusNotFoundDetached  SessionLinkStatus = "not_found_detached"
	SessionLinkStatusConflictDetached  SessionLinkStatus = "conflict_detached"
)

// SessionLink ist die strukturierte Link-Hülle der Analyze-Antwort
// (API-Kontrakt, ). `Status` ist Pflicht; alle
// anderen Felder sind optional und tragen die aufgelöste Verknüpfung
// nur bei `Status == linked`.
type SessionLink struct {
	Status        SessionLinkStatus
	ProjectID     string
	SessionID     string
	CorrelationID string
}

// AnalyzeManifestResult bündelt Analyzer-Ergebnis und Session-Link für
// die Antwort `{analysis, session_link}` aus API-Kontrakt
// Driving-Port liefert dieses Resultat ab; der
// HTTP-Adapter mappt auf die JSON-Hülle.
type AnalyzeManifestResult struct {
	Analysis    StreamAnalysisResult
	SessionLink SessionLink
}

// PlaylistType klassifiziert das Analyseergebnis grob. Weitere Werte
// werden additiv ergänzt, ohne bestehende Konsumenten zu brechen.
type PlaylistType string

const (
	// PlaylistTypeUnknown signalisiert, dass weder Master- noch
	// Media-Playlist sicher erkannt wurde.
	PlaylistTypeUnknown PlaylistType = "unknown"
	// PlaylistTypeMaster steht für eine HLS Master Playlist.
	PlaylistTypeMaster PlaylistType = "master"
	// PlaylistTypeMedia steht für eine HLS Media Playlist.
	PlaylistTypeMedia PlaylistType = "media"
	// PlaylistTypeDash steht für ein DASH-MPD (ab
	// , RAK-58 / NF-12). DASH hat keine analoge Master/
	// Media-Trennung in der Manifest-Form selbst — die Period/
	// AdaptationSet/Representation-Hierarchie wird immer in einem
	// MPD geliefert; der Live-/VOD-Status lebt in
	// `details.type` (`static`/`dynamic`).
	PlaylistTypeDash PlaylistType = "dash"
)

// AnalyzerKind identifiziert den ausführenden Analyzer-Pfad.
// Aktuelle Werte: `hls` (seit ) und `dash` (seit
// RAK-58 / NF-12). CMAF (NF-13 / RAK-60..RAK-64,
// ab ) bekommt **keinen** eigenen AnalyzerKind — die
// CMAF-Erkennung lebt als additives `details.cmaf`-Signalmodell
// unter den HLS-/DASH-Detail-Objekten (siehe `StreamAnalysisResult.
// EncodedDetails`). Weitere Manifestformate werden weiterhin additiv
// ergänzt.
type AnalyzerKind string

const (
	// AnalyzerKindHLS markiert ein HLS-Analyse-Result.
	AnalyzerKindHLS AnalyzerKind = "hls"
	// AnalyzerKindDASH markiert ein DASH-MPD-Analyse-Result.
	AnalyzerKindDASH AnalyzerKind = "dash"
)

// FindingLevel folgt der Drei-Stufen-Skala aus.
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
// `@pt9912/stream-analyzer.AnalysisErrorResult.code`). Driving-
// Adapter mappen die Codes auf HTTP-Statuscodes.
type StreamAnalysisErrorCode string

const (
	// StreamAnalysisInvalidInput meldet eine fehlerhafte Aufrufer-
	// Eingabe (kind/text/url-Form). HTTP 400.
	StreamAnalysisInvalidInput StreamAnalysisErrorCode = "invalid_input"
	// StreamAnalysisManifestNotHLS meldet, dass das Manifest vom
	// HLS-Parser abgelehnt wurde, obwohl der Detector es als HLS
	// klassifiziert hat (erste Zeile beginnt mit `#EXTM3U`-Präfix).
	// HTTP 422 — semantisch wohlgeformt, aber inhaltlich nicht
	// verarbeitbar.
	StreamAnalysisManifestNotHLS StreamAnalysisErrorCode = "manifest_not_hls"
	// StreamAnalysisManifestNotSupported meldet, dass der Detector
	// das Manifest weder als HLS (#EXTM3U-Header) noch als DASH
	// (`<?xml`/`<MPD`-Header) klassifizieren konnte. Beispiele:
	// HTML-Bodies, JSON-Bodies, Plain-Text ohne Manifest-Marker,
	// leere Bodies. HTTP 422 (analog `manifest_not_hls`). Ab
	// RAK-58.
	StreamAnalysisManifestNotSupported StreamAnalysisErrorCode = "manifest_not_supported"
	// StreamAnalysisFetchBlocked meldet, dass die übergebene URL vom
	// SSRF-Schutz abgelehnt wurde (privat/loopback/credentials/scheme).
	// HTTP 400 — der Aufrufer hat eine unsichere URL geliefert.
	StreamAnalysisFetchBlocked StreamAnalysisErrorCode = "fetch_blocked"
	// StreamAnalysisFetchFailed meldet, dass das URL-Laden netzwerk-/
	// status-/content-type-bedingt fehlgeschlagen ist. HTTP 502 — wir
	// agieren als Gateway, Upstream hat versagt.
	StreamAnalysisFetchFailed StreamAnalysisErrorCode = "fetch_failed"
	// StreamAnalysisManifestTooLarge meldet, dass das geladene
	// Manifest das maxBytes-Limit überschritten hat. HTTP 502.
	StreamAnalysisManifestTooLarge StreamAnalysisErrorCode = "manifest_too_large"
	// StreamAnalysisInternalError meldet einen unerwarteten Fehler
	// im Analyzer selbst. HTTP 502 (Gateway-Sicht).
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
// Segmente und ist 0 für `unknown`. in der TS-Library hat
// `AnalysisSummary` als gleichnamigen, additiv erweiterbaren Typ
// dokumentiert (`docs/user/stream-analyzer.md` §2.2); diese Domain-
// Spiegelung erlaubt dem API-Adapter, das Feld vom analyzer-service
// in die `POST /api/analyze`-Antwort durchzureichen, statt es zu
// verwerfen.
type StreamAnalysisSummary struct {
	ItemCount int
}

// StreamAnalysisInputMetadata spiegelt das `input`-Feld aus dem
// AnalysisResult-Wire-Format. `Source` ist `"text"` oder
// `"url"`; `URL` ist nur bei `Source == "url"` gesetzt; `BaseURL`
// trägt die finale URL nach Redirects (URL-Input) bzw. die optionale
// Base-URL aus dem Text-Input.
type StreamAnalysisInputMetadata struct {
	Source  string
	URL     string
	BaseURL string
}

// StreamAnalysisResult ist das Domain-Modell der Analyseausgabe. Das
// stabile JSON-Schema entsteht in; bis dahin
// reicht die Domain die analyzer-spezifischen Detail-Strukturen als
// vorcodiertes JSON weiter, ohne sich auf eine HLS-spezifische
// Struktur festzulegen.
type StreamAnalysisResult struct {
	// AnalyzerVersion stammt aus packages/stream-analyzer/package.json
	// (: Versionssynchronizität).
	AnalyzerVersion string
	// AnalyzerKind identifiziert den Analyzer-Pfad (`hls` oder
	// `dash`). Ab RAK-58 wird er aus dem
	// analyzer-service-Wire-Format durchgereicht; ältere Releases
	// hatten ihn als HLS-Konstante in der API.
	AnalyzerKind AnalyzerKind
	// Input spiegelt die ursprüngliche Eingabeform; vom analyzer-
	// service durchgereicht, damit Konsumenten Text-/URL-Pfade
	// unterscheiden können (RAK-26-Pflichtfeld).
	Input StreamAnalysisInputMetadata
	// PlaylistType klassifiziert das Manifest.
	PlaylistType PlaylistType
	// Summary ist das Aggregat (RAK-26-Pflichtfeld im Wire-Format).
	Summary StreamAnalysisSummary
	// Findings sind die Drei-Stufen-Befunde aus.
	Findings []StreamAnalysisFinding
	// EncodedDetails ist die typspezifische Detail-Sektion, vorcodiert
	// als UTF-8-JSON-Objekt (Tranche 5 fixiert das Schema). Leere
	// Slices signalisieren "kein Detail geliefert"; ungültiges JSON
	// ist ein Adapter-Bug und kein Domain-Zustand.
	EncodedDetails []byte
}
