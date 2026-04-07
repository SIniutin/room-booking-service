package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoomID uuid.UUID

type Room struct {
	ID          RoomID
	Name        string
	Description *string
	Capacity    *int
	CreatedAt   time.Time
}
