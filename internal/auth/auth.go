package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kholodmv/gophermart/internal/models"
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

func GetUserInfo(tokenString string) (*models.User, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	return &models.User{
			Login: claims.Login},
		nil
}
