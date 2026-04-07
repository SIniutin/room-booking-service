package usecase

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	r "github.com/internships-backend/test-backend-SIniutin/internal/repository"
)

var (
	dummyAdminID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	dummyUserID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

type authUsecaseImpl struct {
	ttl        time.Duration
	privateKey *rsa.PrivateKey
	repo       r.UserRepo
}

func NewAuthUsecase(privateKey *rsa.PrivateKey, repo r.UserRepo, ttl time.Duration) AuthUsecase {
	return &authUsecaseImpl{
		repo:       repo,
		privateKey: privateKey,
		ttl:        ttl,
	}
}

func (uc *authUsecaseImpl) Register(ctx context.Context, email, password string, role d.UserRole) (d.User, error) {
	if password == "" {
		return d.User{}, d.ErrEmptyPassword
	}
	if role != d.RoleUser && role != d.RoleAdmin {
		return d.User{}, d.ErrRoleIsUndefined
	}

	validEmail, err := d.NewEmail(email)
	if err != nil {
		return d.User{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return d.User{}, err
	}

	user := d.User{
		ID:    d.UserID(uuid.New()),
		Email: *validEmail,
		Role:  role,
	}

	creds := d.Credentials{
		UserID:       user.ID,
		PasswordHash: string(passwordHash),
	}

	return uc.repo.Create(ctx, user, creds)
}

func (uc *authUsecaseImpl) Login(ctx context.Context, email, password string) (d.Token, error) {
	select {
	case <-ctx.Done():
		return "", d.ErrCtxCancelled
	default:
	}

	validEmail, err := d.NewEmail(email)
	if err != nil {
		return "", d.ErrInvalidCredentials
	}

	user, err := uc.repo.FindByEmail(ctx, *validEmail)
	if err != nil {
		return "", d.ErrInvalidCredentials
	}

	creds, err := uc.repo.GetCredentials(ctx, user.ID)
	if err != nil {
		return "", d.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(password)); err != nil {
		return "", d.ErrInvalidCredentials
	}

	return uc.signToken(user.ID, user.Role)
}

func (uc *authUsecaseImpl) DummyLogin(ctx context.Context, role d.UserRole) (d.Token, error) {
	select {
	case <-ctx.Done():
		return "", d.ErrCtxCancelled
	default:
	}

	var user d.User
	switch role {
	case d.RoleAdmin:
		user = d.User{
			ID:    d.UserID(dummyAdminID),
			Email: d.Email("dummy-admin@example.local"),
			Role:  d.RoleAdmin,
		}
	case d.RoleUser:
		user = d.User{
			ID:    d.UserID(dummyUserID),
			Email: d.Email("dummy-user@example.local"),
			Role:  d.RoleUser,
		}
	default:
		return "", d.ErrRoleIsUndefined
	}

	if err := uc.ensureDummyUser(ctx, user); err != nil {
		return "", err
	}

	return uc.signToken(user.ID, role)
}

func (uc *authUsecaseImpl) signToken(id d.UserID, role d.UserRole) (d.Token, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":  uuid.UUID(id).String(),
		"role": role,
		"exp":  time.Now().Add(uc.ttl).Unix(),
	})

	signed, err := token.SignedString(uc.privateKey)
	if err != nil {
		return "", err
	}

	return d.Token(signed), nil
}

func (uc *authUsecaseImpl) ensureDummyUser(ctx context.Context, user d.User) error {
	_, err := uc.repo.FindByID(ctx, user.ID)
	if err == nil {
		return nil
	}
	if err != nil && err != d.ErrUserNotFound {
		return err
	}

	_, err = uc.repo.Create(ctx, user, d.Credentials{
		UserID:       user.ID,
		PasswordHash: "",
	})
	if err != nil && err != d.ErrUserAlreadyExists {
		return err
	}

	return nil
}
