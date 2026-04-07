package domain

import (
	"time"

	"github.com/google/uuid"
)

type SlotID uuid.UUID

type Slot struct {
	ID     SlotID
	RoomID RoomID
	Start  time.Time
	End    time.Time
}
