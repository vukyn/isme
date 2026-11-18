package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/vukyn/isme/internal/domains/settings/entity"
	"github.com/vukyn/isme/internal/domains/settings/models"

	pkgScheduler "github.com/vukyn/kuery/scheduler"
)

func TestUpdateInvitationCleanupPersistsAndReloads(t *testing.T) {
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.InvitationCleanupUpdateRequest{Enabled: true, Cron: "0 6 * * *"}
	if err := uc.UpdateInvitationCleanup(context.Background(), req); err != nil {
		t.Fatalf("UpdateInvitationCleanup: %v", err)
	}
	if !repo.updateCalled || repo.updateJobKey != entity.JobKeyInvitationCleanup {
		t.Fatalf("repo.UpdateSchedule got called=%v jobKey=%q", repo.updateCalled, repo.updateJobKey)
	}
	if !repo.updateEnabled || repo.updateCron != "0 6 * * *" || repo.updateParams != "{}" {
		t.Fatalf("repo.UpdateSchedule got enabled=%v cron=%q params=%q", repo.updateEnabled, repo.updateCron, repo.updateParams)
	}
	if !reloader.called || reloader.jobKey != pkgScheduler.JobKey(entity.JobKeyInvitationCleanup) {
		t.Fatalf("reloader.Reload got called=%v jobKey=%q", reloader.called, reloader.jobKey)
	}
	if !reloader.enabled || reloader.schedule != pkgScheduler.Cron("0 6 * * *") {
		t.Fatalf("reloader.Reload got enabled=%v schedule=%+v", reloader.enabled, reloader.schedule)
	}
}

func TestUpdateInvitationCleanupRejectsMissingCron(t *testing.T) {
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	if err := uc.UpdateInvitationCleanup(context.Background(), models.InvitationCleanupUpdateRequest{Enabled: true}); err == nil {
		t.Fatal("expected validation error for an enabled job without cron")
	}
	if repo.updateCalled || reloader.called {
		t.Fatal("nothing should be persisted or reloaded when validation fails")
	}
}

func TestGetInvitationCleanupMapsConfig(t *testing.T) {
	ranAt := time.Unix(1700000000, 0).UTC()
	lastResult := `{"expired":7}`
	repo := newFakeRepo()
	repo.configs[entity.JobKeyInvitationCleanup] = entity.ScheduleConfig{
		JobKey: entity.JobKeyInvitationCleanup, Enabled: true, Cron: "0 6 * * *", Params: "{}",
		LastRunAt: &ranAt, LastResult: &lastResult,
	}
	uc := NewUsecase(repo, &fakeReloader{})

	resp, err := uc.GetInvitationCleanup(context.Background())
	if err != nil {
		t.Fatalf("GetInvitationCleanup: %v", err)
	}
	if !resp.Enabled || resp.Cron != "0 6 * * *" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.LastRunAt == nil || *resp.LastRunAt != ranAt.Unix() {
		t.Fatalf("LastRunAt mismatch: %+v", resp.LastRunAt)
	}
	if resp.LastExpiredCount == nil || *resp.LastExpiredCount != 7 {
		t.Fatalf("LastExpiredCount mismatch: %+v", resp.LastExpiredCount)
	}
}

func TestUpdateUnverifiedUserCleanupPersistsAndReloads(t *testing.T) {
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.UnverifiedUserCleanupUpdateRequest{Enabled: true, Cron: "0 7 * * *", MinAgeDays: 30}
	if err := uc.UpdateUnverifiedUserCleanup(context.Background(), req); err != nil {
		t.Fatalf("UpdateUnverifiedUserCleanup: %v", err)
	}
	if !repo.updateCalled || repo.updateJobKey != entity.JobKeyUnverifiedUserCleanup {
		t.Fatalf("repo.UpdateSchedule got called=%v jobKey=%q", repo.updateCalled, repo.updateJobKey)
	}
	if repo.updateParams != `{"min_age_days":30}` {
		t.Fatalf("unverified-user-cleanup params mismatch: %q", repo.updateParams)
	}
	if !reloader.called || reloader.jobKey != pkgScheduler.JobKey(entity.JobKeyUnverifiedUserCleanup) {
		t.Fatalf("reloader.Reload got called=%v jobKey=%q", reloader.called, reloader.jobKey)
	}
	if !reloader.enabled || reloader.schedule != pkgScheduler.Cron("0 7 * * *") {
		t.Fatalf("reloader.Reload got enabled=%v schedule=%+v", reloader.enabled, reloader.schedule)
	}
}

func TestUpdateUnverifiedUserCleanupRejectsBelowFloor(t *testing.T) {
	repo := newFakeRepo()
	reloader := &fakeReloader{}
	uc := NewUsecase(repo, reloader)

	req := models.UnverifiedUserCleanupUpdateRequest{Enabled: true, Cron: "0 7 * * *", MinAgeDays: 0}
	if err := uc.UpdateUnverifiedUserCleanup(context.Background(), req); err == nil {
		t.Fatal("expected validation error for min age below 1 day")
	}
	if repo.updateCalled || reloader.called {
		t.Fatal("nothing should be persisted or reloaded when validation fails")
	}
}

func TestUnverifiedUserCleanupRoundTripsThroughJSON(t *testing.T) {
	repo := newFakeRepo()
	uc := NewUsecase(repo, &fakeReloader{})

	if err := uc.UpdateUnverifiedUserCleanup(context.Background(), models.UnverifiedUserCleanupUpdateRequest{
		Enabled: true, Cron: "0 7 * * *", MinAgeDays: 45,
	}); err != nil {
		t.Fatalf("UpdateUnverifiedUserCleanup: %v", err)
	}
	if err := repo.RecordScheduleRun(context.Background(), entity.JobKeyUnverifiedUserCleanup, time.Unix(1700000000, 0).UTC(), `{"deleted":3}`); err != nil {
		t.Fatalf("RecordScheduleRun: %v", err)
	}

	resp, err := uc.GetUnverifiedUserCleanup(context.Background())
	if err != nil {
		t.Fatalf("GetUnverifiedUserCleanup: %v", err)
	}
	if resp.MinAgeDays != 45 {
		t.Fatalf("min age did not round-trip through params JSON: got %d, want 45", resp.MinAgeDays)
	}
	if resp.LastDeletedCount == nil || *resp.LastDeletedCount != 3 {
		t.Fatalf("LastDeletedCount mismatch: %+v", resp.LastDeletedCount)
	}
}
