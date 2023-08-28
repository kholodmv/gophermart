package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

var SecretKey = []byte("secret-key")

type Claims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

func GenerateToken(login string) (string, error) {
	claims := &Claims{
		Login: login,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(SecretKey)
	return tokenString, err
}

func GetLogin(req *http.Request) string {
	authHeader := req.Header.Get("Authorization")
	jwtString := strings.Split(authHeader, "Bearer ")[1]

	claims := &Claims{}
	_, _ = jwt.ParseWithClaims(jwtString, claims, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	return claims.Login
}
