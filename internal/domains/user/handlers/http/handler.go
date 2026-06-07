package handlers

import (
	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/user/models"

	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgHttp "github.com/vukyn/kuery/http/fiber"

	"github.com/gofiber/fiber/v2"
)

func ListUsers(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetUserUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	listRequest := models.ListRequest{}
	if err := c.QueryParser(&listRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	listResponse, err := uc.List(pkgCtx.NewContextFromFiberCtx(c), listRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, listResponse)
}

func UpdateUserStatus(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetUserUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateStatusRequest := models.UpdateStatusRequest{}
	if err := c.BodyParser(&updateStatusRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.UpdateStatus(pkgCtx.NewContextFromFiberCtx(c), c.Params("userID"), updateStatusRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func VerifyUser(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetUserUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.VerifyUser(pkgCtx.NewContextFromFiberCtx(c), c.Params("userID")); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func DeleteUser(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetUserUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.SoftDelete(pkgCtx.NewContextFromFiberCtx(c), c.Params("userID")); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func ListUserSessions(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetUserUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	sessions, err := uc.ListSessions(pkgCtx.NewContextFromFiberCtx(c), c.Params("userID"))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, sessions)
}

func RevokeUserSession(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetUserUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.RevokeSession(pkgCtx.NewContextFromFiberCtx(c), c.Params("userID"), c.Params("sessionID")); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}
