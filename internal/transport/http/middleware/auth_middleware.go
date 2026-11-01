package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/SIniutin/room-booking-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return publicKey, nil
			})
			if err != nil || !token.Valid {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}

			mapClaims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token claims")
				return
			}

			userID, _ := mapClaims["sub"].(string)
			role, _ := mapClaims["role"].(string)

			if userID == "" || role == "" {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token claims")
				return
			}

			claims := Claims{
				UserID: userID,
				Role:   domain.UserRole(role),
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
