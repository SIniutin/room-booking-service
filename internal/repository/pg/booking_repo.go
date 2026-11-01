package pg

import (
	"context"
	"errors"

	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	d "github.com/SIniutin/room-booking-service/internal/domain"
	"github.com/SIniutin/room-booking-service/internal/repository"
)

var _ repository.BookingRepo = (*bookingRepo)(nil)

type bookingRepo struct {
	db *pgxpool.Pool
}

func NewBookingRepo(db *pgxpool.Pool) repository.BookingRepo {
	return &bookingRepo{db: db}
}

func (r *bookingRepo) Create(ctx context.Context, booking d.Booking) (d.Booking, error) {
	_, err := r.db.Exec(ctx,
		`INSERT INTO bookings (id, slot_id, user_id, status, conference_link)
		 VALUES ($1, $2, $3, $4, $5)`,
		uuid.UUID(booking.ID),
		uuid.UUID(booking.SlotID),
		uuid.UUID(booking.UserID),
		string(booking.Status),
		nullableString(booking.Link),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return d.Booking{}, d.ErrSlotAlreadyBooked
		}
		return d.Booking{}, err
	}

	return r.FindByID(ctx, booking.ID)
}

func (r *bookingRepo) FindByID(ctx context.Context, id d.BookingID) (d.Booking, error) {
	var (
		bid       uuid.UUID
		slotID    uuid.UUID
		userID    uuid.UUID
		status    string
		link      *string
		createdAt time.Time
	)

	err := r.db.QueryRow(ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at
		 FROM bookings WHERE id = $1`,
		uuid.UUID(id),
	).Scan(&bid, &slotID, &userID, &status, &link, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Booking{}, d.ErrBookingNotFound
		}
		return d.Booking{}, err
	}

	return scanBooking(bid, slotID, userID, status, link, createdAt), nil
}

func (r *bookingRepo) FindAll(ctx context.Context, page, pageSize int) ([]d.Booking, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at
		 FROM bookings ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	bookings, err := scanBookings(rows)
	if err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

func (r *bookingRepo) FindByUserID(ctx context.Context, userID d.UserID) ([]d.Booking, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at
		 FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`,
		uuid.UUID(userID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookings(rows)
}

func (r *bookingRepo) UpdateStatus(ctx context.Context, id d.BookingID, status d.BookingStatus) (d.Booking, error) {
	_, err := r.db.Exec(ctx,
		`UPDATE bookings SET status = $1 WHERE id = $2`,
		string(status), uuid.UUID(id),
	)
	if err != nil {
		return d.Booking{}, err
	}

	return r.FindByID(ctx, id)
}

// helpers

func scanBookings(rows pgx.Rows) ([]d.Booking, error) {
	var bookings []d.Booking
	for rows.Next() {
		var (
			bid       uuid.UUID
			slotID    uuid.UUID
			userID    uuid.UUID
			status    string
			link      *string
			createdAt time.Time
		)
		if err := rows.Scan(&bid, &slotID, &userID, &status, &link, &createdAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, scanBooking(bid, slotID, userID, status, link, createdAt))
	}
	return bookings, rows.Err()
}

func scanBooking(id, slotID, userID uuid.UUID, status string, link *string, createdAt time.Time) d.Booking {
	b := d.Booking{
		ID:        d.BookingID(id),
		SlotID:    d.SlotID(slotID),
		UserID:    d.UserID(userID),
		Status:    d.BookingStatus(status),
		CreatedAt: createdAt,
	}
	if link != nil {
		b.Link = *link
	}
	return b
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
