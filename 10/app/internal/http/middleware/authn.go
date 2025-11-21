package middleware

import (
	"context"
	"net/http"
	"strings"

	"app/internal/core"
	"app/internal/platform/jwt"
)

func AuthN(v jwt.Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, `{"error":"unauthorized","details":"Missing or invalid Authorization header"}`, http.StatusUnauthorized)
				return
			}

			raw := strings.TrimPrefix(h, "Bearer ")
			claims, err := v.Parse(raw)
			if err != nil {
				http.Error(w, `{"error":"invalid_token","details":"Token validation failed"}`, http.StatusUnauthorized)
				return
			}

			// Проверяем тип токена
			if tokenType, ok := claims["type"].(string); !ok || tokenType != "access" {
				http.Error(w, `{"error":"invalid_token_type","details":"Not an access token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), core.ClaimsKey, map[string]any(claims))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
