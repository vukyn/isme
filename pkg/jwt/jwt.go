package jwt

import (
	"errors"
	pkgClaims "isme/pkg/claims"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(secretKey string, expireIn int, userID, email string) (string, pkgClaims.Claims, error) {
	claims := pkgClaims.NewClaims(userID, email, int64(expireIn))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims.MapClaims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", pkgClaims.Claims{}, err
	}
	return tokenString, claims, nil
}

func ValidateJWT(tokenString, secretKey string) (pkgClaims.Claims, error) {
	claims := pkgClaims.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims.MapClaims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, errors.New("invalid token")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return pkgClaims.Claims{}, err
	}
	if !token.Valid {
		return pkgClaims.Claims{}, errors.New("invalid token")
	}
	return claims, nil
}
