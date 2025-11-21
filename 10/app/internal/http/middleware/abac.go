package middleware

import (
	"app/internal/core"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// ABACMiddleware проверяет, что пользователь может доступть только свои данные
// если он не администратор
func ABACMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(core.ClaimsKey).(map[string]any)
		if !ok {
			http.Error(w, `{"error":"no_claims","details":"No claims found in context"}`, http.StatusInternalServerError)
			return
		}

		// Извлекаем userID из claims
		sub, ok := claims["sub"].(float64)
		if !ok {
			http.Error(w, `{"error":"invalid_user_id","details":"Invalid user ID in token"}`, http.StatusBadRequest)
			return
		}
		userID := int64(sub)

		// Извлекаем роль
		role, ok := claims["role"].(string)
		if !ok {
			http.Error(w, `{"error":"no_role","details":"No role claim found"}`, http.StatusForbidden)
			return
		}

		// Админы могут всё
		if role == "admin" {
			next.ServeHTTP(w, r)
			return
		}

		// Для пользователей проверяем, что они запрашивают свои данные
		requestedIDStr := chi.URLParam(r, "id")
		requestedID, err := strconv.ParseInt(requestedIDStr, 10, 64)
		if err != nil {
			http.Error(w, `{"error":"invalid_user_id","details":"Invalid user ID in URL"}`, http.StatusBadRequest)
			return
		}

		if userID != requestedID {
			http.Error(w, `{"error":"forbidden","details":"You can only access your own data"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
