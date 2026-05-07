package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// plan-0.9.5 §2 Tranche 1 — SessionsService.ListSessions-Hot-Path-
// Bench für `make api-benchmark-smoke`.
//
// Budget aus `docs/perf/budgets.md` §3 (initial, Tranche-0-Stand):
//   - ListStreamSessions (Default-Limit 100, gefüllte 1k-Session-
//     DB): ≤ 50 ms / Page (Cursor-Decode + Lookup + Domain-
//     Hydratation; gemessen gegen den fakeSessionRepo aus
//     sessions_service_test.go für deterministische Latenz).

// BenchmarkSessionsService_ListSessions_DefaultPage misst eine
// einzelne Page-Anfrage (Default-Limit 100) gegen einen Repo mit
// 1.000 vorbefüllten Sessions. Der `fakeSessionRepo` lebt in
// sessions_service_test.go und implementiert das produktive Sort-
// und Cursor-Verhalten; SQLite-Backend-Bench liegt in
// `adapters/driven/persistence/sqlite/event_repository_bench_test.go`.
func BenchmarkSessionsService_ListSessions_DefaultPage(b *testing.B) {
	repo := newFakeSessionRepo()
	base := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 1_000; i++ {
		startedAt := base.Add(time.Duration(i) * time.Second)
		repo.store[fmt.Sprintf("sess-%04d", i)] = domain.StreamSession{
			ID:          fmt.Sprintf("sess-%04d", i),
			ProjectID:   "demo",
			State:       domain.SessionStateActive,
			StartedAt:   startedAt,
			LastEventAt: startedAt,
			EventCount:  1,
		}
	}
	svc := application.NewSessionsService(repo, &fakeEventRepo{})
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := svc.ListSessions(ctx, driving.ListSessionsInput{ProjectID: "demo"})
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		if len(res.Sessions) == 0 {
			b.Fatalf("expected non-empty page")
		}
	}
}
