package handlers

import (
	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/role/models"

	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgHttp "github.com/vukyn/kuery/http/fiber"

	"github.com/gofiber/fiber/v2"
)

func ListRoles(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	listRequest := models.ListRequest{}
	if err := c.QueryParser(&listRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	roles, err := uc.List(pkgCtx.NewContextFromFiberCtx(c), listRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, roles)
}

func CreateRole(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	createRequest := models.CreateRequest{}
	if err := c.BodyParser(&createRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	createResponse, err := uc.Create(pkgCtx.NewContextFromFiberCtx(c), createRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, createResponse)
}

func GetRoleDetail(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	detailResponse, err := uc.GetDetail(pkgCtx.NewContextFromFiberCtx(c), c.Params("roleID"))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, detailResponse)
}

func UpdateRole(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateRequest := models.UpdateRequest{}
	if err := c.BodyParser(&updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.Update(pkgCtx.NewContextFromFiberCtx(c), c.Params("roleID"), updateRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func DeleteRole(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.Delete(pkgCtx.NewContextFromFiberCtx(c), c.Params("roleID")); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func SetRolePermissions(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	setPermissionsRequest := models.SetPermissionsRequest{}
	if err := c.BodyParser(&setPermissionsRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.SetPermissions(pkgCtx.NewContextFromFiberCtx(c), c.Params("roleID"), setPermissionsRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func ListRoleMembers(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	listMembersRequest := models.ListMembersRequest{}
	if err := c.QueryParser(&listMembersRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	listMembersResponse, err := uc.ListMembers(pkgCtx.NewContextFromFiberCtx(c), c.Params("roleID"), listMembersRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, listMembersResponse)
}

func AddRoleMembers(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	addMembersRequest := models.AddMembersRequest{}
	if err := c.BodyParser(&addMembersRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	if err := uc.AddMembers(pkgCtx.NewContextFromFiberCtx(c), c.Params("roleID"), addMembersRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func RemoveRoleMember(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	var appServiceID *string
	if appServiceIDQuery := c.Query("app_service_id"); appServiceIDQuery != "" {
		appServiceID = &appServiceIDQuery
	}

	if err := uc.RemoveMember(pkgCtx.NewContextFromFiberCtx(c), c.Params("roleID"), c.Params("userID"), appServiceID); err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, nil)
}

func ListPermissions(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetRoleUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	listPermissionsRequest := models.ListPermissionsRequest{}
	if err := c.QueryParser(&listPermissionsRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	permissions, err := uc.ListPermissions(pkgCtx.NewContextFromFiberCtx(c), listPermissionsRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, permissions)
}
