package handlers

import (
	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/user_invitation/models"

	pkgCtx "github.com/vukyn/kuery/ctxv3"
	pkgErr "github.com/vukyn/kuery/http/errors"
	pkgHttp "github.com/vukyn/kuery/http/fiberv3"

	"github.com/gofiber/fiber/v3"
)

func CreateInvitation(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetUserInvitationUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	createRequest := models.CreateRequest{}
	if err := c.Bind().Body(&createRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	createResponse, err := uc.Create(pkgCtx.NewContextFromFiberCtx(c), createRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, createResponse)
}

func ListInvitations(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetUserInvitationUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	listResponse, err := uc.List(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, listResponse)
}

func RevokeInvitation(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetUserInvitationUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.Revoke(pkgCtx.NewContextFromFiberCtx(c), c.Params("invitationID")); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func GetInvitationByToken(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetUserInvitationUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	detailResponse, err := uc.GetByToken(pkgCtx.NewContextFromFiberCtx(c), c.Params("token"))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, detailResponse)
}

func AcceptInvitation(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetUserInvitationUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	acceptRequest := models.AcceptRequest{}
	if err := c.Bind().Body(&acceptRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	if err := uc.Accept(pkgCtx.NewContextFromFiberCtx(c), acceptRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}
