package utils

import (
	"context"
	"encoding/hex"
	"github.com/kholodmv/gophermart/internal/http-server/middleware/auth"
	"golang.org/x/crypto/bcrypt"
)

func IsValidLuhnNumber(number string) bool {
	var sum int
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

func GenerateHashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	hashPassword := hex.EncodeToString(hash)

	return hashPassword, nil
}

func CompareHashAndPassword(hash, password string) error {
	decHash, _ := hex.DecodeString(hash)

	err := bcrypt.CompareHashAndPassword([]byte(decHash), []byte(password))
	return err
}

func GetLogin(ctx context.Context) string {
	return ctx.Value(auth.LoginKey).(string)
}
