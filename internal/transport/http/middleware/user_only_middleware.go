package middleware

import (
	"net/http"

	"github.com/internships-backend/test-backend-SIniutin/internal/domain"
)

func UserOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ClaimsKey).(Claims)
		if !ok || claims.Role != domain.RoleUser {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "user only")
			return
		}
		next.ServeHTTP(w, r)
	})
}
