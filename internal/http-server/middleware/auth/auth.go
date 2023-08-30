package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kholodmv/gophermart/internal/auth"
	"github.com/kholodmv/gophermart/internal/logger"
	"net/http"
	"strings"
)

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.SetupLogger("dev")
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		log.Info("authHeader", authHeader)

		jwtToken := authHeader[1]
		log.Info("jwtToken", jwtToken)
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Info("unexpected signing method: %v", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			log.Info("auth secret key", auth.SecretKey)
			return []byte(auth.SecretKey), nil
		})

		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			log.Info("next.ServeHTTP(w, r)")
			next.ServeHTTP(w, r)
		} else {
			fmt.Println(err)
			log.Info("Unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		}
	})
}
