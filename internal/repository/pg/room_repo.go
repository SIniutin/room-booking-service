package pg

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	"github.com/internships-backend/test-backend-SIniutin/internal/repository"
)

var _ repository.RoomRepo = (*roomRepo)(nil)

type roomRepo struct {
	db *pgxpool.Pool
}

func NewRoomRepo(db *pgxpool.Pool) repository.RoomRepo {
	return &roomRepo{db: db}
}

func (r *roomRepo) Create(ctx context.Context, room d.Room) (d.Room, error) {
	_, err := r.db.Exec(ctx,
		`INSERT INTO rooms (id, name, description, capacity) VALUES ($1, $2, $3, $4)`,
		uuid.UUID(room.ID), room.Name, room.Description, room.Capacity,
	)
	if err != nil {
		return d.Room{}, err
	}

	return r.FindByID(ctx, room.ID)
}

func (r *roomRepo) FindAll(ctx context.Context) ([]d.Room, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []d.Room
	for rows.Next() {
		var room d.Room
		var id uuid.UUID
		if err := rows.Scan(&id, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			return nil, err
		}
		room.ID = d.RoomID(id)
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}

func (r *roomRepo) FindByID(ctx context.Context, id d.RoomID) (d.Room, error) {
	var room d.Room
	var rid uuid.UUID

	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`,
		uuid.UUID(id),
	).Scan(&rid, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Room{}, d.ErrRoomNotFound
		}
		return d.Room{}, err
	}

	room.ID = d.RoomID(rid)
	return room, nil
}
