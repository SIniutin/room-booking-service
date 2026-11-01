package pg

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	d "github.com/SIniutin/room-booking-service/internal/domain"
)

func TestBookingRepo_CreateFindAllFindByUserUpdateStatus(t *testing.T) {
	pool := setupTestDB(t)

	userRepo := NewUserRepo(pool)
	roomRepo := NewRoomRepo(pool)
	slotRepo := NewSlotRepo(pool)
	bookingRepo := NewBookingRepo(pool)

	email, err := d.NewEmail("user@example.com")
	require.NoError(t, err)
	user := d.User{ID: d.UserID(uuid.New()), Email: *email, Role: d.RoleUser}
	_, err = userRepo.Create(context.Background(), user, d.Credentials{UserID: user.ID, PasswordHash: "hash"})
	require.NoError(t, err)

	room := d.Room{ID: d.RoomID(uuid.New()), Name: "Board Room"}
	_, err = roomRepo.Create(context.Background(), room)
	require.NoError(t, err)

	start := time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC)
	slot := d.Slot{ID: d.SlotID(uuid.New()), RoomID: room.ID, Start: start, End: start.Add(30 * time.Minute)}
	err = slotRepo.BulkCreate(context.Background(), []d.Slot{slot})
	require.NoError(t, err)

	booking := d.Booking{ID: d.BookingID(uuid.New()), SlotID: slot.ID, UserID: user.ID, Status: d.ActiveBooking}
	created, err := bookingRepo.Create(context.Background(), booking)
	require.NoError(t, err)
	assert.Equal(t, d.ActiveBooking, created.Status)

	found, err := bookingRepo.FindByID(context.Background(), booking.ID)
	require.NoError(t, err)
	assert.Equal(t, booking.ID, found.ID)

	all, total, err := bookingRepo.FindAll(context.Background(), 1, 10)
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, 1, total)

	byUser, err := bookingRepo.FindByUserID(context.Background(), user.ID)
	require.NoError(t, err)
	require.Len(t, byUser, 1)

	updated, err := bookingRepo.UpdateStatus(context.Background(), booking.ID, d.CancelledBooking)
	require.NoError(t, err)
	assert.Equal(t, d.CancelledBooking, updated.Status)
}

func TestBookingRepo_CreateDuplicateActiveBooking(t *testing.T) {
	pool := setupTestDB(t)

	userRepo := NewUserRepo(pool)
	roomRepo := NewRoomRepo(pool)
	slotRepo := NewSlotRepo(pool)
	bookingRepo := NewBookingRepo(pool)

	email1, _ := d.NewEmail("user1@example.com")
	email2, _ := d.NewEmail("user2@example.com")
	user1 := d.User{ID: d.UserID(uuid.New()), Email: *email1, Role: d.RoleUser}
	user2 := d.User{ID: d.UserID(uuid.New()), Email: *email2, Role: d.RoleUser}
	_, err := userRepo.Create(context.Background(), user1, d.Credentials{UserID: user1.ID, PasswordHash: "hash1"})
	require.NoError(t, err)
	_, err = userRepo.Create(context.Background(), user2, d.Credentials{UserID: user2.ID, PasswordHash: "hash2"})
	require.NoError(t, err)

	room := d.Room{ID: d.RoomID(uuid.New()), Name: "Board Room"}
	_, err = roomRepo.Create(context.Background(), room)
	require.NoError(t, err)

	start := time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC)
	slot := d.Slot{ID: d.SlotID(uuid.New()), RoomID: room.ID, Start: start, End: start.Add(30 * time.Minute)}
	err = slotRepo.BulkCreate(context.Background(), []d.Slot{slot})
	require.NoError(t, err)

	_, err = bookingRepo.Create(context.Background(), d.Booking{ID: d.BookingID(uuid.New()), SlotID: slot.ID, UserID: user1.ID, Status: d.ActiveBooking})
	require.NoError(t, err)

	_, err = bookingRepo.Create(context.Background(), d.Booking{ID: d.BookingID(uuid.New()), SlotID: slot.ID, UserID: user2.ID, Status: d.ActiveBooking})
	assert.ErrorIs(t, err, d.ErrSlotAlreadyBooked)
}

func TestBookingRepo_ConcurrentCreateAllowsOnlyOneActiveBooking(t *testing.T) {
	pool := setupTestDB(t)

	userRepo := NewUserRepo(pool)
	roomRepo := NewRoomRepo(pool)
	slotRepo := NewSlotRepo(pool)
	bookingRepo := NewBookingRepo(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users := make([]d.User, 2)
	for i := range users {
		email, err := d.NewEmail("concurrent-user-" + uuid.NewString() + "@example.com")
		require.NoError(t, err)
		users[i] = d.User{ID: d.UserID(uuid.New()), Email: *email, Role: d.RoleUser}
		_, err = userRepo.Create(ctx, users[i], d.Credentials{UserID: users[i].ID, PasswordHash: "hash"})
		require.NoError(t, err)
	}

	room := d.Room{ID: d.RoomID(uuid.New()), Name: "Concurrent Room"}
	_, err := roomRepo.Create(ctx, room)
	require.NoError(t, err)

	start := time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)
	slot := d.Slot{ID: d.SlotID(uuid.New()), RoomID: room.ID, Start: start, End: start.Add(30 * time.Minute)}
	require.NoError(t, slotRepo.BulkCreate(ctx, []d.Slot{slot}))

	startRequests := make(chan struct{})
	results := make(chan error, len(users))
	var ready sync.WaitGroup
	ready.Add(len(users))

	for _, user := range users {
		user := user
		go func() {
			ready.Done()
			<-startRequests
			_, err := bookingRepo.Create(ctx, d.Booking{
				ID:     d.BookingID(uuid.New()),
				SlotID: slot.ID,
				UserID: user.ID,
				Status: d.ActiveBooking,
			})
			results <- err
		}()
	}

	ready.Wait()
	close(startRequests)

	var succeeded, rejected int
	for range users {
		err := <-results
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, d.ErrSlotAlreadyBooked):
			rejected++
		default:
			require.NoError(t, err)
		}
	}

	assert.Equal(t, 1, succeeded, "the database must accept exactly one active booking")
	assert.Equal(t, 1, rejected, "the competing insert must hit the unique constraint")

	var activeBookings int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings WHERE slot_id = $1 AND status = 'active'`,
		uuid.UUID(slot.ID),
	).Scan(&activeBookings)
	require.NoError(t, err)
	assert.Equal(t, 1, activeBookings)
}
