package di

import (
	"context"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/sarulabs/di/v2"
	pkgScheduler "github.com/vukyn/kuery/scheduler"

	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/constants"
	settingsEntity "github.com/vukyn/isme/internal/domains/settings/entity"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	"github.com/vukyn/isme/internal/testdb"
)

func TestRotationCutoff(t *testing.T) {
	now := time.Date(2026, 6, 11, 4, 0, 0, 0, time.UTC)
	cases := []struct {
		name           string
		retentionHours int64
		want           time.Time
	}{
		{"48h", 48, now.Add(-48 * time.Hour)},
		{"24h floor", 24, now.Add(-24 * time.Hour)},
		{"7d", 168, now.Add(-168 * time.Hour)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rotationCutoff(now, tc.retentionHours); !got.Equal(tc.want) {
				t.Fatalf("rotationCutoff(%v) = %v, want %v", tc.retentionHours, got, tc.want)
			}
		})
	}
}

func TestActivityCutoff(t *testing.T) {
	now := time.Date(2026, 6, 11, 5, 0, 0, 0, time.UTC)
	cases := []struct {
		name          string
		retentionDays int64
		want          time.Time
	}{
		{"30d", 30, now.Add(-30 * 24 * time.Hour)},
		{"7d floor", 7, now.Add(-7 * 24 * time.Hour)},
		{"90d default", 90, now.Add(-90 * 24 * time.Hour)},
		{"365d", 365, now.Add(-365 * 24 * time.Hour)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := activityCutoff(now, tc.retentionDays); !got.Equal(tc.want) {
				t.Fatalf("activityCutoff(%v) = %v, want %v", tc.retentionDays, got, tc.want)
			}
		})
	}
}

// recordingProvider records every job key the engine asks about at Start and
// reports each one disabled, so nothing is installed.
type recordingProvider struct {
	keys []string
}

func (p *recordingProvider) Load(_ context.Context, key pkgScheduler.JobKey) (bool, pkgScheduler.Schedule, error) {
	p.keys = append(p.keys, string(key))
	return false, pkgScheduler.Schedule{}, nil
}

// The seeded schedule_config rows and the jobs the scheduler registers must be
// the SAME set: a row with no job is dead configuration the UI can still edit,
// and a job with no row can never be enabled.
func TestSeededJobsEqualRegisteredJobs(t *testing.T) {
	db := testdb.Open(t)

	builder, err := di.NewBuilder()
	if err != nil {
		t.Fatalf("new builder: %v", err)
	}
	cfg := &config.Config{}
	cfg.Scheduler.Enabled = true
	if err := builder.Add(
		di.Def{Name: constants.CONTAINER_NAME_CONFIG, Scope: di.App, Build: func(di.Container) (any, error) { return cfg, nil }},
		di.Def{Name: constants.CONTAINER_NAME_DB, Scope: di.App, Build: func(di.Container) (any, error) { return db, nil }},
		*defineScheduler(),
	); err != nil {
		t.Fatalf("add definitions: %v", err)
	}
	container := builder.Build()
	t.Cleanup(func() { _ = container.DeleteWithSubContainers() })

	engine := container.Get(constants.CONTAINER_NAME_SCHEDULER).(*pkgScheduler.Engine)
	provider := &recordingProvider{}
	engine.Start(context.Background(), provider)
	registered := provider.keys
	sort.Strings(registered)

	var seeded []string
	if err := db.NewSelect().Table("schedule_config").Column("job_key").Order("job_key").Scan(context.Background(), &seeded); err != nil {
		t.Fatalf("read schedule_config: %v", err)
	}

	if !slices.Equal(seeded, registered) {
		t.Fatalf("seeded job rows %v != registered jobs %v", seeded, registered)
	}
}

// With both configs seeded disabled (the migration default), the schedule
// provider reports every job as not-enabled, so the engine installs no job at
// Start. This is the parity check for the old "both disabled => 0 jobs" test.
func TestScheduleProviderReportsSeededJobsDisabled(t *testing.T) {
	db := testdb.Open(t)
	provider := newScheduleProvider(settingsRepo.NewRepository(db))

	for _, jobKey := range []pkgScheduler.JobKey{
		pkgScheduler.JobKey(settingsEntity.JobKeySessionRevoke),
		pkgScheduler.JobKey(settingsEntity.JobKeyRotationCleanup),
		pkgScheduler.JobKey(settingsEntity.JobKeyActivityCleanup),
		pkgScheduler.JobKey(settingsEntity.JobKeyInvitationCleanup),
		pkgScheduler.JobKey(settingsEntity.JobKeyUnverifiedUserCleanup),
	} {
		enabled, _, err := provider.Load(context.Background(), jobKey)
		if err != nil {
			t.Fatalf("provider.Load(%q): %v", jobKey, err)
		}
		if enabled {
			t.Fatalf("expected job %q seeded disabled, got enabled", jobKey)
		}
	}
}
