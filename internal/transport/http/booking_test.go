package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	d "github.com/SIniutin/room-booking-service/internal/domain"
	"github.com/SIniutin/room-booking-service/internal/dto"
	"github.com/SIniutin/room-booking-service/internal/metrics"
	m "github.com/SIniutin/room-booking-service/internal/transport/http/middleware"
	"github.com/SIniutin/room-booking-service/internal/usecase"
)

type mockBookingUsecase struct {
	createBooking func(ctx context.Context, userID d.UserID, slotID d.SlotID, createConferenceLink bool) (d.Booking, error)
	listBookings  func(ctx context.Context, page, pageSize int) ([]d.Booking, int, error)
	listMy        func(ctx context.Context, userID d.UserID) ([]d.Booking, error)
	cancelBooking func(ctx context.Context, userID d.UserID, bookingID d.BookingID) (d.Booking, error)
}

func (m mockBookingUsecase) CreateBooking(ctx context.Context, userID d.UserID, slotID d.SlotID, createConferenceLink bool) (d.Booking, error) {
	return m.createBooking(ctx, userID, slotID, createConferenceLink)
}
func (m mockBookingUsecase) ListBookings(ctx context.Context, page, pageSize int) ([]d.Booking, int, error) {
	return m.listBookings(ctx, page, pageSize)
}
func (m mockBookingUsecase) ListMyBookings(ctx context.Context, userID d.UserID) ([]d.Booking, error) {
	return m.listMy(ctx, userID)
}
func (m mockBookingUsecase) CancelBooking(ctx context.Context, userID d.UserID, bookingID d.BookingID) (d.Booking, error) {
	return m.cancelBooking(ctx, userID, bookingID)
}

var _ usecase.BookingUsecase = mockBookingUsecase{}

func newBookingHandler(booking usecase.BookingUsecase) *Handler {
	return NewHandler(
		stubAuthUsecase{},
		mockRoomUsecase{
			createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
			listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
			createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
				return d.Schedule{}, nil
			},
			listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
		},
		booking,
		stubCache{},
		zap.NewNop(),
		metrics.NewWithRegisterer(prometheus.NewRegistry()),
		time.Minute,
	)
}

func withClaims(r *http.Request, userID string, role d.UserRole) *http.Request {
	ctx := context.WithValue(r.Context(), m.ClaimsKey, m.Claims{UserID: userID, Role: role})
	return r.WithContext(ctx)
}

func TestCreateBooking_ReturnsConflict(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()
	h := newBookingHandler(mockBookingUsecase{
		createBooking: func(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) {
			return d.Booking{}, d.ErrSlotAlreadyBooked
		},
		listBookings:  func(context.Context, int, int) ([]d.Booking, int, error) { return nil, 0, nil },
		listMy:        func(context.Context, d.UserID) ([]d.Booking, error) { return nil, nil },
		cancelBooking: func(context.Context, d.UserID, d.BookingID) (d.Booking, error) { return d.Booking{}, nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBufferString(`{"slotId":"`+slotID.String()+`"}`))
	req = withClaims(req, userID.String(), d.RoleUser)
	rec := httptest.NewRecorder()

	h.createBooking(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	var got dto.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "SLOT_ALREADY_BOOKED", got.Error.Code)
}

func TestCreateBooking_ReturnsBadRequestForPastSlot(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()
	h := newBookingHandler(mockBookingUsecase{
		createBooking: func(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) {
			return d.Booking{}, d.ErrSlotInThePast
		},
		listBookings:  func(context.Context, int, int) ([]d.Booking, int, error) { return nil, 0, nil },
		listMy:        func(context.Context, d.UserID) ([]d.Booking, error) { return nil, nil },
		cancelBooking: func(context.Context, d.UserID, d.BookingID) (d.Booking, error) { return d.Booking{}, nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBufferString(`{"slotId":"`+slotID.String()+`"}`))
	req = withClaims(req, userID.String(), d.RoleUser)
	rec := httptest.NewRecorder()

	h.createBooking(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCancelBooking_ReturnsForbidden(t *testing.T) {
	bookingID := uuid.New()
	userID := uuid.New()
	h := newBookingHandler(mockBookingUsecase{
		createBooking: func(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) { return d.Booking{}, nil },
		listBookings:  func(context.Context, int, int) ([]d.Booking, int, error) { return nil, 0, nil },
		listMy:        func(context.Context, d.UserID) ([]d.Booking, error) { return nil, nil },
		cancelBooking: func(context.Context, d.UserID, d.BookingID) (d.Booking, error) {
			return d.Booking{}, d.ErrBookingForbidden
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/cancel", nil)
	req = withClaims(req, userID.String(), d.RoleUser)
	req = withURLParam(req, "bookingId", bookingID.String())
	rec := httptest.NewRecorder()

	h.cancelBooking(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	var got dto.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "FORBIDDEN", got.Error.Code)
}

func TestListBookings_DefaultPagination(t *testing.T) {
	var gotPage, gotPageSize int
	h := newBookingHandler(mockBookingUsecase{
		createBooking: func(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) { return d.Booking{}, nil },
		listBookings: func(_ context.Context, page, pageSize int) ([]d.Booking, int, error) {
			gotPage, gotPageSize = page, pageSize
			return []d.Booking{}, 0, nil
		},
		listMy:        func(context.Context, d.UserID) ([]d.Booking, error) { return nil, nil },
		cancelBooking: func(context.Context, d.UserID, d.BookingID) (d.Booking, error) { return d.Booking{}, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/bookings/list?page=0&pageSize=500", nil)
	rec := httptest.NewRecorder()

	h.listBookings(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, gotPage)
	assert.Equal(t, 100, gotPageSize)
}

func TestCreateBooking_UnauthorizedWithoutClaims(t *testing.T) {
	h := newBookingHandler(mockBookingUsecase{
		createBooking: func(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) { return d.Booking{}, nil },
		listBookings:  func(context.Context, int, int) ([]d.Booking, int, error) { return nil, 0, nil },
		listMy:        func(context.Context, d.UserID) ([]d.Booking, error) { return nil, nil },
		cancelBooking: func(context.Context, d.UserID, d.BookingID) (d.Booking, error) { return d.Booking{}, nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBufferString(`{"slotId":"`+uuid.NewString()+`"}`))
	rec := httptest.NewRecorder()

	h.createBooking(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateBooking_BadRequestOnInvalidSlotID(t *testing.T) {
	userID := uuid.New()
	h := newBookingHandler(mockBookingUsecase{
		createBooking: func(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) { return d.Booking{}, nil },
		listBookings:  func(context.Context, int, int) ([]d.Booking, int, error) { return nil, 0, nil },
		listMy:        func(context.Context, d.UserID) ([]d.Booking, error) { return nil, nil },
		cancelBooking: func(context.Context, d.UserID, d.BookingID) (d.Booking, error) { return d.Booking{}, nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBufferString(`{"slotId":"bad-id"}`))
	req = withClaims(req, userID.String(), d.RoleUser)
	rec := httptest.NewRecorder()

	h.createBooking(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListMyBookings_Success(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	slotID := uuid.New()
	h := newBookingHandler(mockBookingUsecase{
		createBooking: func(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) { return d.Booking{}, nil },
		listBookings:  func(context.Context, int, int) ([]d.Booking, int, error) { return nil, 0, nil },
		listMy: func(context.Context, d.UserID) ([]d.Booking, error) {
			return []d.Booking{{ID: d.BookingID(bookingID), SlotID: d.SlotID(slotID), UserID: d.UserID(userID), Status: d.ActiveBooking}}, nil
		},
		cancelBooking: func(context.Context, d.UserID, d.BookingID) (d.Booking, error) { return d.Booking{}, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
	req = withClaims(req, userID.String(), d.RoleUser)
	rec := httptest.NewRecorder()

	h.listMyBookings(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var got dto.ListMyBookingsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Len(t, got.Bookings, 1)
	assert.Equal(t, bookingID.String(), got.Bookings[0].ID)
}

func TestCancelBooking_NotFound(t *testing.T) {
	bookingID := uuid.New()
	userID := uuid.New()
	h := newBookingHandler(mockBookingUsecase{
		createBooking: func(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) { return d.Booking{}, nil },
		listBookings:  func(context.Context, int, int) ([]d.Booking, int, error) { return nil, 0, nil },
		listMy:        func(context.Context, d.UserID) ([]d.Booking, error) { return nil, nil },
		cancelBooking: func(context.Context, d.UserID, d.BookingID) (d.Booking, error) {
			return d.Booking{}, d.ErrBookingNotFound
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/cancel", nil)
	req = withClaims(req, userID.String(), d.RoleUser)
	req = withURLParam(req, "bookingId", bookingID.String())
	rec := httptest.NewRecorder()

	h.cancelBooking(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListMyBookings_UnauthorizedWithoutClaims(t *testing.T) {
	h := newBookingHandler(mockBookingUsecase{
		createBooking: func(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) { return d.Booking{}, nil },
		listBookings:  func(context.Context, int, int) ([]d.Booking, int, error) { return nil, 0, nil },
		listMy:        func(context.Context, d.UserID) ([]d.Booking, error) { return nil, nil },
		cancelBooking: func(context.Context, d.UserID, d.BookingID) (d.Booking, error) { return d.Booking{}, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
	rec := httptest.NewRecorder()

	h.listMyBookings(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
