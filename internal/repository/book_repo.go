package repository

import (
	"context"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
)

type BookingRepo interface {
	Create(ctx context.Context, booking d.Booking) (d.Booking, error)
	FindByID(ctx context.Context, id d.BookingID) (d.Booking, error)
	FindAll(ctx context.Context, page, pageSize int) ([]d.Booking, int, error)
	FindByUserID(ctx context.Context, userID d.UserID) ([]d.Booking, error)
	UpdateStatus(ctx context.Context, id d.BookingID, status d.BookingStatus) (d.Booking, error)
}
