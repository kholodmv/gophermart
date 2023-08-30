package auth

import (
	"context"
	"github.com/kholodmv/gophermart/internal/auth"
	"net/http"
	"strings"
)

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			http.Error(w, "malformed token", http.StatusUnauthorized)
			return
		}

		token := authHeader[1]
		userInfo, err := auth.GetUserInfo(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		newContext := context.WithValue(r.Context(), string("userInfo"), userInfo)
		next.ServeHTTP(w, r.WithContext(newContext))
	})
}
