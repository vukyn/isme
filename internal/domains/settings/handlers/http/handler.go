package handlers

import (
	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/settings/models"

	pkgCtx "github.com/vukyn/kuery/ctxv3"
	pkgErr "github.com/vukyn/kuery/http/errors"
	pkgHttp "github.com/vukyn/kuery/http/fiberv3"

	"github.com/gofiber/fiber/v3"
)

func GetSessionRevokeConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

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

func UpdateSessionRevokeConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.UpdateRequest{}
	if err := c.Bind().Body(&updateRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	if err := uc.Update(pkgCtx.NewContextFromFiberCtx(c), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func GetRotationCleanupConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

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

func UpdateRotationCleanupConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.RotationCleanupUpdateRequest{}
	if err := c.Bind().Body(&updateRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	if err := uc.UpdateRotationCleanup(pkgCtx.NewContextFromFiberCtx(c), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func GetActivityCleanupConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

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

func UpdateActivityCleanupConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.ActivityCleanupUpdateRequest{}
	if err := c.Bind().Body(&updateRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	if err := uc.UpdateActivityCleanup(pkgCtx.NewContextFromFiberCtx(c), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func GetInvitationCleanupConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	getResponse, err := uc.GetInvitationCleanup(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, getResponse)
}

func UpdateInvitationCleanupConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.InvitationCleanupUpdateRequest{}
	if err := c.Bind().Body(&updateRequest); err != nil {
		// a malformed body is the client's fault, not a server error
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	if err := uc.UpdateInvitationCleanup(pkgCtx.NewContextFromFiberCtx(c), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func GetUnverifiedUserCleanupConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	getResponse, err := uc.GetUnverifiedUserCleanup(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, getResponse)
}

func UpdateUnverifiedUserCleanupConfig(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetSettingsUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.UnverifiedUserCleanupUpdateRequest{}
	if err := c.Bind().Body(&updateRequest); err != nil {
		// a malformed body is the client's fault, not a server error
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	if err := uc.UpdateUnverifiedUserCleanup(pkgCtx.NewContextFromFiberCtx(c), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}
