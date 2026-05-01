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

// StreamAnalysisResult ist das Domain-Modell der Analyseausgabe. Das
// stabile JSON-Schema entsteht in plan-0.3.0 Tranche 5; bis dahin
// nutzt die API RawJSON, um analyzer-spezifische Detail-Strukturen
// unverändert weiterzureichen, ohne die Domain auf eine HLS-spezi-
// fische Struktur festzulegen.
type StreamAnalysisResult struct {
	// AnalyzerVersion stammt aus packages/stream-analyzer/package.json
	// (plan-0.3.0 §2 Tranche 1: Versionssynchronizität).
	AnalyzerVersion string
	// PlaylistType klassifiziert das Manifest.
	PlaylistType PlaylistType
	// Findings sind die Drei-Stufen-Befunde aus Tranche 4.
	Findings []StreamAnalysisFinding
	// RawJSON ist die vom Analyzer gelieferte JSON-Repräsentation des
	// Ergebnisses, additiv erweiterbar (Tranche 5). Leere Slices
	// signalisieren „kein Detail geliefert".
	RawJSON []byte
}
