package domain

import "github.com/google/uuid"

type ScheduleID uuid.UUID

type Schedule struct {
	ID     ScheduleID
	RoomID RoomID
	Days   Workday
	Start  WorkTime
	End    WorkTime
}
