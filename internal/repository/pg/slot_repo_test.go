package pg

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	d "github.com/SIniutin/room-booking-service/internal/domain"
)

func TestSlotRepo_BulkCreateFindByIDFindByIDsFindByRoomAndDate(t *testing.T) {
	pool := setupTestDB(t)
	roomRepo := NewRoomRepo(pool)
	slotRepo := NewSlotRepo(pool)

	room := d.Room{ID: d.RoomID(uuid.New()), Name: "Board Room"}
	_, err := roomRepo.Create(context.Background(), room)
	require.NoError(t, err)

	day := time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC)
	slot1 := d.Slot{ID: d.SlotID(uuid.New()), RoomID: room.ID, Start: day.Add(9 * time.Hour), End: day.Add(9*time.Hour + 30*time.Minute)}
	slot2 := d.Slot{ID: d.SlotID(uuid.New()), RoomID: room.ID, Start: day.Add(10 * time.Hour), End: day.Add(10*time.Hour + 30*time.Minute)}

	err = slotRepo.BulkCreate(context.Background(), []d.Slot{slot1, slot2})
	require.NoError(t, err)

	found, err := slotRepo.FindByID(context.Background(), slot1.ID)
	require.NoError(t, err)
	assert.Equal(t, slot1.Start, found.Start)

	foundMap, err := slotRepo.FindByIDs(context.Background(), []d.SlotID{slot1.ID, slot2.ID})
	require.NoError(t, err)
	require.Len(t, foundMap, 2)

	foundList, err := slotRepo.FindByRoomAndDate(context.Background(), room.ID, day)
	require.NoError(t, err)
	require.Len(t, foundList, 2)
}

func TestSlotRepo_FindByIDNotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewSlotRepo(pool)

	_, err := repo.FindByID(context.Background(), d.SlotID(uuid.New()))

	assert.ErrorIs(t, err, d.ErrSlotNotFound)
}
