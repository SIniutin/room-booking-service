package usecase

import (
	"context"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
)

type BookingUsecase interface {
	CreateBooking(ctx context.Context, userID d.UserID, slotID d.SlotID, createConferenceLink bool) (d.Booking, error)
	ListBookings(ctx context.Context, page, pageSize int) ([]d.Booking, int, error)
	ListMyBookings(ctx context.Context, userID d.UserID) ([]d.Booking, error)
	CancelBooking(ctx context.Context, userID d.UserID, bookingID d.BookingID) (d.Booking, error)
}
