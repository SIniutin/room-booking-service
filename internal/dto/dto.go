package dto

import "time"

// Auth
type DummyLoginRequest struct {
	Role string `json:"role"`
}

type DummyLoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type RegisterResponse struct {
	User UserDTO `json:"user"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Rooms
type CreateRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Capacity    *int    `json:"capacity,omitempty"`
}

type CreateRoomResponse struct {
	Room RoomDTO `json:"room"`
}

type ListRoomsResponse struct {
	Rooms []RoomDTO `json:"rooms"`
}

// Schedules
type CreateScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type CreateScheduleResponse struct {
	Schedule ScheduleDTO `json:"schedule"`
}

// Slots
type ListSlotsResponse struct {
	Slots []SlotDTO `json:"slots"`
}

// Bookings
type CreateBookingRequest struct {
	SlotID               string `json:"slotId"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

type CreateBookingResponse struct {
	Booking BookingDTO `json:"booking"`
}

type ListBookingsResponse struct {
	Bookings   []BookingDTO  `json:"bookings"`
	Pagination PaginationDTO `json:"pagination"`
}

type ListMyBookingsResponse struct {
	Bookings []BookingDTO `json:"bookings"`
}

type CancelBookingResponse struct {
	Booking BookingDTO `json:"booking"`
}

// Shared DTOs
type UserDTO struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}

type RoomDTO struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Capacity    *int       `json:"capacity,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
}

type ScheduleDTO struct {
	ID         string `json:"id"`
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type SlotDTO struct {
	ID     string    `json:"id"`
	RoomID string    `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type BookingDTO struct {
	ID             string     `json:"id"`
	SlotID         string     `json:"slotId"`
	UserID         string     `json:"userId"`
	Status         string     `json:"status"`
	ConferenceLink *string    `json:"conferenceLink,omitempty"`
	CreatedAt      *time.Time `json:"createdAt,omitempty"`
}

type PaginationDTO struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDTO `json:"error"`
}
