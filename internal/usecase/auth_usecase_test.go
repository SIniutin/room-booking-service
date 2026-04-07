package usecase_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	"github.com/internships-backend/test-backend-SIniutin/internal/usecase"
)

func newAuthUsecase(t *testing.T, repo *mockUserRepo) (*rsa.PrivateKey, usecase.AuthUsecase) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return key, usecase.NewAuthUsecase(key, repo, time.Hour)
}

func TestRegister_Success(t *testing.T) {
	var createdUser d.User
	var createdCreds d.Credentials

	_, uc := newAuthUsecase(t, &mockUserRepo{
		create: func(_ context.Context, user d.User, creds d.Credentials) (d.User, error) {
			createdUser = user
			createdCreds = creds
			return user, nil
		},
	})

	got, err := uc.Register(context.Background(), " User@Example.COM ", "secret123", d.RoleUser)

	require.NoError(t, err)
	assert.Equal(t, d.RoleUser, got.Role)
	assert.Equal(t, d.Email("user@example.com"), got.Email)
	assert.Equal(t, got, createdUser)
	assert.Equal(t, got.ID, createdCreds.UserID)
	assert.NotEqual(t, "secret123", createdCreds.PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(createdCreds.PasswordHash), []byte("secret123")))
}

func TestRegister_EmptyPassword(t *testing.T) {
	_, uc := newAuthUsecase(t, &mockUserRepo{})

	_, err := uc.Register(context.Background(), "user@example.com", "", d.RoleUser)

	assert.ErrorIs(t, err, d.ErrEmptyPassword)
}

func TestRegister_InvalidRole(t *testing.T) {
	_, uc := newAuthUsecase(t, &mockUserRepo{})

	_, err := uc.Register(context.Background(), "user@example.com", "secret123", d.UserRole("guest"))

	assert.ErrorIs(t, err, d.ErrRoleIsUndefined)
}

func TestRegister_RepoError(t *testing.T) {
	repoErr := errors.New("create user failed")
	_, uc := newAuthUsecase(t, &mockUserRepo{
		create: func(_ context.Context, _ d.User, _ d.Credentials) (d.User, error) {
			return d.User{}, repoErr
		},
	})

	_, err := uc.Register(context.Background(), "user@example.com", "secret123", d.RoleUser)

	assert.ErrorIs(t, err, repoErr)
}

func TestLogin_Success(t *testing.T) {
	privateKey, uc := newAuthUsecase(t, &mockUserRepo{
		findByEmail: func(_ context.Context, email d.Email) (d.User, error) {
			assert.Equal(t, d.Email("user@example.com"), email)
			return d.User{
				ID:    d.UserID(uuid.New()),
				Email: email,
				Role:  d.RoleAdmin,
			}, nil
		},
		getCredentials: func(_ context.Context, _ d.UserID) (d.Credentials, error) {
			hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
			require.NoError(t, err)
			return d.Credentials{PasswordHash: string(hash)}, nil
		},
	})

	token, err := uc.Login(context.Background(), "User@Example.com", "secret123")

	require.NoError(t, err)
	parsed, err := jwt.Parse(string(token), func(token *jwt.Token) (any, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, string(d.RoleAdmin), claims["role"])
	assert.NotEmpty(t, claims["sub"])
	assert.NotZero(t, claims["exp"])
}

func TestLogin_ContextCancelled(t *testing.T) {
	_, uc := newAuthUsecase(t, &mockUserRepo{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := uc.Login(ctx, "user@example.com", "secret123")

	assert.ErrorIs(t, err, d.ErrCtxCancelled)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	t.Run("invalid email", func(t *testing.T) {
		_, uc := newAuthUsecase(t, &mockUserRepo{})

		_, err := uc.Login(context.Background(), "not-an-email", "secret123")

		assert.ErrorIs(t, err, d.ErrInvalidCredentials)
	})

	tests := []struct {
		name string
		repo *mockUserRepo
	}{
		{
			name: "user not found",
			repo: &mockUserRepo{
				findByEmail: func(_ context.Context, _ d.Email) (d.User, error) {
					return d.User{}, d.ErrUserNotFound
				},
			},
		},
		{
			name: "credentials missing",
			repo: &mockUserRepo{
				findByEmail: func(_ context.Context, email d.Email) (d.User, error) {
					return d.User{ID: d.UserID(uuid.New()), Email: email, Role: d.RoleUser}, nil
				},
				getCredentials: func(_ context.Context, _ d.UserID) (d.Credentials, error) {
					return d.Credentials{}, d.ErrUserNotFound
				},
			},
		},
		{
			name: "wrong password",
			repo: &mockUserRepo{
				findByEmail: func(_ context.Context, email d.Email) (d.User, error) {
					return d.User{ID: d.UserID(uuid.New()), Email: email, Role: d.RoleUser}, nil
				},
				getCredentials: func(_ context.Context, _ d.UserID) (d.Credentials, error) {
					hash, err := bcrypt.GenerateFromPassword([]byte("another-password"), bcrypt.DefaultCost)
					require.NoError(t, err)
					return d.Credentials{PasswordHash: string(hash)}, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, uc := newAuthUsecase(t, tt.repo)

			_, err := uc.Login(context.Background(), "user@example.com", "secret123")

			assert.ErrorIs(t, err, d.ErrInvalidCredentials)
		})
	}
}

func TestDummyLogin_Success(t *testing.T) {
	privateKey, uc := newAuthUsecase(t, &mockUserRepo{})

	token, err := uc.DummyLogin(context.Background(), d.RoleUser)

	require.NoError(t, err)
	parsed, err := jwt.Parse(string(token), func(token *jwt.Token) (any, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, string(d.RoleUser), claims["role"])
	assert.Equal(t, uuid.UUID(usecase_dummyUserID()).String(), claims["sub"])
}

func TestDummyLogin_CreatesMissingDummyUser(t *testing.T) {
	var (
		createdUser  d.User
		createdCreds d.Credentials
	)

	_, uc := newAuthUsecase(t, &mockUserRepo{
		findByID: func(_ context.Context, _ d.UserID) (d.User, error) {
			return d.User{}, d.ErrUserNotFound
		},
		create: func(_ context.Context, user d.User, creds d.Credentials) (d.User, error) {
			createdUser = user
			createdCreds = creds
			return user, nil
		},
	})

	_, err := uc.DummyLogin(context.Background(), d.RoleAdmin)

	require.NoError(t, err)
	assert.Equal(t, d.UserID(usecase_dummyAdminID()), createdUser.ID)
	assert.Equal(t, d.Email("dummy-admin@example.local"), createdUser.Email)
	assert.Equal(t, d.RoleAdmin, createdUser.Role)
	assert.Equal(t, createdUser.ID, createdCreds.UserID)
	assert.Empty(t, createdCreds.PasswordHash)
}

func TestDummyLogin_DoesNotCreateExistingDummyUser(t *testing.T) {
	createCalled := false
	existingUser := d.User{
		ID:    d.UserID(usecase_dummyUserID()),
		Email: d.Email("dummy-user@example.local"),
		Role:  d.RoleUser,
	}

	_, uc := newAuthUsecase(t, &mockUserRepo{
		findByID: func(_ context.Context, id d.UserID) (d.User, error) {
			assert.Equal(t, existingUser.ID, id)
			return existingUser, nil
		},
		create: func(_ context.Context, _ d.User, _ d.Credentials) (d.User, error) {
			createCalled = true
			return d.User{}, nil
		},
	})

	_, err := uc.DummyLogin(context.Background(), d.RoleUser)

	require.NoError(t, err)
	assert.False(t, createCalled)
}

func TestDummyLogin_ContextCancelled(t *testing.T) {
	_, uc := newAuthUsecase(t, &mockUserRepo{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := uc.DummyLogin(ctx, d.RoleUser)

	assert.ErrorIs(t, err, d.ErrCtxCancelled)
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	_, uc := newAuthUsecase(t, &mockUserRepo{})

	_, err := uc.DummyLogin(context.Background(), d.UserRole("guest"))

	assert.ErrorIs(t, err, d.ErrRoleIsUndefined)
}

func usecase_dummyUserID() d.UserID {
	return d.UserID(uuid.MustParse("00000000-0000-0000-0000-000000000002"))
}

func usecase_dummyAdminID() d.UserID {
	return d.UserID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
}
