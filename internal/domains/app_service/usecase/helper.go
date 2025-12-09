package usecase

import (
	"github.com/vukyn/isme/pkg/cryp/aes"
	"github.com/vukyn/isme/pkg/cryp/rand"
	pkgErr "github.com/vukyn/isme/pkg/http/errors"
)

func generateAndEncryptAppSecret(aesSecret string, ctxInfo string) (string, string, error) {
	appSecret := rand.RandMixedString(8, true, true)
	if appSecret == "" {
		return "", "", pkgErr.InvalidRequest("failed to generate app_secret")
	}
	encryptedAppSecret, err := aes.Encrypt(appSecret, aesSecret, ctxInfo)
	if err != nil {
		return "", "", pkgErr.InvalidRequest("failed to encrypt app_secret: " + err.Error())
	}
	return appSecret, encryptedAppSecret, nil
}

func compareAppSecret(appSecret1 string, appSecret2 string, aesSecret string, ctxInfo string) (bool, error) {
	decryptedAppSecret1, err := aes.Decrypt(appSecret1, aesSecret, ctxInfo)
	if err != nil {
		return false, pkgErr.InvalidRequest("failed to decrypt app_secret: " + err.Error())
	}
	decryptedAppSecret2, err := aes.Decrypt(appSecret2, aesSecret, ctxInfo)
	if err != nil {
		return false, pkgErr.InvalidRequest("failed to decrypt app_secret: " + err.Error())
	}
	return decryptedAppSecret1 == decryptedAppSecret2, nil
}
