package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/vukyn/isme/internal/domains/settings/entity"
	"github.com/vukyn/isme/internal/domains/settings/models"

	pkgScheduler "github.com/vukyn/kuery/scheduler"
)

// fakeRepo is an in-memory IRepository keyed by job_key, capturing the last
// UpdateSchedule / RecordScheduleRun calls per job.
type fakeRepo struct {
	configs map[string]entity.ScheduleConfig

	updateCalled  bool
	updateJobKey  string
	updateEnabled bool
	updateCron    string
	updateParams  string
	updateBy      string

	recordCalled bool
	recordJobKey string
	recordResult string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{configs: make(map[string]entity.ScheduleConfig)}
}

func (f *fakeRepo) GetSchedule(ctx context.Context, jobKey string) (entity.ScheduleConfig, error) {
	return f.configs[jobKey], nil
}

func (f *fakeRepo) UpdateSchedule(ctx context.Context, jobKey string, enabled bool, cron string, params string, updatedBy string) error {
	f.updateCalled = true
	f.updateJobKey = jobKey
	f.updateEnabled = enabled
	f.updateCron = cron
	f.updateParams = params
	f.updateBy = updatedBy

	config := f.configs[jobKey]
	config.JobKey = jobKey
	config.Enabled = enabled
	config.Cron = cron
	config.Params = params
	f.configs[jobKey] = config
	return nil
}

func (f *fakeRepo) RecordScheduleRun(ctx context.Context, jobKey string, ranAt time.Time, result string) error {
	f.recordCalled = true
	f.recordJobKey = jobKey
	f.recordResult = result

	config := f.configs[jobKey]
	config.JobKey = jobKey
	config.LastRunAt = &ranAt
	config.LastResult = &result
	f.configs[jobKey] = config
	return nil
}

// fakeReloader records whether Reload was invoked and with what arguments. It
// implements the kuery scheduler.IReloader seam: the schedule is the opaque
// pkgScheduler.Schedule value, compared against pkgScheduler.Cron(expected).
type fakeReloader struct {
	called   bool
	jobKey   pkgScheduler.JobKey
	enabled  bool
	schedule pkgScheduler.Schedule
}

func (f *fakeReloader) Reload(ctx context.Context, jobKey pkgScheduler.JobKey, enabled bool, schedule pkgScheduler.Schedule) error {
	f.called = true
	f.jobKey = jobKey
	f.enabled = enabled
	f.schedule = schedule
	return nil
}

func TestUpdatePersistsAndReloads(t *testing.T) {
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.UpdateRequest{Enabled: true, Cron: "0 3 * * *"}
	if err := uc.Update(context.Background(), req); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if !repo.updateCalled {
		t.Fatal("expected repo.UpdateSchedule to be called")
	}
	if repo.updateJobKey != entity.JobKeySessionRevoke {
		t.Fatalf("repo.UpdateSchedule got jobKey=%q, want %q", repo.updateJobKey, entity.JobKeySessionRevoke)
	}
	if repo.updateEnabled != true || repo.updateCron != "0 3 * * *" {
		t.Fatalf("repo.UpdateSchedule got enabled=%v cron=%q", repo.updateEnabled, repo.updateCron)
	}
	if repo.updateParams != "{}" {
		t.Fatalf("session-revoke params should be empty {}, got %q", repo.updateParams)
	}
	if !reloader.called {
		t.Fatal("expected reloader.Reload to be called")
	}
	if reloader.jobKey != pkgScheduler.JobKey(entity.JobKeySessionRevoke) {
		t.Fatalf("reloader.Reload got jobKey=%q, want %q", reloader.jobKey, entity.JobKeySessionRevoke)
	}
	if reloader.enabled != true || reloader.schedule != pkgScheduler.Cron("0 3 * * *") {
		t.Fatalf("reloader.Reload got enabled=%v schedule=%+v", reloader.enabled, reloader.schedule)
	}
}

func TestUpdateRejectsBadCron(t *testing.T) {
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.UpdateRequest{Enabled: true, Cron: "bogus"}
	if err := uc.Update(context.Background(), req); err == nil {
		t.Fatal("expected validation error for bad cron")
	}
	if repo.updateCalled {
		t.Fatal("repo.UpdateSchedule should not be called when validation fails")
	}
	if reloader.called {
		t.Fatal("reloader.Reload should not be called when validation fails")
	}
}

func TestGetMapsConfig(t *testing.T) {
	ranAt := time.Unix(1700000000, 0).UTC()
	lastResult := `{"revoked":7}`
	repo := newFakeRepo()
	repo.configs[entity.JobKeySessionRevoke] = entity.ScheduleConfig{
		JobKey:     entity.JobKeySessionRevoke,
		Enabled:    true,
		Cron:       "0 4 * * *",
		Params:     "{}",
		LastRunAt:  &ranAt,
		LastResult: &lastResult,
	}
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
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.RotationCleanupUpdateRequest{Enabled: true, Cron: "0 4 * * *", RetentionHours: 48}
	if err := uc.UpdateRotationCleanup(context.Background(), req); err != nil {
		t.Fatalf("UpdateRotationCleanup: %v", err)
	}

	if !repo.updateCalled {
		t.Fatal("expected repo.UpdateSchedule to be called")
	}
	if repo.updateJobKey != entity.JobKeyRotationCleanup {
		t.Fatalf("repo.UpdateSchedule got jobKey=%q, want %q", repo.updateJobKey, entity.JobKeyRotationCleanup)
	}
	if repo.updateEnabled != true || repo.updateCron != "0 4 * * *" {
		t.Fatalf("repo.UpdateSchedule got enabled=%v cron=%q", repo.updateEnabled, repo.updateCron)
	}
	if repo.updateParams != `{"retention_hours":48}` {
		t.Fatalf("rotation-cleanup params mismatch: %q", repo.updateParams)
	}
	if !reloader.called {
		t.Fatal("expected reloader.Reload to be called")
	}
	if reloader.jobKey != pkgScheduler.JobKey(entity.JobKeyRotationCleanup) {
		t.Fatalf("reloader.Reload got jobKey=%q, want %q", reloader.jobKey, entity.JobKeyRotationCleanup)
	}
	if reloader.enabled != true || reloader.schedule != pkgScheduler.Cron("0 4 * * *") {
		t.Fatalf("reloader.Reload got enabled=%v schedule=%+v", reloader.enabled, reloader.schedule)
	}
}

func TestUpdateRotationCleanupRejectsLowRetention(t *testing.T) {
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.RotationCleanupUpdateRequest{Enabled: true, Cron: "0 4 * * *", RetentionHours: 12}
	if err := uc.UpdateRotationCleanup(context.Background(), req); err == nil {
		t.Fatal("expected validation error for retention below 24h")
	}
	if repo.updateCalled {
		t.Fatal("repo.UpdateSchedule should not be called when validation fails")
	}
	if reloader.called {
		t.Fatal("reloader.Reload should not be called when validation fails")
	}
}

func TestGetRotationCleanupMapsConfig(t *testing.T) {
	ranAt := time.Unix(1700000000, 0).UTC()
	lastResult := `{"cleaned":42}`
	repo := newFakeRepo()
	repo.configs[entity.JobKeyRotationCleanup] = entity.ScheduleConfig{
		JobKey:     entity.JobKeyRotationCleanup,
		Enabled:    true,
		Cron:       "0 4 * * *",
		Params:     `{"retention_hours":72}`,
		LastRunAt:  &ranAt,
		LastResult: &lastResult,
	}
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

// TestRotationCleanupRoundTripsParamsThroughJSON verifies a retention value
// written via UpdateRotationCleanup survives the params JSON column and reads
// back identically via GetRotationCleanup (the new generic storage path).
func TestRotationCleanupRoundTripsParamsThroughJSON(t *testing.T) {
	repo := newFakeRepo()
	uc := NewUsecase(repo, &fakeReloader{})

	if err := uc.UpdateRotationCleanup(context.Background(), models.RotationCleanupUpdateRequest{
		Enabled: true, Cron: "0 4 * * *", RetentionHours: 96,
	}); err != nil {
		t.Fatalf("UpdateRotationCleanup: %v", err)
	}

	resp, err := uc.GetRotationCleanup(context.Background())
	if err != nil {
		t.Fatalf("GetRotationCleanup: %v", err)
	}
	if resp.RetentionHours != 96 {
		t.Fatalf("retention did not round-trip through params JSON: got %d, want 96", resp.RetentionHours)
	}
}

func TestUpdateActivityCleanupPersistsAndReloads(t *testing.T) {
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.ActivityCleanupUpdateRequest{Enabled: true, Cron: "0 5 * * *", RetentionDays: 90}
	if err := uc.UpdateActivityCleanup(context.Background(), req); err != nil {
		t.Fatalf("UpdateActivityCleanup: %v", err)
	}

	if !repo.updateCalled {
		t.Fatal("expected repo.UpdateSchedule to be called")
	}
	if repo.updateJobKey != entity.JobKeyActivityCleanup {
		t.Fatalf("repo.UpdateSchedule got jobKey=%q, want %q", repo.updateJobKey, entity.JobKeyActivityCleanup)
	}
	if repo.updateEnabled != true || repo.updateCron != "0 5 * * *" {
		t.Fatalf("repo.UpdateSchedule got enabled=%v cron=%q", repo.updateEnabled, repo.updateCron)
	}
	if repo.updateParams != `{"retention_days":90}` {
		t.Fatalf("activity-cleanup params mismatch: %q", repo.updateParams)
	}
	if !reloader.called {
		t.Fatal("expected reloader.Reload to be called")
	}
	if reloader.jobKey != pkgScheduler.JobKey(entity.JobKeyActivityCleanup) {
		t.Fatalf("reloader.Reload got jobKey=%q, want %q", reloader.jobKey, entity.JobKeyActivityCleanup)
	}
	if reloader.enabled != true || reloader.schedule != pkgScheduler.Cron("0 5 * * *") {
		t.Fatalf("reloader.Reload got enabled=%v schedule=%+v", reloader.enabled, reloader.schedule)
	}
}

func TestUpdateActivityCleanupRejectsLowRetention(t *testing.T) {
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.ActivityCleanupUpdateRequest{Enabled: true, Cron: "0 5 * * *", RetentionDays: 3}
	if err := uc.UpdateActivityCleanup(context.Background(), req); err == nil {
		t.Fatal("expected validation error for retention below 7 days")
	}
	if repo.updateCalled {
		t.Fatal("repo.UpdateSchedule should not be called when validation fails")
	}
	if reloader.called {
		t.Fatal("reloader.Reload should not be called when validation fails")
	}
}

func TestGetActivityCleanupMapsConfig(t *testing.T) {
	ranAt := time.Unix(1700000000, 0).UTC()
	lastResult := `{"pruned":5120}`
	repo := newFakeRepo()
	repo.configs[entity.JobKeyActivityCleanup] = entity.ScheduleConfig{
		JobKey:     entity.JobKeyActivityCleanup,
		Enabled:    true,
		Cron:       "0 5 * * *",
		Params:     `{"retention_days":180}`,
		LastRunAt:  &ranAt,
		LastResult: &lastResult,
	}
	uc := NewUsecase(repo, &fakeReloader{})

	resp, err := uc.GetActivityCleanup(context.Background())
	if err != nil {
		t.Fatalf("GetActivityCleanup: %v", err)
	}
	if resp.Enabled != true || resp.Cron != "0 5 * * *" || resp.RetentionDays != 180 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.LastRunAt == nil || *resp.LastRunAt != ranAt.Unix() {
		t.Fatalf("LastRunAt mismatch: %+v", resp.LastRunAt)
	}
	if resp.LastPrunedCount == nil || *resp.LastPrunedCount != 5120 {
		t.Fatalf("LastPrunedCount mismatch: %+v", resp.LastPrunedCount)
	}
}

// TestActivityCleanupRoundTripsParamsThroughJSON verifies a retention value
// written via UpdateActivityCleanup survives the params JSON column and reads
// back identically via GetActivityCleanup.
func TestActivityCleanupRoundTripsParamsThroughJSON(t *testing.T) {
	repo := newFakeRepo()
	uc := NewUsecase(repo, &fakeReloader{})

	if err := uc.UpdateActivityCleanup(context.Background(), models.ActivityCleanupUpdateRequest{
		Enabled: true, Cron: "0 5 * * *", RetentionDays: 365,
	}); err != nil {
		t.Fatalf("UpdateActivityCleanup: %v", err)
	}

	resp, err := uc.GetActivityCleanup(context.Background())
	if err != nil {
		t.Fatalf("GetActivityCleanup: %v", err)
	}
	if resp.RetentionDays != 365 {
		t.Fatalf("retention did not round-trip through params JSON: got %d, want 365", resp.RetentionDays)
	}
}
