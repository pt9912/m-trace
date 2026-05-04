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
//
// Ab plan-0.4.0 §4.5 trägt der Request optional Session-Link-Felder
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
	// API-Kontrakt §3.6/§4).
	ProjectID string
	// CorrelationID ist die optionale Server-Korrelations-ID einer
	// bestehenden Session. Hat innerhalb des `ProjectID`-Kontexts
	// Vorrang vor `SessionID`.
	CorrelationID string
	// SessionID ist der optionale Fallback-Link auf eine bestehende
	// Session (siehe API-Kontrakt §3.6).
	SessionID string
}

// SessionLinkStatus ist die Statusdomäne der Analyzer-Session-Link-
// Hülle (`{analysis, session_link}`) aus API-Kontrakt §3.6 und
// plan-0.4.0 §4.5. Use-Case entscheidet den Wert pro Request:
//
//   - linked: Session ist bekannt und konsistent.
//   - detached: Request hat keine Link-Felder — bewusst session-los.
//   - not_found_detached: Link-Felder gesetzt, aber Session nicht im
//     Project gefunden.
//   - conflict_detached: beide Link-Felder gesetzt, beide bekannt, aber
//     widersprüchlich (`session_id` zeigt nicht auf die Session der
//     `correlation_id`).
type SessionLinkStatus string

// SessionLink-Statuswerte aus API-Kontrakt §3.6.
const (
	SessionLinkStatusLinked            SessionLinkStatus = "linked"
	SessionLinkStatusDetached          SessionLinkStatus = "detached"
	SessionLinkStatusNotFoundDetached  SessionLinkStatus = "not_found_detached"
	SessionLinkStatusConflictDetached  SessionLinkStatus = "conflict_detached"
)

// SessionLink ist die strukturierte Link-Hülle der Analyze-Antwort
// (API-Kontrakt §3.6, plan-0.4.0 §4.5). `Status` ist Pflicht; alle
// anderen Felder sind optional und tragen die aufgelöste Verknüpfung
// nur bei `Status == linked`.
type SessionLink struct {
	Status        SessionLinkStatus
	ProjectID     string
	SessionID     string
	CorrelationID string
}

// AnalyzeManifestResult bündelt Analyzer-Ergebnis und Session-Link für
// die Antwort `{analysis, session_link}` aus API-Kontrakt §3.6.
// Driving-Port liefert dieses Resultat ab plan-0.4.0 §4.5; der
// HTTP-Adapter mappt auf die JSON-Hülle.
type AnalyzeManifestResult struct {
	Analysis    StreamAnalysisResult
	SessionLink SessionLink
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
	// StreamAnalysisInvalidInput meldet eine fehlerhafte Aufrufer-
	// Eingabe (kind/text/url-Form). HTTP 400.
	StreamAnalysisInvalidInput StreamAnalysisErrorCode = "invalid_input"
	// StreamAnalysisManifestNotHLS meldet, dass das Manifest kein
	// HLS-Inhalt ist. HTTP 422 — semantisch wohlgeformt, aber
	// inhaltlich nicht verarbeitbar.
	StreamAnalysisManifestNotHLS StreamAnalysisErrorCode = "manifest_not_hls"
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
