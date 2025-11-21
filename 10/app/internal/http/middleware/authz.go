package middleware

import (
	"app/internal/core"
	"net/http"
)

func AuthZRoles(allowed ...string) func(http.Handler) http.Handler {
	set := make(map[string]struct{})
	for _, a := range allowed {
		set[a] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(core.ClaimsKey).(map[string]any)
			if !ok {
				http.Error(w, `{"error":"no_claims","details":"No claims found in context"}`, http.StatusInternalServerError)
				return
			}

			role, ok := claims["role"].(string)
			if !ok {
				http.Error(w, `{"error":"no_role","details":"No role claim found"}`, http.StatusForbidden)
				return
			}

			if _, ok := set[role]; !ok {
				http.Error(w, `{"error":"forbidden","details":"Insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
