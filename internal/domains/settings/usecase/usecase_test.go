package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/vukyn/isme/internal/domains/settings/entity"
	"github.com/vukyn/isme/internal/domains/settings/models"
	"github.com/vukyn/isme/internal/scheduler"
)

// fakeRepo is an in-memory IRepository capturing the last Update calls.
type fakeRepo struct {
	config        entity.SessionRevokeConfig
	updateEnabled bool
	updateCron    string
	updateBy      string
	updateCalled  bool
	recordRevoked int64
	recordCalled  bool

	cleanupConfig        entity.RotationCleanupConfig
	cleanupUpdateEnabled bool
	cleanupUpdateCron    string
	cleanupUpdateHours   int64
	cleanupUpdateBy      string
	cleanupUpdateCalled  bool
	cleanupRecordCleaned int64
	cleanupRecordCalled  bool
}

func (f *fakeRepo) Get(ctx context.Context) (entity.SessionRevokeConfig, error) {
	return f.config, nil
}

func (f *fakeRepo) Update(ctx context.Context, enabled bool, cron string, updatedBy string) error {
	f.updateCalled = true
	f.updateEnabled = enabled
	f.updateCron = cron
	f.updateBy = updatedBy
	f.config.Enabled = enabled
	f.config.Cron = cron
	return nil
}

func (f *fakeRepo) RecordRun(ctx context.Context, ranAt time.Time, revoked int64) error {
	f.recordCalled = true
	f.recordRevoked = revoked
	return nil
}

func (f *fakeRepo) GetRotationCleanup(ctx context.Context) (entity.RotationCleanupConfig, error) {
	return f.cleanupConfig, nil
}

func (f *fakeRepo) UpdateRotationCleanup(ctx context.Context, enabled bool, cron string, retentionHours int64, updatedBy string) error {
	f.cleanupUpdateCalled = true
	f.cleanupUpdateEnabled = enabled
	f.cleanupUpdateCron = cron
	f.cleanupUpdateHours = retentionHours
	f.cleanupUpdateBy = updatedBy
	f.cleanupConfig.Enabled = enabled
	f.cleanupConfig.Cron = cron
	f.cleanupConfig.RetentionHours = retentionHours
	return nil
}

func (f *fakeRepo) RecordRotationCleanupRun(ctx context.Context, ranAt time.Time, cleaned int64) error {
	f.cleanupRecordCalled = true
	f.cleanupRecordCleaned = cleaned
	return nil
}

// fakeReloader records whether Reload was invoked and with what arguments.
type fakeReloader struct {
	called  bool
	jobKey  scheduler.JobKey
	enabled bool
	cron    string
}

func (f *fakeReloader) Reload(ctx context.Context, jobKey scheduler.JobKey, enabled bool, cronExpr string) error {
	f.called = true
	f.jobKey = jobKey
	f.enabled = enabled
	f.cron = cronExpr
	return nil
}

func TestUpdatePersistsAndReloads(t *testing.T) {
	repo := &fakeRepo{}
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.UpdateRequest{Enabled: true, Cron: "0 3 * * *"}
	if err := uc.Update(context.Background(), req); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if !repo.updateCalled {
		t.Fatal("expected repo.Update to be called")
	}
	if repo.updateEnabled != true || repo.updateCron != "0 3 * * *" {
		t.Fatalf("repo.Update got enabled=%v cron=%q", repo.updateEnabled, repo.updateCron)
	}
	if !reloader.called {
		t.Fatal("expected reloader.Reload to be called")
	}
	if reloader.jobKey != scheduler.JobSessionRevoke {
		t.Fatalf("reloader.Reload got jobKey=%q, want %q", reloader.jobKey, scheduler.JobSessionRevoke)
	}
	if reloader.enabled != true || reloader.cron != "0 3 * * *" {
		t.Fatalf("reloader.Reload got enabled=%v cron=%q", reloader.enabled, reloader.cron)
	}
}

func TestUpdateRejectsBadCron(t *testing.T) {
	repo := &fakeRepo{}
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.UpdateRequest{Enabled: true, Cron: "bogus"}
	if err := uc.Update(context.Background(), req); err == nil {
		t.Fatal("expected validation error for bad cron")
	}
	if repo.updateCalled {
		t.Fatal("repo.Update should not be called when validation fails")
	}
	if reloader.called {
		t.Fatal("reloader.Reload should not be called when validation fails")
	}
}

func TestGetMapsConfig(t *testing.T) {
	ranAt := time.Unix(1700000000, 0).UTC()
	revoked := int64(7)
	repo := &fakeRepo{config: entity.SessionRevokeConfig{
		Enabled:          true,
		Cron:             "0 4 * * *",
		LastRunAt:        &ranAt,
		LastRevokedCount: &revoked,
	}}
	uc := NewUsecase(repo, &fakeReloader{})

	resp, err := uc.Get(context.Background())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if resp.Enabled != true || resp.Cron != "0 4 * * *" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.LastRunAt == nil || *resp.LastRunAt != ranAt.Unix() {
		t.Fatalf("LastRunAt mismatch: %+v", resp.LastRunAt)
	}
	if resp.LastRevokedCount == nil || *resp.LastRevokedCount != 7 {
		t.Fatalf("LastRevokedCount mismatch: %+v", resp.LastRevokedCount)
	}
}

func TestUpdateRotationCleanupPersistsAndReloads(t *testing.T) {
	repo := &fakeRepo{}
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.RotationCleanupUpdateRequest{Enabled: true, Cron: "0 4 * * *", RetentionHours: 48}
	if err := uc.UpdateRotationCleanup(context.Background(), req); err != nil {
		t.Fatalf("UpdateRotationCleanup: %v", err)
	}

	if !repo.cleanupUpdateCalled {
		t.Fatal("expected repo.UpdateRotationCleanup to be called")
	}
	if repo.cleanupUpdateEnabled != true || repo.cleanupUpdateCron != "0 4 * * *" || repo.cleanupUpdateHours != 48 {
		t.Fatalf("repo.UpdateRotationCleanup got enabled=%v cron=%q hours=%d",
			repo.cleanupUpdateEnabled, repo.cleanupUpdateCron, repo.cleanupUpdateHours)
	}
	if !reloader.called {
		t.Fatal("expected reloader.Reload to be called")
	}
	if reloader.jobKey != scheduler.JobRotationCleanup {
		t.Fatalf("reloader.Reload got jobKey=%q, want %q", reloader.jobKey, scheduler.JobRotationCleanup)
	}
	if reloader.enabled != true || reloader.cron != "0 4 * * *" {
		t.Fatalf("reloader.Reload got enabled=%v cron=%q", reloader.enabled, reloader.cron)
	}
}

func TestUpdateRotationCleanupRejectsLowRetention(t *testing.T) {
	repo := &fakeRepo{}
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.RotationCleanupUpdateRequest{Enabled: true, Cron: "0 4 * * *", RetentionHours: 12}
	if err := uc.UpdateRotationCleanup(context.Background(), req); err == nil {
		t.Fatal("expected validation error for retention below 24h")
	}
	if repo.cleanupUpdateCalled {
		t.Fatal("repo.UpdateRotationCleanup should not be called when validation fails")
	}
	if reloader.called {
		t.Fatal("reloader.Reload should not be called when validation fails")
	}
}

func TestGetRotationCleanupMapsConfig(t *testing.T) {
	ranAt := time.Unix(1700000000, 0).UTC()
	cleaned := int64(42)
	repo := &fakeRepo{cleanupConfig: entity.RotationCleanupConfig{
		Enabled:          true,
		Cron:             "0 4 * * *",
		RetentionHours:   72,
		LastRunAt:        &ranAt,
		LastCleanedCount: &cleaned,
	}}
	uc := NewUsecase(repo, &fakeReloader{})

	resp, err := uc.GetRotationCleanup(context.Background())
	if err != nil {
		t.Fatalf("GetRotationCleanup: %v", err)
	}
	if resp.Enabled != true || resp.Cron != "0 4 * * *" || resp.RetentionHours != 72 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.LastRunAt == nil || *resp.LastRunAt != ranAt.Unix() {
		t.Fatalf("LastRunAt mismatch: %+v", resp.LastRunAt)
	}
	if resp.LastCleanedCount == nil || *resp.LastCleanedCount != 42 {
		t.Fatalf("LastCleanedCount mismatch: %+v", resp.LastCleanedCount)
	}
}
