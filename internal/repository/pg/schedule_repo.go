package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	d "github.com/SIniutin/room-booking-service/internal/domain"
	"github.com/SIniutin/room-booking-service/internal/repository"
)

var _ repository.ScheduleRepo = (*scheduleRepoImpl)(nil)

type scheduleRepoImpl struct {
	db *pgxpool.Pool
}

func NewScheduleRepo(db *pgxpool.Pool) repository.ScheduleRepo {
	return &scheduleRepoImpl{db: db}
}

func (r *scheduleRepoImpl) Create(ctx context.Context, s d.Schedule) (d.Schedule, error) {
	startTime := time.Time(s.Start).Format("15:04")
	endTime := time.Time(s.End).Format("15:04")

	_, err := r.db.Exec(ctx,
		`INSERT INTO schedules (id, room_id, days, start_time, end_time) VALUES ($1, $2, $3, $4, $5)`,
		uuid.UUID(s.ID), uuid.UUID(s.RoomID), int(s.Days), startTime, endTime,
	)
	if err != nil {
		return d.Schedule{}, err
	}

	return r.FindByRoomID(ctx, s.RoomID)
}

func (r *scheduleRepoImpl) FindAll(ctx context.Context) ([]d.Schedule, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, room_id, days, start_time, end_time FROM schedules`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []d.Schedule
	for rows.Next() {
		var (
			id        uuid.UUID
			rid       uuid.UUID
			days      int
			startTime time.Time
			endTime   time.Time
		)
		if err := rows.Scan(&id, &rid, &days, &startTime, &endTime); err != nil {
			return nil, err
		}
		start, err := d.NewWorkTime(startTime.Format("15:04"))
		if err != nil {
			return nil, err
		}
		end, err := d.NewWorkTime(endTime.Format("15:04"))
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, d.Schedule{
			ID:     d.ScheduleID(id),
			RoomID: d.RoomID(rid),
			Days:   d.Workday(days),
			Start:  *start,
			End:    *end,
		})
	}
	return schedules, rows.Err()
}

func (r *scheduleRepoImpl) FindByRoomID(ctx context.Context, roomID d.RoomID) (d.Schedule, error) {
	var (
		id        uuid.UUID
		rid       uuid.UUID
		days      int
		startTime time.Time
		endTime   time.Time
	)

	err := r.db.QueryRow(ctx,
		`SELECT id, room_id, days, start_time, end_time FROM schedules WHERE room_id = $1`,
		uuid.UUID(roomID),
	).Scan(&id, &rid, &days, &startTime, &endTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Schedule{}, d.ErrScheduleNotFound
		}
		return d.Schedule{}, err
	}

	start, err := d.NewWorkTime(startTime.Format("15:04"))
	if err != nil {
		return d.Schedule{}, err
	}

	end, err := d.NewWorkTime(endTime.Format("15:04"))
	if err != nil {
		return d.Schedule{}, err
	}

	return d.Schedule{
		ID:     d.ScheduleID(id),
		RoomID: d.RoomID(rid),
		Days:   d.Workday(days),
		Start:  *start,
		End:    *end,
	}, nil
}
