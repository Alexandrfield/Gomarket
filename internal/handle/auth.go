package handle

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type AuthorizationServer struct {
	secretKey []byte
	tockenExp time.Duration
}

func (autServer *AuthorizationServer) Init() {
	autServer.secretKey = []byte("supersecretkey")
	autServer.tockenExp = time.Hour * 3
}

func (autServer *AuthorizationServer) BuildJWTString(idUser string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(autServer.tockenExp)),
		},
		UserID: idUser,
	})

	tokenString, err := token.SignedString(autServer.secretKey)
	if err != nil {
		return "", fmt.Errorf("errors sign. err:%w", err)
	}
	return tokenString, nil
}

func (autServer *AuthorizationServer) CheckToken(tokenString string) (bool, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return autServer.secretKey, nil
	})

	if err != nil {
		return false, fmt.Errorf("error parse jwt. err:%w", err)
	}
	if !token.Valid {
		return false, nil
	}
	return true, nil
}

func (autServer *AuthorizationServer) CheckTokenGetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return autServer.secretKey, nil
	})

	if err != nil {
		return "-1", fmt.Errorf("error parse jwt. err:%w", err)
	}
	if !token.Valid {
		return "-1", nil
	}
	return claims.UserID, nil
}
