package auth

import (
	"context"
	"github.com/kholodmv/gophermart/internal/auth"
	"github.com/labstack/gommon/log"
	"net/http"
	"strings"
)

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeaderValue := r.Header.Get("Authorization")
		log.Error("Auth header from AuthenticationMiddleware is: " + authHeaderValue)
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			http.Error(w, "malformed token", http.StatusUnauthorized)
			return
		}

		token := authHeader[1]
		log.Error("Token from AuthenticationMiddleware is: " + token)
		userInfo, err := auth.GetUserInfo(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		newContext := context.WithValue(r.Context(), string("userInfo"), userInfo)
		next.ServeHTTP(w, r.WithContext(newContext))
	})
}
