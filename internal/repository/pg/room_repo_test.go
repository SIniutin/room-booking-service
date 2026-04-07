package pg

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
)

func TestRoomRepo_CreateFindByIDFindAll(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewRoomRepo(pool)

	description := "Main room"
	capacity := 10
	room := d.Room{
		ID:          d.RoomID(uuid.New()),
		Name:        "Board Room",
		Description: &description,
		Capacity:    &capacity,
	}

	created, err := repo.Create(context.Background(), room)
	require.NoError(t, err)
	assert.Equal(t, room.Name, created.Name)

	found, err := repo.FindByID(context.Background(), room.ID)
	require.NoError(t, err)
	assert.Equal(t, room.ID, found.ID)

	all, err := repo.FindAll(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, room.Name, all[0].Name)
}

func TestRoomRepo_FindByIDNotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewRoomRepo(pool)

	_, err := repo.FindByID(context.Background(), d.RoomID(uuid.New()))

	assert.ErrorIs(t, err, d.ErrRoomNotFound)
}
