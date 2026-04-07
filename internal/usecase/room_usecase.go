package usecase

import (
	"context"
	"time"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
)

type RoomUsecase interface {
	CreateRoom(ctx context.Context, name string, description *string, capacity *int) (d.Room, error)
	ListRooms(ctx context.Context) ([]d.Room, error)
	CreateSchedule(ctx context.Context, roomID d.RoomID, days []int, startTime, endTime string) (d.Schedule, error)
	ListSlots(ctx context.Context, roomID d.RoomID, date time.Time) ([]d.Slot, error)
}
