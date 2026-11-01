package pg

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	d "github.com/SIniutin/room-booking-service/internal/domain"
)

func TestUserRepo_CreateAndRead(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewUserRepo(pool)

	email, err := d.NewEmail("user@example.com")
	require.NoError(t, err)

	user := d.User{
		ID:    d.UserID(uuid.New()),
		Email: *email,
		Role:  d.RoleUser,
	}
	creds := d.Credentials{
		UserID:       user.ID,
		PasswordHash: "hash",
	}

	created, err := repo.Create(context.Background(), user, creds)
	require.NoError(t, err)
	assert.Equal(t, user.ID, created.ID)

	foundByEmail, err := repo.FindByEmail(context.Background(), *email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundByEmail.ID)

	foundByID, err := repo.FindByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, *email, foundByID.Email)

	gotCreds, err := repo.GetCredentials(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, "hash", gotCreds.PasswordHash)
}

func TestUserRepo_CreateDuplicateEmail(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewUserRepo(pool)

	email, err := d.NewEmail("user@example.com")
	require.NoError(t, err)

	user1 := d.User{ID: d.UserID(uuid.New()), Email: *email, Role: d.RoleUser}
	user2 := d.User{ID: d.UserID(uuid.New()), Email: *email, Role: d.RoleAdmin}

	_, err = repo.Create(context.Background(), user1, d.Credentials{UserID: user1.ID, PasswordHash: "hash1"})
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), user2, d.Credentials{UserID: user2.ID, PasswordHash: "hash2"})
	assert.ErrorIs(t, err, d.ErrUserAlreadyExists)
}
