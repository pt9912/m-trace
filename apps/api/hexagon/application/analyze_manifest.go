package application

import (
	"context"
	"errors"
	"strings"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// ErrAnalyzeManifestEmpty wird zurückgegeben, wenn weder ManifestText
// noch ManifestURL gesetzt ist; HTTP-Mapping: 400.
var ErrAnalyzeManifestEmpty = errors.New("analyze manifest: weder ManifestText noch ManifestURL gesetzt")

// AnalyzeManifestUseCase orchestriert den Analyzer-Aufruf gegen den
// driven Port und löst optionale Session-Link-Felder gegen das
// SessionRepository auf (plan-0.4.0 §4.5). Die Implementierung selbst
// ist dünn — der eigentliche Lade- und Parse-Aufwand liegt im
// analyzer-service hinter dem HTTPStreamAnalyzer-Adapter (plan-0.3.0
// §7 Tranche 6); die Link-Auflösung läuft project-skopiert über
// `(ProjectID, CorrelationID)` bzw. `(ProjectID, SessionID)`.
type AnalyzeManifestUseCase struct {
	analyzer driven.StreamAnalyzer
	sessions driven.SessionRepository
}

// NewAnalyzeManifestUseCase verdrahtet den Use Case mit den Ports.
// `sessions` darf `nil` sein, wenn Link-Resolution nicht benötigt wird;
// in dem Fall liefert jeder Request mit Link-Feldern
// `SessionLinkStatusNotFoundDetached`. Production-Wiring (`cmd/api`)
// muss immer das echte SessionRepository übergeben.
func NewAnalyzeManifestUseCase(analyzer driven.StreamAnalyzer, sessions driven.SessionRepository) *AnalyzeManifestUseCase {
	return &AnalyzeManifestUseCase{analyzer: analyzer, sessions: sessions}
}

// Compile-time check.
var _ driving.StreamAnalysisInbound = (*AnalyzeManifestUseCase)(nil)

// AnalyzeManifest validiert die Eingabe, delegiert die Manifest-
// Analyse an den Adapter und löst — falls Link-Felder gesetzt sind —
// die Session-Verknüpfung über `(ProjectID, CorrelationID)` /
// `(ProjectID, SessionID)` auf (Statusmatrix aus API-Kontrakt §3.6).
//
// Ein Aufruf ohne Link-Felder bleibt session-los (`detached`); ohne
// `ProjectID` (= ungebundener Request) gilt das auch dann, wenn
// Link-Felder leer sind. Wenn Link-Felder gesetzt sind, MUSS
// `ProjectID` gesetzt sein — der HTTP-Adapter erzwingt das vor dem
// Use-Case-Aufruf (Token-Pflicht aus §4); ein Use-Case-Aufruf mit
// gesetzten Link-Feldern ohne Project ist ein Adapter-Bug und führt
// hier zu `not_found_detached`.
func (u *AnalyzeManifestUseCase) AnalyzeManifest(ctx context.Context, req domain.StreamAnalysisRequest) (domain.AnalyzeManifestResult, error) {
	text := strings.TrimSpace(req.ManifestText)
	url := strings.TrimSpace(req.ManifestURL)
	if text == "" && url == "" {
		return domain.AnalyzeManifestResult{}, ErrAnalyzeManifestEmpty
	}
	analysis, err := u.analyzer.AnalyzeManifest(ctx, req)
	if err != nil {
		return domain.AnalyzeManifestResult{}, err
	}
	link := u.resolveSessionLink(ctx, req)
	return domain.AnalyzeManifestResult{Analysis: analysis, SessionLink: link}, nil
}

// resolveSessionLink implementiert die Statusmatrix aus API-Kontrakt
// §3.6:
//
//   - keine Link-Felder → detached
//   - correlation_id allein, im Project bekannt → linked
//   - correlation_id allein, unbekannt/project-fremd → not_found_detached
//   - session_id allein, im Project bekannt → linked
//   - session_id allein, unbekannt → not_found_detached
//   - beide gesetzt, correlation_id unbekannt/project-fremd → not_found_detached
//     (auch wenn session_id im Project bekannt wäre — correlation_id
//     gewinnt; eine unbekannte cid darf nicht durch einen sid-Fallback
//     "gerettet" werden)
//   - beide gesetzt, correlation_id bekannt, session_id matched → linked
//   - beide gesetzt, correlation_id bekannt, session_id mismatch → conflict_detached
func (u *AnalyzeManifestUseCase) resolveSessionLink(ctx context.Context, req domain.StreamAnalysisRequest) domain.SessionLink {
	if req.CorrelationID == "" && req.SessionID == "" {
		return domain.SessionLink{Status: domain.SessionLinkStatusDetached}
	}
	if u.sessions == nil || req.ProjectID == "" {
		return domain.SessionLink{Status: domain.SessionLinkStatusNotFoundDetached}
	}
	if req.CorrelationID != "" {
		return u.resolveByCorrelationID(ctx, req)
	}
	return u.resolveBySessionID(ctx, req)
}

// resolveByCorrelationID deckt die zwei Branches ab, in denen eine
// `correlation_id` gesetzt ist (mit oder ohne zusätzliche
// `session_id`).
func (u *AnalyzeManifestUseCase) resolveByCorrelationID(ctx context.Context, req domain.StreamAnalysisRequest) domain.SessionLink {
	cidSession, err := u.sessions.GetByCorrelationID(ctx, req.ProjectID, req.CorrelationID)
	if err != nil {
		// ErrSessionNotFound oder Cross-Project-Treffer — beides als
		// not_found_detached behandeln.
		return domain.SessionLink{Status: domain.SessionLinkStatusNotFoundDetached}
	}
	if req.SessionID == "" {
		return domain.SessionLink{
			Status:        domain.SessionLinkStatusLinked,
			ProjectID:     cidSession.ProjectID,
			SessionID:     cidSession.ID,
			CorrelationID: cidSession.CorrelationID,
		}
	}
	if req.SessionID != cidSession.ID {
		return domain.SessionLink{Status: domain.SessionLinkStatusConflictDetached}
	}
	return domain.SessionLink{
		Status:        domain.SessionLinkStatusLinked,
		ProjectID:     cidSession.ProjectID,
		SessionID:     cidSession.ID,
		CorrelationID: cidSession.CorrelationID,
	}
}

// resolveBySessionID wird nur erreicht, wenn `correlation_id` leer und
// `session_id` gesetzt ist (Fallback-Pfad). Eine unbekannte `session_id`
// erzeugt keine neue Session und liefert `not_found_detached`.
func (u *AnalyzeManifestUseCase) resolveBySessionID(ctx context.Context, req domain.StreamAnalysisRequest) domain.SessionLink {
	session, err := u.sessions.Get(ctx, req.ProjectID, req.SessionID)
	if err != nil {
		return domain.SessionLink{Status: domain.SessionLinkStatusNotFoundDetached}
	}
	return domain.SessionLink{
		Status:        domain.SessionLinkStatusLinked,
		ProjectID:     session.ProjectID,
		SessionID:     session.ID,
		CorrelationID: session.CorrelationID,
	}
}
