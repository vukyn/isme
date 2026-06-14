package handlers

import (
	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/auth/models"
	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgHttp "github.com/vukyn/kuery/http/fiber"

	"github.com/gofiber/fiber/v2"
)

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

func UpdateMe(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateMeRequest := models.UpdateMeRequest{}
	if err := c.BodyParser(&updateMeRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	getMeResponse, err := uc.UpdateMe(pkgCtx.NewContextFromFiberCtx(c), updateMeRequest)
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

func ExchangeCode(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	exchangeCodeRequest := models.ExchangeCodeRequest{}
	if err := c.BodyParser(&exchangeCodeRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	exchangeCodeResponse, err := uc.ExchangeCode(pkgCtx.NewContextFromFiberCtx(c), exchangeCodeRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, exchangeCodeResponse)
}

func SSOCheck(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	ssoCheckRequest := models.SSOCheckRequest{}
	if err := c.BodyParser(&ssoCheckRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	ssoCheckResponse, err := uc.SSOCheck(pkgCtx.NewContextFromFiberCtx(c), ssoCheckRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, ssoCheckResponse)
}

func ListMySessions(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	sessions, err := uc.ListMySessions(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, sessions)
}

func CountMySessions(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	count, err := uc.CountMySessions(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, count)
}

func RevokeMySession(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	err = uc.RevokeMySession(pkgCtx.NewContextFromFiberCtx(c), c.Params("id"))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, map[string]string{"message": "Session revoked successfully"})
}

func RevokeMyOtherSessions(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	err = uc.RevokeMyOtherSessions(pkgCtx.NewContextFromFiberCtx(c))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, map[string]string{"message": "Other sessions revoked successfully"})
}

func GetMyActivity(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	activity, err := uc.GetMyActivity(pkgCtx.NewContextFromFiberCtx(c), c.QueryInt("limit", 0))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, activity)
}

func SSOConsent(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	ssoConsentRequest := models.SSOConsentRequest{}
	if err := c.BodyParser(&ssoConsentRequest); err != nil {
		return pkgHttp.Err(c, err)
	}

	ssoConsentResponse, err := uc.SSOConsent(pkgCtx.NewContextFromFiberCtx(c), ssoConsentRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, ssoConsentResponse)
}
