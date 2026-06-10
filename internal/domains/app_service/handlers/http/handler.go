package handlers

import (
	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/app_service/models"
	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgHttp "github.com/vukyn/kuery/http/fiber"

	"github.com/gofiber/fiber/v2"
)

func RegisterApp(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	registerRequest := models.RegisterRequest{}
	if err := c.BodyParser(&registerRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	registerResponse, err := uc.RegisterApp(pkgCtx.NewContextFromFiberCtx(c), registerRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, registerResponse)
}

func VerifyApp(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	verifyRequest := models.VerifyRequest{}
	if err := c.BodyParser(&verifyRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	verifyResponse, err := uc.VerifyApp(pkgCtx.NewContextFromFiberCtx(c), verifyRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, verifyResponse)
}

func RefreshApp(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	refreshRequest := models.RefreshRequest{}
	if err := c.BodyParser(&refreshRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	refreshResponse, err := uc.RefreshApp(pkgCtx.NewContextFromFiberCtx(c), refreshRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, refreshResponse)
}

func ListApps(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	listRequest := models.ListRequest{}
	if err := c.QueryParser(&listRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	listResponse, err := uc.ListApps(pkgCtx.NewContextFromFiberCtx(c), listRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, listResponse)
}

func GetApp(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

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

func UpdateAppAppearance(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateAppearanceRequest := models.UpdateAppearanceRequest{}
	if err := c.BodyParser(&updateAppearanceRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.UpdateAppearance(pkgCtx.NewContextFromFiberCtx(c), c.Params("appServiceID"), updateAppearanceRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func UpdateAppStatus(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAppServiceUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateStatusRequest := models.UpdateStatusRequest{}
	if err := c.BodyParser(&updateStatusRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.UpdateStatus(pkgCtx.NewContextFromFiberCtx(c), c.Params("appServiceID"), updateStatusRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}
