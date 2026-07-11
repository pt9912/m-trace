package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/postgres"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestIngestStreamRepository_PgLab deckt den Postgres-ingest_stream-
// Adapter gegen echte PG ab: CreateStream (Happy-Path + Endpoint/Target-
// NotFound aus den Vorab-Checks), den zentralen Dialekt-Punkt
// (display_name-Konflikt → ErrIngestStreamNameConflict via SQLSTATE
// 23505 auf idx_ingest_streams_active_display_name), GetStreamByID inkl.
// aktivem Key + Cross-Project-Leak-Schutz, ListByProject-Sortierung,
// RotateKey (Deaktivierungs-Audit-Trail + neuer aktiver Key),
// GetRoutingRuleByID und AppendLifecycleEvent. Gated über MTRACE_PG_LAB_DSN.
func TestIngestStreamRepository_PgLab(t *testing.T) {
	dsn := os.Getenv("MTRACE_PG_LAB_DSN")
	if dsn == "" {
		t.Skip("MTRACE_PG_LAB_DSN nicht gesetzt — PG-Lab-Integrationstest übersprungen")
	}
	ctx := context.Background()
	db, err := storage.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("OpenPostgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := postgres.NewIngestStreamRepository(db)
	seedFixtures(ctx, t, repo)

	base := time.Date(2026, 7, 11, 9, 0, 0, 0, time.UTC)
	const proj = "ing-lab-proj"

	stream, err := repo.CreateStream(ctx, mkPgInput(proj, "Lab Stream", mkPgKey(t, base)))
	if err != nil {
		t.Fatalf("CreateStream: %v", err)
	}
	if stream.ProjectID != proj || stream.Status != domain.IngestStreamStatusReady {
		t.Fatalf("CreateStream = %+v, want project %s / status ready", stream, proj)
	}
	if stream.Key.Fingerprint == "" || stream.RoutingRuleID == "" {
		t.Fatalf("CreateStream: erwartet gesetzten Key-Fingerprint + RoutingRuleID, got %+v", stream)
	}

	t.Run("missing endpoint / target aus Vorab-Check", func(t *testing.T) {
		badEndpoint := mkPgInput(proj, "Bad EP", mkPgKey(t, base))
		badEndpoint.EndpointID = "ep-missing"
		if _, err := repo.CreateStream(ctx, badEndpoint); !errors.Is(err, domain.ErrIngestEndpointNotFound) {
			t.Errorf("CreateStream(bad endpoint) err = %v, want ErrIngestEndpointNotFound", err)
		}
		badTarget := mkPgInput(proj, "Bad TGT", mkPgKey(t, base))
		badTarget.TargetID = "tgt-missing"
		if _, err := repo.CreateStream(ctx, badTarget); !errors.Is(err, domain.ErrIngestTargetNotFound) {
			t.Errorf("CreateStream(bad target) err = %v, want ErrIngestTargetNotFound", err)
		}
	})

	t.Run("display_name-Konflikt → ErrIngestStreamNameConflict (SQLSTATE 23505)", func(t *testing.T) {
		_, err := repo.CreateStream(ctx, mkPgInput(proj, "Lab Stream", mkPgKey(t, base.Add(time.Minute))))
		if !errors.Is(err, domain.ErrIngestStreamNameConflict) {
			t.Errorf("CreateStream(dup display_name) err = %v, want ErrIngestStreamNameConflict", err)
		}
	})

	t.Run("GetStreamByID inkl. Key + Cross-Project-Leak-Schutz", func(t *testing.T) {
		got, err := repo.GetStreamByID(ctx, proj, stream.ID)
		if err != nil {
			t.Fatalf("GetStreamByID: %v", err)
		}
		if got.Key.Fingerprint != stream.Key.Fingerprint {
			t.Errorf("GetStreamByID: Key-Fingerprint = %q, want %q", got.Key.Fingerprint, stream.Key.Fingerprint)
		}
		if _, err := repo.GetStreamByID(ctx, "other-proj", stream.ID); !errors.Is(err, domain.ErrIngestStreamNotFound) {
			t.Errorf("GetStreamByID(cross-project) err = %v, want ErrIngestStreamNotFound", err)
		}
	})

	t.Run("ListByProject", func(t *testing.T) {
		list, err := repo.ListByProject(ctx, proj)
		if err != nil {
			t.Fatalf("ListByProject: %v", err)
		}
		if len(list) != 1 || list[0].ID != stream.ID {
			t.Errorf("ListByProject = %d Einträge (%+v), want genau %s", len(list), list, stream.ID)
		}
	})

	t.Run("RotateKey deaktiviert alten Key + setzt neuen aktiven", func(t *testing.T) {
		oldFingerprint := stream.Key.Fingerprint
		newKey := mkPgKey(t, base.Add(time.Hour))
		rotated, err := repo.RotateKey(ctx, proj, stream.ID, newKey)
		if err != nil {
			t.Fatalf("RotateKey: %v", err)
		}
		if rotated.Key.Fingerprint == oldFingerprint {
			t.Errorf("RotateKey: Fingerprint unverändert (%q) — Rotation wirkungslos", oldFingerprint)
		}
		active, err := repo.FindActiveStreamKey(ctx, proj, stream.ID)
		if err != nil {
			t.Fatalf("FindActiveStreamKey nach Rotation: %v", err)
		}
		if active.Fingerprint != newKey.Fingerprint {
			t.Errorf("aktiver Key nach Rotation = %q, want %q (genau ein aktiver Key)", active.Fingerprint, newKey.Fingerprint)
		}
	})

	t.Run("GetRoutingRuleByID", func(t *testing.T) {
		rule, err := repo.GetRoutingRuleByID(ctx, proj, stream.ID)
		if err != nil {
			t.Fatalf("GetRoutingRuleByID: %v", err)
		}
		if rule.ID != stream.RoutingRuleID || !rule.Enabled || rule.Mode != domain.RoutingRuleModeSingle {
			t.Errorf("GetRoutingRuleByID = %+v, want rule %s / enabled / single", rule, stream.RoutingRuleID)
		}
	})

	t.Run("AppendLifecycleEvent", func(t *testing.T) {
		evt := domain.StreamLifecycleEvent{
			EventID:        "lc-lab-1",
			Kind:           domain.StreamLifecycleEventStarted,
			StreamID:       stream.ID,
			ProjectID:      proj,
			OccurredAt:     base.Add(2 * time.Hour),
			Source:         domain.StreamLifecycleSourceSmoke,
			KeyFingerprint: stream.Key.Fingerprint,
			ConnectionID:   "conn-1",
			Reason:         "",
		}
		if err := repo.AppendLifecycleEvent(ctx, evt); err != nil {
			t.Fatalf("AppendLifecycleEvent: %v", err)
		}
		// Append gegen einen fremden Stream muss den Cross-Project-/
		// Nichtexistenz-Schutz greifen (ErrIngestStreamNotFound).
		missing := evt
		missing.StreamID = "ing_missing"
		missing.EventID = "lc-lab-2"
		if err := repo.AppendLifecycleEvent(ctx, missing); !errors.Is(err, domain.ErrIngestStreamNotFound) {
			t.Errorf("AppendLifecycleEvent(missing stream) err = %v, want ErrIngestStreamNotFound", err)
		}
	})
}

func seedFixtures(ctx context.Context, t *testing.T, repo *postgres.IngestStreamRepository) {
	t.Helper()
	if err := repo.SeedEndpoint(ctx, domain.IngestEndpoint{
		ID:            "ep-srt",
		Protocol:      domain.IngestProtocolSRT,
		ListenHost:    "127.0.0.1",
		ListenPort:    8890,
		PathTemplate:  "publish:{stream_path}",
		LabStack:      "mtrace-srt",
		PublicURLHint: "srt://localhost:8890",
	}); err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}
	if err := repo.SeedTarget(ctx, domain.MediaServerTarget{
		ID:             "tgt-mediamtx",
		Kind:           domain.MediaServerKindMediaMTX,
		ConfigPath:     "examples/ingest-control/mediamtx.generated.yml",
		HLSURLTemplate: "http://localhost:8889/{stream_path}/index.m3u8",
		ControlAPIURL:  "",
	}); err != nil {
		t.Fatalf("seed target: %v", err)
	}
}

func mkPgKey(t *testing.T, when time.Time) domain.StreamKey {
	t.Helper()
	material, err := domain.GenerateStreamKey(when)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return material.ToPersistable()
}

func mkPgInput(projectID, displayName string, key domain.StreamKey) driven.CreateStreamInput {
	return driven.CreateStreamInput{
		ProjectID:   projectID,
		DisplayName: displayName,
		Protocol:    domain.IngestProtocolSRT,
		EndpointID:  "ep-srt",
		TargetID:    "tgt-mediamtx",
		InitialKey:  key,
		CreatedAt:   key.CreatedAt,
	}
}
