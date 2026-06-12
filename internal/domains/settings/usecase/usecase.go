package usecase

import (
	"context"
	"encoding/json"

	"github.com/vukyn/isme/internal/domains/settings/entity"
	"github.com/vukyn/isme/internal/domains/settings/models"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"

	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"
	pkgScheduler "github.com/vukyn/kuery/scheduler"
)

// emptyParams is the params JSON for jobs with no job-specific config (revoke).
const emptyParams = "{}"

// sessionRevokeResult mirrors the last_result JSON of the session-revoke job.
type sessionRevokeResult struct {
	Revoked int64 `json:"revoked"`
}

// rotationCleanupParams mirrors the params JSON of the rotation-cleanup job.
type rotationCleanupParams struct {
	RetentionHours int64 `json:"retention_hours"`
}

// rotationCleanupResult mirrors the last_result JSON of the rotation-cleanup job.
type rotationCleanupResult struct {
	Cleaned int64 `json:"cleaned"`
}

// activityCleanupParams mirrors the params JSON of the activity-cleanup job.
type activityCleanupParams struct {
	RetentionDays int64 `json:"retention_days"`
}

// activityCleanupResult mirrors the last_result JSON of the activity-cleanup job.
type activityCleanupResult struct {
	Pruned int64 `json:"pruned"`
}

type usecase struct {
	settingsRepo settingsRepo.IRepository
	reloader     pkgScheduler.IReloader
}

func NewUsecase(
	settingsRepo settingsRepo.IRepository,
	reloader pkgScheduler.IReloader,
) IUseCase {
	return &usecase{
		settingsRepo: settingsRepo,
		reloader:     reloader,
	}
}

func (u *usecase) Get(ctx context.Context) (models.GetResponse, error) {
	config, err := u.settingsRepo.GetSchedule(ctx, entity.JobKeySessionRevoke)
	if err != nil {
		return models.GetResponse{}, err
	}

	response := models.GetResponse{
		Enabled: config.Enabled,
		Cron:    config.Cron,
	}
	if config.LastRunAt != nil {
		lastRun := config.LastRunAt.Unix()
		response.LastRunAt = &lastRun
	}
	if config.LastResult != nil {
		result := sessionRevokeResult{}
		if err := json.Unmarshal([]byte(*config.LastResult), &result); err != nil {
			return models.GetResponse{}, pkgErr.InternalServerError(err.Error())
		}
		revoked := result.Revoked
		response.LastRevokedCount = &revoked
	}
	return response, nil
}

func (u *usecase) Update(ctx context.Context, req models.UpdateRequest) error {
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	updatedBy := pkgCtx.GetUserID(ctx)
	if err := u.settingsRepo.UpdateSchedule(ctx, entity.JobKeySessionRevoke, req.Enabled, req.Cron, emptyParams, updatedBy); err != nil {
		return err
	}

	// live-reload the scheduler so the change takes effect without a restart
	return u.reloader.Reload(ctx, pkgScheduler.JobKey(entity.JobKeySessionRevoke), req.Enabled, pkgScheduler.Cron(req.Cron))
}

func (u *usecase) GetRotationCleanup(ctx context.Context) (models.RotationCleanupGetResponse, error) {
	config, err := u.settingsRepo.GetSchedule(ctx, entity.JobKeyRotationCleanup)
	if err != nil {
		return models.RotationCleanupGetResponse{}, err
	}

	response := models.RotationCleanupGetResponse{
		Enabled: config.Enabled,
		Cron:    config.Cron,
	}
	if config.Params != "" {
		params := rotationCleanupParams{}
		if err := json.Unmarshal([]byte(config.Params), &params); err != nil {
			return models.RotationCleanupGetResponse{}, pkgErr.InternalServerError(err.Error())
		}
		response.RetentionHours = params.RetentionHours
	}
	if config.LastRunAt != nil {
		lastRun := config.LastRunAt.Unix()
		response.LastRunAt = &lastRun
	}
	if config.LastResult != nil {
		result := rotationCleanupResult{}
		if err := json.Unmarshal([]byte(*config.LastResult), &result); err != nil {
			return models.RotationCleanupGetResponse{}, pkgErr.InternalServerError(err.Error())
		}
		cleaned := result.Cleaned
		response.LastCleanedCount = &cleaned
	}
	return response, nil
}

func (u *usecase) UpdateRotationCleanup(ctx context.Context, req models.RotationCleanupUpdateRequest) error {
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	params, err := json.Marshal(rotationCleanupParams{RetentionHours: req.RetentionHours})
	if err != nil {
		return pkgErr.InternalServerError(err.Error())
	}

	updatedBy := pkgCtx.GetUserID(ctx)
	if err := u.settingsRepo.UpdateSchedule(ctx, entity.JobKeyRotationCleanup, req.Enabled, req.Cron, string(params), updatedBy); err != nil {
		return err
	}

	// live-reload the scheduler so the change takes effect without a restart.
	// A retention-only change still reloads here, but retention is read fresh on
	// each run regardless, so it would also take effect on the next run.
	return u.reloader.Reload(ctx, pkgScheduler.JobKey(entity.JobKeyRotationCleanup), req.Enabled, pkgScheduler.Cron(req.Cron))
}

func (u *usecase) GetActivityCleanup(ctx context.Context) (models.ActivityCleanupGetResponse, error) {
	config, err := u.settingsRepo.GetSchedule(ctx, entity.JobKeyActivityCleanup)
	if err != nil {
		return models.ActivityCleanupGetResponse{}, err
	}

	response := models.ActivityCleanupGetResponse{
		Enabled: config.Enabled,
		Cron:    config.Cron,
	}
	if config.Params != "" {
		params := activityCleanupParams{}
		if err := json.Unmarshal([]byte(config.Params), &params); err != nil {
			return models.ActivityCleanupGetResponse{}, pkgErr.InternalServerError(err.Error())
		}
		response.RetentionDays = params.RetentionDays
	}
	if config.LastRunAt != nil {
		lastRun := config.LastRunAt.Unix()
		response.LastRunAt = &lastRun
	}
	if config.LastResult != nil {
		result := activityCleanupResult{}
		if err := json.Unmarshal([]byte(*config.LastResult), &result); err != nil {
			return models.ActivityCleanupGetResponse{}, pkgErr.InternalServerError(err.Error())
		}
		pruned := result.Pruned
		response.LastPrunedCount = &pruned
	}
	return response, nil
}

func (u *usecase) UpdateActivityCleanup(ctx context.Context, req models.ActivityCleanupUpdateRequest) error {
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	params, err := json.Marshal(activityCleanupParams{RetentionDays: req.RetentionDays})
	if err != nil {
		return pkgErr.InternalServerError(err.Error())
	}

	updatedBy := pkgCtx.GetUserID(ctx)
	if err := u.settingsRepo.UpdateSchedule(ctx, entity.JobKeyActivityCleanup, req.Enabled, req.Cron, string(params), updatedBy); err != nil {
		return err
	}

	// live-reload the scheduler so the change takes effect without a restart.
	// A retention-only change still reloads here, but retention is read fresh on
	// each run regardless, so it would also take effect on the next run.
	return u.reloader.Reload(ctx, pkgScheduler.JobKey(entity.JobKeyActivityCleanup), req.Enabled, pkgScheduler.Cron(req.Cron))
}
