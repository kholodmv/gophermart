package auth

import (
	"context"
	"github.com/kholodmv/gophermart/internal/auth"
	"net/http"
	"strings"
)

type key string

var LoginKey key = "login"

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			http.Error(w, "malformed token", http.StatusUnauthorized)
			return
		}

		token := authHeader[1]
		login, err := auth.GetLogin(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		newContext := context.WithValue(r.Context(), LoginKey, login)
		next.ServeHTTP(w, r.WithContext(newContext))
	})
}
