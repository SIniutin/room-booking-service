package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	r "github.com/internships-backend/test-backend-SIniutin/internal/repository"
)

type roomUsecaseImpl struct {
	roomRepo      r.RoomRepo
	scheduleRepo  r.ScheduleRepo
	slotRepo      r.SlotRepo
	slotGenerator *SlotGenerator
}

func NewRoomUsecase(
	roomRepo r.RoomRepo,
	scheduleRepo r.ScheduleRepo,
	slotRepo r.SlotRepo,
	slotGenerator *SlotGenerator,
) RoomUsecase {
	return &roomUsecaseImpl{
		roomRepo:      roomRepo,
		scheduleRepo:  scheduleRepo,
		slotRepo:      slotRepo,
		slotGenerator: slotGenerator,
	}
}

func (uc *roomUsecaseImpl) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (d.Room, error) {
	room := d.Room{
		ID:          d.RoomID(uuid.New()),
		Name:        name,
		Description: description,
		Capacity:    capacity,
	}
	return uc.roomRepo.Create(ctx, room)
}

func (uc *roomUsecaseImpl) ListRooms(ctx context.Context) ([]d.Room, error) {
	return uc.roomRepo.FindAll(ctx)
}

func (uc *roomUsecaseImpl) CreateSchedule(ctx context.Context, roomID d.RoomID, days []int, startTime, endTime string) (d.Schedule, error) {
	_, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, d.ErrRoomNotFound) {
			return d.Schedule{}, d.ErrRoomNotFound
		}
		return d.Schedule{}, err
	}

	_, err = uc.scheduleRepo.FindByRoomID(ctx, roomID)
	if err == nil {
		return d.Schedule{}, d.ErrScheduleAlreadyExists
	}
	if !errors.Is(err, d.ErrScheduleNotFound) {
		return d.Schedule{}, err
	}

	workdays, err := d.NewWorkday(days)
	if err != nil {
		return d.Schedule{}, err
	}

	start, err := d.NewWorkTime(startTime)
	if err != nil {
		return d.Schedule{}, err
	}

	end, err := d.NewWorkTime(endTime)
	if err != nil {
		return d.Schedule{}, err
	}

	if !time.Time(*start).Before(time.Time(*end)) {
		return d.Schedule{}, d.ErrInvalidScheduleTime
	}

	schedule := d.Schedule{
		ID:     d.ScheduleID(uuid.New()),
		RoomID: roomID,
		Days:   workdays,
		Start:  *start,
		End:    *end,
	}

	created, err := uc.scheduleRepo.Create(ctx, schedule)
	if err != nil {
		return d.Schedule{}, err
	}

	from := truncateToDay(time.Now().UTC())
	if err := uc.slotGenerator.GenerateForSchedule(ctx, created, from, from.AddDate(0, 0, 7)); err != nil {
		return d.Schedule{}, err
	}

	return created, nil
}

func (uc *roomUsecaseImpl) ListSlots(ctx context.Context, roomID d.RoomID, date time.Time) ([]d.Slot, error) {
	_, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return uc.slotRepo.FindByRoomAndDate(ctx, roomID, date)
}
