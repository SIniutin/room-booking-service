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

var (
	testUserID    = d.UserID(uuid.New())
	testSlotID    = d.SlotID(uuid.New())
	testBookingID = d.BookingID(uuid.New())
	testRoomID    = d.RoomID(uuid.New())
	futureSlot    = d.Slot{
		ID:     testSlotID,
		RoomID: testRoomID,
		Start:  time.Now().Add(24 * time.Hour),
		End:    time.Now().Add(25 * time.Hour),
	}
	pastSlot = d.Slot{
		ID:     testSlotID,
		RoomID: testRoomID,
		Start:  time.Now().Add(-2 * time.Hour),
		End:    time.Now().Add(-1 * time.Hour),
	}
	activeBooking = d.Booking{
		ID:     testBookingID,
		SlotID: testSlotID,
		UserID: testUserID,
		Status: d.ActiveBooking,
	}
	cancelledBooking = d.Booking{
		ID:     testBookingID,
		SlotID: testSlotID,
		UserID: testUserID,
		Status: d.CancelledBooking,
	}
)

func newBookingUsecase(bookRepo *mockBookingRepo, slotRepo *mockSlotRepo) usecase.BookingUsecase {
	return usecase.NewBookingUsecase(bookRepo, slotRepo, &mockCache{})
}

// ── CreateBooking ─────────────────────────────────────────────────────────────

func TestCreateBooking_Success(t *testing.T) {
	slotRepo := &mockSlotRepo{
		findByID: func(_ context.Context, _ d.SlotID) (d.Slot, error) {
			return futureSlot, nil
		},
	}
	bookRepo := &mockBookingRepo{
		create: func(_ context.Context, b d.Booking) (d.Booking, error) {
			b.ID = testBookingID
			return b, nil
		},
	}

	uc := newBookingUsecase(bookRepo, slotRepo)
	got, err := uc.CreateBooking(context.Background(), testUserID, testSlotID, false)

	require.NoError(t, err)
	assert.Equal(t, d.ActiveBooking, got.Status)
	assert.Equal(t, testSlotID, got.SlotID)
	assert.Equal(t, testUserID, got.UserID)
}

func TestCreateBooking_WithConferenceLink(t *testing.T) {
	slotRepo := &mockSlotRepo{
		findByID: func(_ context.Context, _ d.SlotID) (d.Slot, error) {
			return futureSlot, nil
		},
	}
	bookRepo := &mockBookingRepo{
		create: func(_ context.Context, b d.Booking) (d.Booking, error) {
			return b, nil
		},
	}

	uc := newBookingUsecase(bookRepo, slotRepo)
	got, err := uc.CreateBooking(context.Background(), testUserID, testSlotID, true)

	require.NoError(t, err)
	assert.NotEmpty(t, got.Link)
}

func TestCreateBooking_InvalidatesSlotCache(t *testing.T) {
	expectedKey := "route:/rooms/" + uuid.UUID(testRoomID).String() + "/slots/list?date=" + futureSlot.Start.UTC().Format("2006-01-02")

	slotRepo := &mockSlotRepo{
		findByID: func(_ context.Context, _ d.SlotID) (d.Slot, error) {
			return futureSlot, nil
		},
	}
	bookRepo := &mockBookingRepo{
		create: func(_ context.Context, b d.Booking) (d.Booking, error) {
			return b, nil
		},
	}

	var deletedKey string
	uc := usecase.NewBookingUsecase(bookRepo, slotRepo, &mockCache{
		del: func(_ context.Context, key string) error {
			deletedKey = key
			return nil
		},
	})

	_, err := uc.CreateBooking(context.Background(), testUserID, testSlotID, false)

	require.NoError(t, err)
	assert.Equal(t, expectedKey, deletedKey)
}

func TestCreateBooking_SlotNotFound(t *testing.T) {
	slotRepo := &mockSlotRepo{
		findByID: func(_ context.Context, _ d.SlotID) (d.Slot, error) {
			return d.Slot{}, d.ErrSlotNotFound
		},
	}

	uc := newBookingUsecase(&mockBookingRepo{}, slotRepo)
	_, err := uc.CreateBooking(context.Background(), testUserID, testSlotID, false)

	assert.ErrorIs(t, err, d.ErrSlotNotFound)
}

func TestCreateBooking_SlotInThePast(t *testing.T) {
	slotRepo := &mockSlotRepo{
		findByID: func(_ context.Context, _ d.SlotID) (d.Slot, error) {
			return pastSlot, nil
		},
	}

	uc := newBookingUsecase(&mockBookingRepo{}, slotRepo)
	_, err := uc.CreateBooking(context.Background(), testUserID, testSlotID, false)

	assert.ErrorIs(t, err, d.ErrSlotInThePast)
}

func TestCreateBooking_SlotAlreadyBooked(t *testing.T) {
	slotRepo := &mockSlotRepo{
		findByID: func(_ context.Context, _ d.SlotID) (d.Slot, error) {
			return futureSlot, nil
		},
	}
	bookRepo := &mockBookingRepo{
		create: func(_ context.Context, _ d.Booking) (d.Booking, error) {
			return d.Booking{}, d.ErrSlotAlreadyBooked
		},
	}

	uc := newBookingUsecase(bookRepo, slotRepo)
	_, err := uc.CreateBooking(context.Background(), testUserID, testSlotID, false)

	assert.ErrorIs(t, err, d.ErrSlotAlreadyBooked)
}

func TestCreateBooking_RepoError(t *testing.T) {
	repoErr := errors.New("create failed")
	slotRepo := &mockSlotRepo{
		findByID: func(_ context.Context, _ d.SlotID) (d.Slot, error) {
			return futureSlot, nil
		},
	}
	bookRepo := &mockBookingRepo{
		create: func(_ context.Context, _ d.Booking) (d.Booking, error) {
			return d.Booking{}, repoErr
		},
	}

	uc := newBookingUsecase(bookRepo, slotRepo)
	_, err := uc.CreateBooking(context.Background(), testUserID, testSlotID, false)

	assert.ErrorIs(t, err, repoErr)
}

// ── CancelBooking ─────────────────────────────────────────────────────────────

func TestCancelBooking_Success(t *testing.T) {
	updateCalled := false
	slotRepo := &mockSlotRepo{
		findByID: func(_ context.Context, _ d.SlotID) (d.Slot, error) {
			return futureSlot, nil
		},
	}
	bookRepo := &mockBookingRepo{
		findByID: func(_ context.Context, _ d.BookingID) (d.Booking, error) {
			return activeBooking, nil
		},
		updateStatus: func(_ context.Context, _ d.BookingID, status d.BookingStatus) (d.Booking, error) {
			updateCalled = true
			b := activeBooking
			b.Status = status
			return b, nil
		},
	}

	uc := newBookingUsecase(bookRepo, slotRepo)
	got, err := uc.CancelBooking(context.Background(), testUserID, testBookingID)

	require.NoError(t, err)
	assert.True(t, updateCalled)
	assert.Equal(t, d.CancelledBooking, got.Status)
}

func TestCancelBooking_NotFound(t *testing.T) {
	bookRepo := &mockBookingRepo{
		findByID: func(_ context.Context, _ d.BookingID) (d.Booking, error) {
			return d.Booking{}, d.ErrBookingNotFound
		},
	}

	uc := newBookingUsecase(bookRepo, &mockSlotRepo{})
	_, err := uc.CancelBooking(context.Background(), testUserID, testBookingID)

	assert.ErrorIs(t, err, d.ErrBookingNotFound)
}

func TestCancelBooking_Forbidden(t *testing.T) {
	otherUserID := d.UserID(uuid.New())
	bookRepo := &mockBookingRepo{
		findByID: func(_ context.Context, _ d.BookingID) (d.Booking, error) {
			return activeBooking, nil // activeBooking.UserID = testUserID
		},
	}

	uc := newBookingUsecase(bookRepo, &mockSlotRepo{})
	_, err := uc.CancelBooking(context.Background(), otherUserID, testBookingID)

	assert.ErrorIs(t, err, d.ErrBookingForbidden)
}

func TestCancelBooking_AlreadyCancelled_Idempotent(t *testing.T) {
	updateCalled := false
	bookRepo := &mockBookingRepo{
		findByID: func(_ context.Context, _ d.BookingID) (d.Booking, error) {
			return cancelledBooking, nil
		},
		updateStatus: func(_ context.Context, _ d.BookingID, _ d.BookingStatus) (d.Booking, error) {
			updateCalled = true
			return d.Booking{}, nil
		},
	}

	uc := newBookingUsecase(bookRepo, &mockSlotRepo{})
	got, err := uc.CancelBooking(context.Background(), testUserID, testBookingID)

	require.NoError(t, err)
	assert.False(t, updateCalled, "UpdateStatus must not be called for already-cancelled booking")
	assert.Equal(t, d.CancelledBooking, got.Status)
}

// ── ListMyBookings ────────────────────────────────────────────────────────────

func TestListMyBookings_ReturnsFutureOnly(t *testing.T) {
	futureSlotID := d.SlotID(uuid.New())
	pastSlotID := d.SlotID(uuid.New())

	futureBooking := d.Booking{ID: d.BookingID(uuid.New()), SlotID: futureSlotID, UserID: testUserID, Status: d.ActiveBooking}
	pastBooking := d.Booking{ID: d.BookingID(uuid.New()), SlotID: pastSlotID, UserID: testUserID, Status: d.ActiveBooking}

	bookRepo := &mockBookingRepo{
		findByUserID: func(_ context.Context, _ d.UserID) ([]d.Booking, error) {
			return []d.Booking{futureBooking, pastBooking}, nil
		},
	}
	slotRepo := &mockSlotRepo{
		findByIDs: func(_ context.Context, ids []d.SlotID) (map[d.SlotID]d.Slot, error) {
			return map[d.SlotID]d.Slot{
				futureSlotID: {ID: futureSlotID, Start: time.Now().Add(time.Hour)},
				pastSlotID:   {ID: pastSlotID, Start: time.Now().Add(-time.Hour)},
			}, nil
		},
	}

	uc := newBookingUsecase(bookRepo, slotRepo)
	got, err := uc.ListMyBookings(context.Background(), testUserID)

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, futureSlotID, got[0].SlotID)
}

func TestListMyBookings_Empty(t *testing.T) {
	bookRepo := &mockBookingRepo{
		findByUserID: func(_ context.Context, _ d.UserID) ([]d.Booking, error) {
			return []d.Booking{}, nil
		},
	}

	uc := newBookingUsecase(bookRepo, &mockSlotRepo{})
	got, err := uc.ListMyBookings(context.Background(), testUserID)

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestListMyBookings_RepoError(t *testing.T) {
	repoErr := errors.New("find by user failed")
	bookRepo := &mockBookingRepo{
		findByUserID: func(_ context.Context, _ d.UserID) ([]d.Booking, error) {
			return nil, repoErr
		},
	}

	uc := newBookingUsecase(bookRepo, &mockSlotRepo{})
	_, err := uc.ListMyBookings(context.Background(), testUserID)

	assert.ErrorIs(t, err, repoErr)
}

func TestListMyBookings_SlotRepoError(t *testing.T) {
	repoErr := errors.New("find slots failed")
	bookRepo := &mockBookingRepo{
		findByUserID: func(_ context.Context, _ d.UserID) ([]d.Booking, error) {
			return []d.Booking{{ID: testBookingID, SlotID: testSlotID, UserID: testUserID, Status: d.ActiveBooking}}, nil
		},
	}
	slotRepo := &mockSlotRepo{
		findByIDs: func(_ context.Context, _ []d.SlotID) (map[d.SlotID]d.Slot, error) {
			return nil, repoErr
		},
	}

	uc := newBookingUsecase(bookRepo, slotRepo)
	_, err := uc.ListMyBookings(context.Background(), testUserID)

	assert.ErrorIs(t, err, repoErr)
}

func TestListBookings_DelegatesToRepo(t *testing.T) {
	expected := []d.Booking{{ID: testBookingID, SlotID: testSlotID, UserID: testUserID, Status: d.ActiveBooking}}
	bookRepo := &mockBookingRepo{
		findAll: func(_ context.Context, page, pageSize int) ([]d.Booking, int, error) {
			assert.Equal(t, 2, page)
			assert.Equal(t, 5, pageSize)
			return expected, 11, nil
		},
	}

	uc := newBookingUsecase(bookRepo, &mockSlotRepo{})
	got, total, err := uc.ListBookings(context.Background(), 2, 5)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	assert.Equal(t, 11, total)
}
