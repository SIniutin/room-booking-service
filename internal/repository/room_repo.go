package repository

import (
	"context"
	"time"

	d "github.com/SIniutin/room-booking-service/internal/domain"
)

type RoomRepo interface {
	Create(ctx context.Context, room d.Room) (d.Room, error)
	FindAll(ctx context.Context) ([]d.Room, error)
	FindByID(ctx context.Context, id d.RoomID) (d.Room, error)
}

type ScheduleRepo interface {
	Create(ctx context.Context, schedule d.Schedule) (d.Schedule, error)
	FindByRoomID(ctx context.Context, roomID d.RoomID) (d.Schedule, error)
	FindAll(ctx context.Context) ([]d.Schedule, error)
}

type SlotRepo interface {
	FindByRoomAndDate(ctx context.Context, roomID d.RoomID, date time.Time) ([]d.Slot, error)
	FindByID(ctx context.Context, id d.SlotID) (d.Slot, error)
	FindByIDs(ctx context.Context, id []d.SlotID) (map[d.SlotID]d.Slot, error)
	BulkCreate(ctx context.Context, slots []d.Slot) error
}
