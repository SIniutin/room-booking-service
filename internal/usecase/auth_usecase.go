package usecase

import (
	"context"

	d "github.com/SIniutin/room-booking-service/internal/domain"
)

type AuthUsecase interface {
	Register(ctx context.Context, email, password string, role d.UserRole) (d.User, error)
	Login(ctx context.Context, email, password string) (d.Token, error)
	DummyLogin(ctx context.Context, role d.UserRole) (d.Token, error)
}
