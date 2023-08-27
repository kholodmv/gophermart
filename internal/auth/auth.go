package auth

import (
	"github.com/dgrijalva/jwt-go"
)

var secretKey = []byte("your-secret-key")

func CreateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
