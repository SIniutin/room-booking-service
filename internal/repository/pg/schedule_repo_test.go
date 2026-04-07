package pg

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
)

func TestScheduleRepo_CreateFindByRoomIDFindAll(t *testing.T) {
	pool := setupTestDB(t)
	roomRepo := NewRoomRepo(pool)
	scheduleRepo := NewScheduleRepo(pool)

	room := d.Room{ID: d.RoomID(uuid.New()), Name: "Board Room"}
	_, err := roomRepo.Create(context.Background(), room)
	require.NoError(t, err)

	start, err := d.NewWorkTime("09:00")
	require.NoError(t, err)
	end, err := d.NewWorkTime("18:00")
	require.NoError(t, err)

	schedule := d.Schedule{
		ID:     d.ScheduleID(uuid.New()),
		RoomID: room.ID,
		Days:   d.Monday | d.Wednesday,
		Start:  *start,
		End:    *end,
	}

	created, err := scheduleRepo.Create(context.Background(), schedule)
	require.NoError(t, err)
	assert.Equal(t, room.ID, created.RoomID)

	found, err := scheduleRepo.FindByRoomID(context.Background(), room.ID)
	require.NoError(t, err)
	assert.Equal(t, schedule.Days, found.Days)

	all, err := scheduleRepo.FindAll(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, room.ID, all[0].RoomID)
}

func TestScheduleRepo_FindByRoomIDNotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewScheduleRepo(pool)

	_, err := repo.FindByRoomID(context.Background(), d.RoomID(uuid.New()))

	assert.ErrorIs(t, err, d.ErrScheduleNotFound)
}
