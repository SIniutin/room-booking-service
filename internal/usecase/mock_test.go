package usecase_test

import (
	"context"
	"time"

	d "github.com/SIniutin/room-booking-service/internal/domain"
)

type mockBookingRepo struct {
	create       func(ctx context.Context, b d.Booking) (d.Booking, error)
	findByID     func(ctx context.Context, id d.BookingID) (d.Booking, error)
	findAll      func(ctx context.Context, page, pageSize int) ([]d.Booking, int, error)
	findByUserID func(ctx context.Context, userID d.UserID) ([]d.Booking, error)
	updateStatus func(ctx context.Context, id d.BookingID, status d.BookingStatus) (d.Booking, error)
}

func (m *mockBookingRepo) Create(ctx context.Context, b d.Booking) (d.Booking, error) {
	return m.create(ctx, b)
}
func (m *mockBookingRepo) FindByID(ctx context.Context, id d.BookingID) (d.Booking, error) {
	return m.findByID(ctx, id)
}
func (m *mockBookingRepo) FindAll(ctx context.Context, page, pageSize int) ([]d.Booking, int, error) {
	return m.findAll(ctx, page, pageSize)
}
func (m *mockBookingRepo) FindByUserID(ctx context.Context, userID d.UserID) ([]d.Booking, error) {
	return m.findByUserID(ctx, userID)
}
func (m *mockBookingRepo) UpdateStatus(ctx context.Context, id d.BookingID, status d.BookingStatus) (d.Booking, error) {
	return m.updateStatus(ctx, id, status)
}

type mockSlotRepo struct {
	findByRoomAndDate func(ctx context.Context, roomID d.RoomID, date time.Time) ([]d.Slot, error)
	findByID          func(ctx context.Context, id d.SlotID) (d.Slot, error)
	findByIDs         func(ctx context.Context, ids []d.SlotID) (map[d.SlotID]d.Slot, error)
	bulkCreate        func(ctx context.Context, slots []d.Slot) error
}

func (m *mockSlotRepo) FindByRoomAndDate(ctx context.Context, roomID d.RoomID, date time.Time) ([]d.Slot, error) {
	return m.findByRoomAndDate(ctx, roomID, date)
}
func (m *mockSlotRepo) FindByID(ctx context.Context, id d.SlotID) (d.Slot, error) {
	return m.findByID(ctx, id)
}
func (m *mockSlotRepo) FindByIDs(ctx context.Context, ids []d.SlotID) (map[d.SlotID]d.Slot, error) {
	return m.findByIDs(ctx, ids)
}
func (m *mockSlotRepo) BulkCreate(ctx context.Context, slots []d.Slot) error {
	return m.bulkCreate(ctx, slots)
}

type mockRoomRepo struct {
	create   func(ctx context.Context, room d.Room) (d.Room, error)
	findAll  func(ctx context.Context) ([]d.Room, error)
	findByID func(ctx context.Context, id d.RoomID) (d.Room, error)
}

func (m *mockRoomRepo) Create(ctx context.Context, room d.Room) (d.Room, error) {
	return m.create(ctx, room)
}
func (m *mockRoomRepo) FindAll(ctx context.Context) ([]d.Room, error) {
	return m.findAll(ctx)
}
func (m *mockRoomRepo) FindByID(ctx context.Context, id d.RoomID) (d.Room, error) {
	return m.findByID(ctx, id)
}

// ── mockScheduleRepo ──────────────────────────────────────────────────────────

type mockScheduleRepo struct {
	create       func(ctx context.Context, s d.Schedule) (d.Schedule, error)
	findByRoomID func(ctx context.Context, roomID d.RoomID) (d.Schedule, error)
	findAll      func(ctx context.Context) ([]d.Schedule, error)
}

func (m *mockScheduleRepo) Create(ctx context.Context, s d.Schedule) (d.Schedule, error) {
	return m.create(ctx, s)
}
func (m *mockScheduleRepo) FindByRoomID(ctx context.Context, roomID d.RoomID) (d.Schedule, error) {
	return m.findByRoomID(ctx, roomID)
}
func (m *mockScheduleRepo) FindAll(ctx context.Context) ([]d.Schedule, error) {
	return m.findAll(ctx)
}

type mockUserRepo struct {
	create         func(ctx context.Context, user d.User, creds d.Credentials) (d.User, error)
	findByEmail    func(ctx context.Context, email d.Email) (d.User, error)
	findByID       func(ctx context.Context, id d.UserID) (d.User, error)
	getCredentials func(ctx context.Context, id d.UserID) (d.Credentials, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user d.User, creds d.Credentials) (d.User, error) {
	if m.create == nil {
		return user, nil
	}
	return m.create(ctx, user, creds)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email d.Email) (d.User, error) {
	if m.findByEmail == nil {
		return d.User{}, d.ErrUserNotFound
	}
	return m.findByEmail(ctx, email)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id d.UserID) (d.User, error) {
	if m.findByID == nil {
		return d.User{}, d.ErrUserNotFound
	}
	return m.findByID(ctx, id)
}
func (m *mockUserRepo) GetCredentials(ctx context.Context, id d.UserID) (d.Credentials, error) {
	if m.getCredentials == nil {
		return d.Credentials{}, d.ErrUserNotFound
	}
	return m.getCredentials(ctx, id)
}

type mockCache struct {
	get func(ctx context.Context, key string) (string, error)
	set func(ctx context.Context, key string, value any, ttl time.Duration) error
	del func(ctx context.Context, key string) error
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	if m.get != nil {
		return m.get(ctx, key)
	}
	return "", nil
}
func (m *mockCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if m.set != nil {
		return m.set(ctx, key, value, ttl)
	}
	return nil
}
func (m *mockCache) Del(ctx context.Context, key string) error {
	if m.del != nil {
		return m.del(ctx, key)
	}
	return nil
}

func (m *mockCache) Close() error {
	return nil
}
