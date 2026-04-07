package domain

import "errors"

var (
	// user
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrEmptyPassword      = errors.New("password is required")
	ErrEmptyRole          = errors.New("role is required")
	ErrRoleIsUndefined    = errors.New("role is undefined")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrCtxCancelled       = errors.New("context cancelled")

	// room
	ErrRoomNotFound          = errors.New("room not found")
	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrScheduleAlreadyExists = errors.New("schedule for this room already exists and cannot be changed")
	ErrInvalidScheduleTime   = errors.New("start time must be before end time")

	// booking
	ErrSlotNotFound      = errors.New("slot not found")
	ErrSlotAlreadyBooked = errors.New("slot is already booked")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrBookingForbidden  = errors.New("cannot cancel another user's booking")
	ErrSlotInThePast     = errors.New("slot is in the past")
)
