package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	"github.com/internships-backend/test-backend-SIniutin/internal/usecase"
)

func newRoomUsecase(roomRepo *mockRoomRepo, schedRepo *mockScheduleRepo, slotRepo *mockSlotRepo) usecase.RoomUsecase {
	return usecase.NewRoomUsecase(roomRepo, schedRepo, slotRepo, usecase.NewSlotGenerator(schedRepo, slotRepo))
}

var testRoom = d.Room{ID: testRoomID, Name: "Room A"}

func TestCreateSchedule_Success(t *testing.T) {
	var createdSlots []d.Slot

	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return testRoom, nil
		},
	}
	schedRepo := &mockScheduleRepo{
		findByRoomID: func(_ context.Context, _ d.RoomID) (d.Schedule, error) {
			return d.Schedule{}, d.ErrScheduleNotFound
		},
		create: func(_ context.Context, s d.Schedule) (d.Schedule, error) {
			return s, nil
		},
	}
	slotRepo := &mockSlotRepo{
		bulkCreate: func(_ context.Context, slots []d.Slot) error {
			createdSlots = append(createdSlots, slots...)
			return nil
		},
	}

	uc := newRoomUsecase(roomRepo, schedRepo, slotRepo)
	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	dayISO := int(tomorrow.Weekday())
	if dayISO == 0 {
		dayISO = 7
	}

	got, err := uc.CreateSchedule(context.Background(), testRoomID, []int{dayISO}, "09:00", "18:00")

	require.NoError(t, err)
	assert.Equal(t, testRoomID, got.RoomID)
	assert.NotEmpty(t, createdSlots)
	for _, slot := range createdSlots {
		assert.Equal(t, testRoomID, slot.RoomID)
		assert.True(t, slot.Start.Before(slot.End))
	}
}

func TestCreateSchedule_RoomNotFound(t *testing.T) {
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return d.Room{}, d.ErrRoomNotFound
		},
	}

	uc := newRoomUsecase(roomRepo, &mockScheduleRepo{}, &mockSlotRepo{})
	_, err := uc.CreateSchedule(context.Background(), testRoomID, []int{1}, "09:00", "18:00")

	assert.ErrorIs(t, err, d.ErrRoomNotFound)
}

func TestCreateSchedule_AlreadyExists(t *testing.T) {
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return testRoom, nil
		},
	}
	schedRepo := &mockScheduleRepo{
		findByRoomID: func(_ context.Context, _ d.RoomID) (d.Schedule, error) {
			return d.Schedule{ID: d.ScheduleID(uuid.New()), RoomID: testRoomID}, nil
		},
	}

	uc := newRoomUsecase(roomRepo, schedRepo, &mockSlotRepo{})
	_, err := uc.CreateSchedule(context.Background(), testRoomID, []int{1}, "09:00", "18:00")

	assert.ErrorIs(t, err, d.ErrScheduleAlreadyExists)
}

func TestCreateSchedule_InvalidTime_StartAfterEnd(t *testing.T) {
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return testRoom, nil
		},
	}
	schedRepo := &mockScheduleRepo{
		findByRoomID: func(_ context.Context, _ d.RoomID) (d.Schedule, error) {
			return d.Schedule{}, d.ErrScheduleNotFound
		},
	}

	uc := newRoomUsecase(roomRepo, schedRepo, &mockSlotRepo{})
	_, err := uc.CreateSchedule(context.Background(), testRoomID, []int{1}, "18:00", "09:00")

	assert.ErrorIs(t, err, d.ErrInvalidScheduleTime)
}

func TestCreateSchedule_InvalidTime_EqualStartEnd(t *testing.T) {
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return testRoom, nil
		},
	}
	schedRepo := &mockScheduleRepo{
		findByRoomID: func(_ context.Context, _ d.RoomID) (d.Schedule, error) {
			return d.Schedule{}, d.ErrScheduleNotFound
		},
	}

	uc := newRoomUsecase(roomRepo, schedRepo, &mockSlotRepo{})
	_, err := uc.CreateSchedule(context.Background(), testRoomID, []int{1}, "09:00", "09:00")

	assert.ErrorIs(t, err, d.ErrInvalidScheduleTime)
}

func TestCreateSchedule_InvalidWorkday(t *testing.T) {
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return testRoom, nil
		},
	}
	schedRepo := &mockScheduleRepo{
		findByRoomID: func(_ context.Context, _ d.RoomID) (d.Schedule, error) {
			return d.Schedule{}, d.ErrScheduleNotFound
		},
	}

	uc := newRoomUsecase(roomRepo, schedRepo, &mockSlotRepo{})
	_, err := uc.CreateSchedule(context.Background(), testRoomID, []int{0}, "09:00", "18:00") // day 0 invalid

	assert.ErrorIs(t, err, d.ErrInvalidWorkday)
}

// ── ListSlots ─────────────────────────────────────────────────────────────────

func TestListSlots_Success(t *testing.T) {
	date := time.Now().UTC().Truncate(24 * time.Hour)
	slots := []d.Slot{
		{ID: d.SlotID(uuid.New()), RoomID: testRoomID, Start: date.Add(9 * time.Hour)},
		{ID: d.SlotID(uuid.New()), RoomID: testRoomID, Start: date.Add(10 * time.Hour)},
	}

	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return testRoom, nil
		},
	}
	slotRepo := &mockSlotRepo{
		findByRoomAndDate: func(_ context.Context, _ d.RoomID, _ time.Time) ([]d.Slot, error) {
			return slots, nil
		},
	}

	uc := newRoomUsecase(roomRepo, &mockScheduleRepo{}, slotRepo)
	got, err := uc.ListSlots(context.Background(), testRoomID, date)

	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestListSlots_RoomNotFound(t *testing.T) {
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return d.Room{}, d.ErrRoomNotFound
		},
	}

	uc := newRoomUsecase(roomRepo, &mockScheduleRepo{}, &mockSlotRepo{})
	_, err := uc.ListSlots(context.Background(), testRoomID, time.Now())

	assert.ErrorIs(t, err, d.ErrRoomNotFound)
}

// ── CreateRoom ────────────────────────────────────────────────────────────────

func TestCreateRoom_Success(t *testing.T) {
	roomRepo := &mockRoomRepo{
		create: func(_ context.Context, room d.Room) (d.Room, error) {
			return room, nil
		},
	}

	uc := newRoomUsecase(roomRepo, &mockScheduleRepo{}, &mockSlotRepo{})
	name := "Board Room"
	got, err := uc.CreateRoom(context.Background(), name, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, name, got.Name)
}

func TestCreateSchedule_RoomRepoUnexpectedError(t *testing.T) {
	repoErr := errors.New("db unavailable")
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return d.Room{}, repoErr
		},
	}

	uc := newRoomUsecase(roomRepo, &mockScheduleRepo{}, &mockSlotRepo{})
	_, err := uc.CreateSchedule(context.Background(), testRoomID, []int{1}, "09:00", "18:00")

	assert.ErrorIs(t, err, repoErr)
}

func TestCreateSchedule_ScheduleRepoUnexpectedError(t *testing.T) {
	repoErr := errors.New("read schedule failed")
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return testRoom, nil
		},
	}
	schedRepo := &mockScheduleRepo{
		findByRoomID: func(_ context.Context, _ d.RoomID) (d.Schedule, error) {
			return d.Schedule{}, repoErr
		},
	}

	uc := newRoomUsecase(roomRepo, schedRepo, &mockSlotRepo{})
	_, err := uc.CreateSchedule(context.Background(), testRoomID, []int{1}, "09:00", "18:00")

	assert.ErrorIs(t, err, repoErr)
}

func TestCreateSchedule_BulkCreateError(t *testing.T) {
	repoErr := errors.New("bulk create failed")
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return testRoom, nil
		},
	}
	schedRepo := &mockScheduleRepo{
		findByRoomID: func(_ context.Context, _ d.RoomID) (d.Schedule, error) {
			return d.Schedule{}, d.ErrScheduleNotFound
		},
		create: func(_ context.Context, s d.Schedule) (d.Schedule, error) {
			return s, nil
		},
	}
	slotRepo := &mockSlotRepo{
		bulkCreate: func(_ context.Context, _ []d.Slot) error {
			return repoErr
		},
	}

	uc := newRoomUsecase(roomRepo, schedRepo, slotRepo)
	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	dayISO := int(tomorrow.Weekday())
	if dayISO == 0 {
		dayISO = 7
	}

	_, err := uc.CreateSchedule(context.Background(), testRoomID, []int{dayISO}, "09:00", "18:00")

	assert.ErrorIs(t, err, repoErr)
}

func TestListSlots_RepoError(t *testing.T) {
	repoErr := errors.New("list slots failed")
	roomRepo := &mockRoomRepo{
		findByID: func(_ context.Context, _ d.RoomID) (d.Room, error) {
			return testRoom, nil
		},
	}
	slotRepo := &mockSlotRepo{
		findByRoomAndDate: func(_ context.Context, _ d.RoomID, _ time.Time) ([]d.Slot, error) {
			return nil, repoErr
		},
	}

	uc := newRoomUsecase(roomRepo, &mockScheduleRepo{}, slotRepo)
	_, err := uc.ListSlots(context.Background(), testRoomID, time.Now())

	assert.ErrorIs(t, err, repoErr)
}

func TestListRooms_DelegatesToRepo(t *testing.T) {
	expected := []d.Room{{ID: testRoomID, Name: "Board Room"}}
	roomRepo := &mockRoomRepo{
		findAll: func(_ context.Context) ([]d.Room, error) {
			return expected, nil
		},
	}

	uc := newRoomUsecase(roomRepo, &mockScheduleRepo{}, &mockSlotRepo{})
	got, err := uc.ListRooms(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
