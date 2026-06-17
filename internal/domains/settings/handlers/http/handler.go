package handlers

import (
	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/settings/models"

	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgHttp "github.com/vukyn/kuery/http/fiber"

	"github.com/gofiber/fiber/v2"
)

func GetSessionRevokeConfig(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	getResponse, err := uc.Get(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, getResponse)
}

func UpdateSessionRevokeConfig(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.UpdateRequest{}
	if err := c.BodyParser(&updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.Update(pkgCtx.NewContextFromFiberCtx(c), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func GetRotationCleanupConfig(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	getResponse, err := uc.GetRotationCleanup(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, getResponse)
}

func UpdateRotationCleanupConfig(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.RotationCleanupUpdateRequest{}
	if err := c.BodyParser(&updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.UpdateRotationCleanup(pkgCtx.NewContextFromFiberCtx(c), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func GetActivityCleanupConfig(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	getResponse, err := uc.GetActivityCleanup(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, getResponse)
}

func UpdateActivityCleanupConfig(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.ActivityCleanupUpdateRequest{}
	if err := c.BodyParser(&updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.UpdateActivityCleanup(pkgCtx.NewContextFromFiberCtx(c), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func GetDatabaseBackupConfig(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	getResponse, err := uc.GetDatabaseBackup(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, getResponse)
}

func UpdateDatabaseBackupConfig(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.DatabaseBackupUpdateRequest{}
	if err := c.BodyParser(&updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.UpdateDatabaseBackup(pkgCtx.NewContextFromFiberCtx(c), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}
