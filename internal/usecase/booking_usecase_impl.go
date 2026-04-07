package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-SIniutin/internal/cache"
	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	r "github.com/internships-backend/test-backend-SIniutin/internal/repository"
)

type bookingUsecaseImpl struct {
	bookingRepo r.BookingRepo
	slotRepo    r.SlotRepo
	cache       cache.Cache
}

func NewBookingUsecase(bookingRepo r.BookingRepo, slotRepo r.SlotRepo, c cache.Cache) BookingUsecase {
	return &bookingUsecaseImpl{
		bookingRepo: bookingRepo,
		slotRepo:    slotRepo,
		cache:       c,
	}
}

func (uc *bookingUsecaseImpl) invalidateSlotCache(ctx context.Context, roomID d.RoomID, slotStart time.Time) {
	key := fmt.Sprintf("route:/rooms/%s/slots/list?date=%s",
		uuid.UUID(roomID).String(),
		slotStart.UTC().Format("2006-01-02"),
	)
	_ = uc.cache.Del(ctx, key)
}

func (uc *bookingUsecaseImpl) CreateBooking(ctx context.Context, userID d.UserID, slotID d.SlotID, createConferenceLink bool) (d.Booking, error) {
	slot, err := uc.slotRepo.FindByID(ctx, slotID)
	if err != nil {
		return d.Booking{}, d.ErrSlotNotFound
	}

	if slot.Start.Before(time.Now()) {
		return d.Booking{}, d.ErrSlotInThePast
	}

	booking := d.Booking{
		ID:     d.BookingID(uuid.New()),
		SlotID: slotID,
		UserID: userID,
		Status: d.ActiveBooking,
	}

	if createConferenceLink {
		// TODO: integrate conference link generation
		link := "https://meet.example.com/" + uuid.New().String()
		booking.Link = link
	}

	created, err := uc.bookingRepo.Create(ctx, booking)
	if err != nil {
		return d.Booking{}, err
	}

	uc.invalidateSlotCache(ctx, slot.RoomID, slot.Start)
	return created, nil
}

func (uc *bookingUsecaseImpl) ListBookings(ctx context.Context, page, pageSize int) ([]d.Booking, int, error) {
	return uc.bookingRepo.FindAll(ctx, page, pageSize)
}

func (uc *bookingUsecaseImpl) ListMyBookings(ctx context.Context, userID d.UserID) ([]d.Booking, error) {
	bookings, err := uc.bookingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(bookings) == 0 {
		return bookings, nil
	}

	slotIDs := make([]d.SlotID, len(bookings))
	for i, b := range bookings {
		slotIDs[i] = b.SlotID
	}

	slots, err := uc.slotRepo.FindByIDs(ctx, slotIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := make([]d.Booking, 0, len(bookings))
	for _, b := range bookings {
		slot, ok := slots[b.SlotID]
		if ok && slot.Start.After(now) {
			result = append(result, b)
		}
	}

	return result, nil
}

func (uc *bookingUsecaseImpl) CancelBooking(ctx context.Context, userID d.UserID, bookingID d.BookingID) (d.Booking, error) {
	booking, err := uc.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return d.Booking{}, err
	}

	if booking.UserID != userID {
		return d.Booking{}, d.ErrBookingForbidden
	}

	// идемпотентность — уже отменена, просто возвращаем
	if booking.Status == d.CancelledBooking {
		return booking, nil
	}

	slot, err := uc.slotRepo.FindByID(ctx, booking.SlotID)
	if err == nil {
		uc.invalidateSlotCache(ctx, slot.RoomID, slot.Start)
	}

	return uc.bookingRepo.UpdateStatus(ctx, bookingID, d.CancelledBooking)
}
