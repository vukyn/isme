package handlers

import (
	idi "isme/internal/di"
	"isme/internal/domains/auth/models"
	pkgCtx "isme/pkg/ctx"
	pkgHttp "isme/pkg/http/fiber"

	"github.com/gofiber/fiber/v2"
)

func SignUp(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	signUpRequest := models.SignUpRequest{}
	if err := c.BodyParser(&signUpRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	signUpRepsonse, err := uc.SignUp(pkgCtx.NewContextFromFiberCtx(c), signUpRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, signUpRepsonse)
}

func Login(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	loginRequest := models.LoginRequest{}
	if err := c.BodyParser(&loginRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	loginResponse, err := uc.Login(pkgCtx.NewContextFromFiberCtx(c), loginRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, loginResponse)
}

func GetMe(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	getMeResponse, err := uc.GetMe(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, getMeResponse)
}

func RefreshToken(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	refreshTokenRequest := models.RefreshTokenRequest{}
	if err := c.BodyParser(&refreshTokenRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	refreshTokenResponse, err := uc.RefreshToken(pkgCtx.NewContextFromFiberCtx(c), refreshTokenRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, refreshTokenResponse)
}

func ChangePassword(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	changePasswordRequest := models.ChangePasswordRequest{}
	if err := c.BodyParser(&changePasswordRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	err = uc.ChangePassword(pkgCtx.NewContextFromFiberCtx(c), changePasswordRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, map[string]string{"message": "Password changed successfully. Please login again!"})
}

func Logout(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	err = uc.Logout(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, map[string]string{"message": "Logged out successfully"})
}

func RequestLogin(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	requestLoginRequest := models.RequestLoginRequest{}
	if err := c.BodyParser(&requestLoginRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	requestLoginResponse, err := uc.RequestLogin(pkgCtx.NewContextFromFiberCtx(c), requestLoginRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, requestLoginResponse)
}
