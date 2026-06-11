package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/settings/models"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	"github.com/vukyn/isme/internal/scheduler"

	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"
)

type usecase struct {
	settingsRepo settingsRepo.IRepository
	reloader     scheduler.IReloader
}

func NewUsecase(
	settingsRepo settingsRepo.IRepository,
	reloader scheduler.IReloader,
) IUseCase {
	return &usecase{
		settingsRepo: settingsRepo,
		reloader:     reloader,
	}
}

func (u *usecase) Get(ctx context.Context) (models.GetResponse, error) {
	config, err := u.settingsRepo.Get(ctx)
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
	response.LastRevokedCount = config.LastRevokedCount
	return response, nil
}

func (u *usecase) Update(ctx context.Context, req models.UpdateRequest) error {
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	updatedBy := pkgCtx.GetUserID(ctx)
	if err := u.settingsRepo.Update(ctx, req.Enabled, req.Cron, updatedBy); err != nil {
		return err
	}

	// live-reload the scheduler so the change takes effect without a restart
	return u.reloader.Reload(ctx, scheduler.JobSessionRevoke, req.Enabled, req.Cron)
}

func (u *usecase) GetRotationCleanup(ctx context.Context) (models.RotationCleanupGetResponse, error) {
	config, err := u.settingsRepo.GetRotationCleanup(ctx)
	if err != nil {
		return models.RotationCleanupGetResponse{}, err
	}

	response := models.RotationCleanupGetResponse{
		Enabled:        config.Enabled,
		Cron:           config.Cron,
		RetentionHours: config.RetentionHours,
	}
	if config.LastRunAt != nil {
		lastRun := config.LastRunAt.Unix()
		response.LastRunAt = &lastRun
	}
	response.LastCleanedCount = config.LastCleanedCount
	return response, nil
}

func (u *usecase) UpdateRotationCleanup(ctx context.Context, req models.RotationCleanupUpdateRequest) error {
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	updatedBy := pkgCtx.GetUserID(ctx)
	if err := u.settingsRepo.UpdateRotationCleanup(ctx, req.Enabled, req.Cron, req.RetentionHours, updatedBy); err != nil {
		return err
	}

	// live-reload the scheduler so the change takes effect without a restart.
	// A retention-only change still reloads here, but retention is read fresh on
	// each run regardless, so it would also take effect on the next run.
	return u.reloader.Reload(ctx, scheduler.JobRotationCleanup, req.Enabled, req.Cron)
}
