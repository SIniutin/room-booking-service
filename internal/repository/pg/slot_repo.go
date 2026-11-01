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

var _ repository.SlotRepo = (*slotRepoImpl)(nil)

type slotRepoImpl struct {
	db *pgxpool.Pool
}

func NewSlotRepo(db *pgxpool.Pool) repository.SlotRepo {
	return &slotRepoImpl{db: db}
}

func (r *slotRepoImpl) FindByRoomAndDate(ctx context.Context, roomID d.RoomID, date time.Time) ([]d.Slot, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	rows, err := r.db.Query(ctx,
		`SELECT s.id, s.room_id, s.start_at, s.end_at
		 FROM slots s
		 WHERE s.room_id = $1
		   AND s.start_at >= $2
		   AND s.start_at < $3
		   AND NOT EXISTS (
		       SELECT 1 FROM bookings b
		       WHERE b.slot_id = s.id AND b.status = 'active'
		   )
		 ORDER BY s.start_at`,
		uuid.UUID(roomID), dayStart, dayEnd,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []d.Slot
	for rows.Next() {
		var (
			slot d.Slot
			id   uuid.UUID
			rid  uuid.UUID
		)
		if err := rows.Scan(&id, &rid, &slot.Start, &slot.End); err != nil {
			return nil, err
		}
		slot.ID = d.SlotID(id)
		slot.RoomID = d.RoomID(rid)
		slot.Start = slot.Start.UTC()
		slot.End = slot.End.UTC()
		slots = append(slots, slot)
	}

	return slots, rows.Err()
}

func (r *slotRepoImpl) FindByID(ctx context.Context, id d.SlotID) (d.Slot, error) {
	var (
		slot d.Slot
		sid  uuid.UUID
		rid  uuid.UUID
	)

	err := r.db.QueryRow(ctx,
		`SELECT id, room_id, start_at, end_at FROM slots WHERE id = $1`,
		uuid.UUID(id),
	).Scan(&sid, &rid, &slot.Start, &slot.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Slot{}, d.ErrSlotNotFound
		}
		return d.Slot{}, err
	}

	slot.ID = d.SlotID(sid)
	slot.RoomID = d.RoomID(rid)
	slot.Start = slot.Start.UTC()
	slot.End = slot.End.UTC()
	return slot, nil
}

func (r *slotRepoImpl) FindByIDs(ctx context.Context, ids []d.SlotID) (map[d.SlotID]d.Slot, error) {
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uuids[i] = uuid.UUID(id)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, room_id, start_at, end_at FROM slots WHERE id = ANY($1)`,
		uuids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[d.SlotID]d.Slot, len(ids))
	for rows.Next() {
		var (
			slot d.Slot
			sid  uuid.UUID
			rid  uuid.UUID
		)
		if err := rows.Scan(&sid, &rid, &slot.Start, &slot.End); err != nil {
			return nil, err
		}
		slot.ID = d.SlotID(sid)
		slot.RoomID = d.RoomID(rid)
		slot.Start = slot.Start.UTC()
		slot.End = slot.End.UTC()
		result[slot.ID] = slot
	}

	return result, rows.Err()
}

func (r *slotRepoImpl) BulkCreate(ctx context.Context, slots []d.Slot) error {
	if len(slots) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, s := range slots {
		batch.Queue(
			`INSERT INTO slots (id, room_id, start_at, end_at) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (room_id, start_at) DO NOTHING`,
			uuid.UUID(s.ID), uuid.UUID(s.RoomID), s.Start, s.End,
		)
	}

	return r.db.SendBatch(ctx, batch).Close()
}
