package middleware

import "github.com/internships-backend/test-backend-SIniutin/internal/domain"

type contextKey string

const ClaimsKey contextKey = "claims"

type Claims struct {
	UserID string
	Role   domain.UserRole
}
