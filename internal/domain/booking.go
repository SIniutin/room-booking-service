package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingID uuid.UUID

type BookingStatus string

const (
	ActiveBooking    BookingStatus = "active"
	CancelledBooking BookingStatus = "cancelled"
)

type Booking struct {
	ID        BookingID
	SlotID    SlotID
	UserID    UserID
	Status    BookingStatus
	Link      string // must be URI
	CreatedAt time.Time
}
