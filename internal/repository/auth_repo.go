package repository

import (
	"context"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
)

type UserRepo interface {
	Create(ctx context.Context, user d.User, creds d.Credentials) (d.User, error)
	FindByEmail(ctx context.Context, email d.Email) (d.User, error)
	FindByID(ctx context.Context, id d.UserID) (d.User, error)
	GetCredentials(ctx context.Context, id d.UserID) (d.Credentials, error)
}
