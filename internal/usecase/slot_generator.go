package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	r "github.com/internships-backend/test-backend-SIniutin/internal/repository"
)

const slotDuration = 30 * time.Minute

type SlotGenerator struct {
	scheduleRepo r.ScheduleRepo
	slotRepo     r.SlotRepo
}

func NewSlotGenerator(scheduleRepo r.ScheduleRepo, slotRepo r.SlotRepo) *SlotGenerator {
	return &SlotGenerator{scheduleRepo: scheduleRepo, slotRepo: slotRepo}
}

func (g *SlotGenerator) GenerateForSchedule(ctx context.Context, schedule d.Schedule, from, to time.Time) error {
	slots := buildSlots(schedule, from, to)
	return g.slotRepo.BulkCreate(ctx, slots)
}

func (g *SlotGenerator) GenerateAll(ctx context.Context, from time.Time, days int) error {
	schedules, err := g.scheduleRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	to := from.AddDate(0, 0, days)
	for _, s := range schedules {
		slots := buildSlots(s, from, to)
		if err := g.slotRepo.BulkCreate(ctx, slots); err != nil {
			return err
		}
	}
	return nil
}

func buildSlots(schedule d.Schedule, from, to time.Time) []d.Slot {
	startHour := time.Time(schedule.Start).Hour()
	startMin := time.Time(schedule.Start).Minute()
	endHour := time.Time(schedule.End).Hour()
	endMin := time.Time(schedule.End).Minute()

	var slots []d.Slot

	for date := truncateToDay(from); date.Before(to); date = date.AddDate(0, 0, 1) {
		weekday := isoWeekday(date.Weekday())
		if !schedule.Days.IsWorkday(d.Workday(1 << (weekday - 1))) {
			continue
		}

		dayStart := time.Date(date.Year(), date.Month(), date.Day(), startHour, startMin, 0, 0, time.UTC)
		dayEnd := time.Date(date.Year(), date.Month(), date.Day(), endHour, endMin, 0, 0, time.UTC)

		for slotStart := dayStart; slotStart.Before(dayEnd); slotStart = slotStart.Add(slotDuration) {
			slotEnd := slotStart.Add(slotDuration)
			if slotEnd.After(dayEnd) {
				break
			}
			slots = append(slots, d.Slot{
				ID:     d.SlotID(uuid.New()),
				RoomID: schedule.RoomID,
				Start:  slotStart,
				End:    slotEnd,
			})
		}
	}

	return slots
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func isoWeekday(w time.Weekday) int {
	if w == time.Sunday {
		return 7
	}
	return int(w)
}
