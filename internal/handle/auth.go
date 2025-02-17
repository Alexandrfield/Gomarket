package handle

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

//const TOKEN_EXP = time.Hour * 3

// const SECRET_KEY = "supersecretkey"
type AuthorizationServer struct {
	secretKey []byte
	tockenExp time.Duration
}

func (autServer *AuthorizationServer) Init() {
	autServer.secretKey = []byte("supersecretkey")
	autServer.tockenExp = time.Hour * 3
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (autServer *AuthorizationServer) BuildJWTString() (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(autServer.tockenExp)),
		},
		// собственное утверждение
		UserID: 1,
	})

	// создаём строку токена
	tokenString, err := token.SignedString(autServer.secretKey)
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (autServer *AuthorizationServer) CheckToken(tokenString string) (bool, error) {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
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

func (autServer *AuthorizationServer) CheckTokenGetUserID(tokenString string) (int, error) {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return autServer.secretKey, nil
	})

	if err != nil {
		return -1, fmt.Errorf("error parse jwt. err:%w", err)
	}
	if !token.Valid {
		return -1, nil
	}
	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, nil
}
