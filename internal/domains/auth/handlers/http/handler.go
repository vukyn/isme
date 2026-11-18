package handlers

import (
	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/auth/models"
	pkgCtx "github.com/vukyn/kuery/ctxv3"
	pkgErr "github.com/vukyn/kuery/http/errors"
	pkgHttp "github.com/vukyn/kuery/http/fiberv3"

	"github.com/gofiber/fiber/v3"
)

func Login(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	loginRequest := models.LoginRequest{}
	if err := c.Bind().Body(&loginRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	loginResponse, err := uc.Login(pkgCtx.NewContextFromFiberCtx(c), loginRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, loginResponse)
}

func GetMe(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

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

func UpdateMe(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	updateMeRequest := models.UpdateMeRequest{}
	if err := c.Bind().Body(&updateMeRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	getMeResponse, err := uc.UpdateMe(pkgCtx.NewContextFromFiberCtx(c), updateMeRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, getMeResponse)
}

func RefreshToken(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	refreshTokenRequest := models.RefreshTokenRequest{}
	if err := c.Bind().Body(&refreshTokenRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	refreshTokenResponse, err := uc.RefreshToken(pkgCtx.NewContextFromFiberCtx(c), refreshTokenRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, refreshTokenResponse)
}

func ChangePassword(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	changePasswordRequest := models.ChangePasswordRequest{}
	if err := c.Bind().Body(&changePasswordRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	err = uc.ChangePassword(pkgCtx.NewContextFromFiberCtx(c), changePasswordRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, map[string]string{"message": "Password changed successfully. Please login again!"})
}

func Logout(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

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

func RequestLogin(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	requestLoginRequest := models.RequestLoginRequest{}
	if err := c.Bind().Body(&requestLoginRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	requestLoginResponse, err := uc.RequestLogin(pkgCtx.NewContextFromFiberCtx(c), requestLoginRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, requestLoginResponse)
}

func ExchangeCode(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	exchangeCodeRequest := models.ExchangeCodeRequest{}
	if err := c.Bind().Body(&exchangeCodeRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	exchangeCodeResponse, err := uc.ExchangeCode(pkgCtx.NewContextFromFiberCtx(c), exchangeCodeRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, exchangeCodeResponse)
}

func SSOCheck(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	ssoCheckRequest := models.SSOCheckRequest{}
	if err := c.Bind().Body(&ssoCheckRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	ssoCheckResponse, err := uc.SSOCheck(pkgCtx.NewContextFromFiberCtx(c), ssoCheckRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, ssoCheckResponse)
}

func ListMySessions(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

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

func CountMySessions(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

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

func RevokeMySession(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

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

func RevokeMyOtherSessions(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

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

func GetMyActivity(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	activity, err := uc.GetMyActivity(pkgCtx.NewContextFromFiberCtx(c), fiber.Query[int](c, "limit", 0))
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, activity)
}

func SSOConsent(c fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)

	uc, err := idi.GetAuthUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	ssoConsentRequest := models.SSOConsentRequest{}
	if err := c.Bind().Body(&ssoConsentRequest); err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest(err.Error()))
	}

	ssoConsentResponse, err := uc.SSOConsent(pkgCtx.NewContextFromFiberCtx(c), ssoConsentRequest)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, ssoConsentResponse)
}
