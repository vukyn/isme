package handlers

import (
	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/app_service/models"
	pkgCtx "github.com/vukyn/kuery/ctxv3"
	pkgErr "github.com/vukyn/kuery/http/errors"
	pkgHttp "github.com/vukyn/kuery/http/fiberv3"

	"github.com/gofiber/fiber/v3"
)

func RegisterApp(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	registerRequest := models.RegisterRequest{}
	if err := c.Bind().Body(&registerRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	registerResponse, err := uc.RegisterApp(pkgCtx.NewContextFromFiberCtx(c), registerRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, registerResponse)
}

func VerifyApp(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	verifyRequest := models.VerifyRequest{}
	if err := c.Bind().Body(&verifyRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	verifyResponse, err := uc.VerifyApp(pkgCtx.NewContextFromFiberCtx(c), verifyRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, verifyResponse)
}

func RefreshApp(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	refreshRequest := models.RefreshRequest{}
	if err := c.Bind().Body(&refreshRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	refreshResponse, err := uc.RefreshApp(pkgCtx.NewContextFromFiberCtx(c), refreshRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, refreshResponse)
}

func ListApps(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	listRequest := models.ListRequest{}
	if err := c.Bind().Query(&listRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	listResponse, err := uc.ListApps(pkgCtx.NewContextFromFiberCtx(c), listRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, listResponse)
}

func GetApp(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	appService, err := uc.GetApp(pkgCtx.NewContextFromFiberCtx(c), c.Params("appServiceID"))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, appService)
}

func UpdateAppAppearance(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateAppearanceRequest := models.UpdateAppearanceRequest{}
	if err := c.Bind().Body(&updateAppearanceRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	if err := uc.UpdateAppearance(pkgCtx.NewContextFromFiberCtx(c), c.Params("appServiceID"), updateAppearanceRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func UpdateAppStatus(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateStatusRequest := models.UpdateStatusRequest{}
	if err := c.Bind().Body(&updateStatusRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	if err := uc.UpdateStatus(pkgCtx.NewContextFromFiberCtx(c), c.Params("appServiceID"), updateStatusRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}
