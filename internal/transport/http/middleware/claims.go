package middleware

import "github.com/SIniutin/room-booking-service/internal/domain"

type contextKey string

const ClaimsKey contextKey = "claims"

type Claims struct {
	UserID string
	Role   domain.UserRole
}
