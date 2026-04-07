package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/internships-backend/test-backend-SIniutin/internal/cache"
	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	"github.com/internships-backend/test-backend-SIniutin/internal/dto"
	"github.com/internships-backend/test-backend-SIniutin/internal/metrics"
	"github.com/internships-backend/test-backend-SIniutin/internal/usecase"
)

type stubAuthUsecase struct{}

func (stubAuthUsecase) Register(context.Context, string, string, d.UserRole) (d.User, error) {
	return d.User{}, nil
}
func (stubAuthUsecase) Login(context.Context, string, string) (d.Token, error) {
	return "", nil
}
func (stubAuthUsecase) DummyLogin(context.Context, d.UserRole) (d.Token, error) {
	return "", nil
}

type stubBookingUsecase struct{}

func (stubBookingUsecase) CreateBooking(context.Context, d.UserID, d.SlotID, bool) (d.Booking, error) {
	return d.Booking{}, nil
}
func (stubBookingUsecase) ListBookings(context.Context, int, int) ([]d.Booking, int, error) {
	return nil, 0, nil
}
func (stubBookingUsecase) ListMyBookings(context.Context, d.UserID) ([]d.Booking, error) {
	return nil, nil
}
func (stubBookingUsecase) CancelBooking(context.Context, d.UserID, d.BookingID) (d.Booking, error) {
	return d.Booking{}, nil
}

type stubCache struct{}

func (stubCache) Get(context.Context, string) (string, error) { return "", nil }
func (stubCache) Set(context.Context, string, any, time.Duration) error {
	return nil
}
func (stubCache) Del(context.Context, string) error { return nil }
func (stubCache) Close() error                      { return nil }

type mockRoomUsecase struct {
	createRoom     func(ctx context.Context, name string, description *string, capacity *int) (d.Room, error)
	listRooms      func(ctx context.Context) ([]d.Room, error)
	createSchedule func(ctx context.Context, roomID d.RoomID, days []int, startTime, endTime string) (d.Schedule, error)
	listSlots      func(ctx context.Context, roomID d.RoomID, date time.Time) ([]d.Slot, error)
}

func (m mockRoomUsecase) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (d.Room, error) {
	return m.createRoom(ctx, name, description, capacity)
}
func (m mockRoomUsecase) ListRooms(ctx context.Context) ([]d.Room, error) {
	return m.listRooms(ctx)
}
func (m mockRoomUsecase) CreateSchedule(ctx context.Context, roomID d.RoomID, days []int, startTime, endTime string) (d.Schedule, error) {
	return m.createSchedule(ctx, roomID, days, startTime, endTime)
}
func (m mockRoomUsecase) ListSlots(ctx context.Context, roomID d.RoomID, date time.Time) ([]d.Slot, error) {
	return m.listSlots(ctx, roomID, date)
}

var (
	_ usecase.AuthUsecase    = stubAuthUsecase{}
	_ usecase.BookingUsecase = stubBookingUsecase{}
	_ cache.Cache            = stubCache{}
	_ usecase.RoomUsecase    = mockRoomUsecase{}
)

func newTestHandler(room usecase.RoomUsecase) *Handler {
	return NewHandler(
		stubAuthUsecase{},
		room,
		stubBookingUsecase{},
		stubCache{},
		zap.NewNop(),
		metrics.NewWithRegisterer(prometheus.NewRegistry()),
		time.Minute,
	)
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateSchedule_ReturnsBadRequestForInvalidScheduleTime(t *testing.T) {
	roomID := uuid.New()
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
		listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, d.ErrInvalidScheduleTime
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
	})

	body, err := json.Marshal(dto.CreateScheduleRequest{
		DaysOfWeek: []int{1},
		StartTime:  "18:00",
		EndTime:    "09:00",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "roomId", roomID.String())
	rec := httptest.NewRecorder()

	h.createSchedule(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var got dto.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "INVALID_REQUEST", got.Error.Code)
	assert.Equal(t, d.ErrInvalidScheduleTime.Error(), got.Error.Message)
}

func TestCreateSchedule_ReturnsBadRequestForInvalidWorkTime(t *testing.T) {
	roomID := uuid.New()
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
		listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, d.ErrInvalidWorkTimeFormat
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
	})

	body, err := json.Marshal(dto.CreateScheduleRequest{
		DaysOfWeek: []int{1},
		StartTime:  "ab:cd",
		EndTime:    "18:00",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "roomId", roomID.String())
	rec := httptest.NewRecorder()

	h.createSchedule(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var got dto.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "INVALID_REQUEST", got.Error.Code)
	assert.Equal(t, d.ErrInvalidWorkTimeFormat.Error(), got.Error.Message)
}

func TestListSlots_ReturnsNotFoundForUnknownRoom(t *testing.T) {
	roomID := uuid.New()
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
		listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, nil
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) {
			return nil, d.ErrRoomNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/slots/list?date=2026-04-07", nil)
	req = withURLParam(req, "roomId", roomID.String())
	rec := httptest.NewRecorder()

	h.listSlots(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	var got dto.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "ROOM_NOT_FOUND", got.Error.Code)
	assert.Equal(t, d.ErrRoomNotFound.Error(), got.Error.Message)
}

func TestCreateRoom_Success(t *testing.T) {
	roomID := uuid.New()
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(_ context.Context, name string, description *string, capacity *int) (d.Room, error) {
			return d.Room{ID: d.RoomID(roomID), Name: name, Description: description, Capacity: capacity}, nil
		},
		listRooms: func(context.Context) ([]d.Room, error) { return nil, nil },
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, nil
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/rooms/create", bytes.NewBufferString(`{"name":"Board Room"}`))
	rec := httptest.NewRecorder()

	h.createRoom(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var got dto.CreateRoomResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, roomID.String(), got.Room.ID)
	assert.Equal(t, "Board Room", got.Room.Name)
}

func TestCreateRoom_BadRequestOnMissingName(t *testing.T) {
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
		listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, nil
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/rooms/create", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	h.createRoom(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListRooms_Success(t *testing.T) {
	roomID := uuid.New()
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
		listRooms: func(context.Context) ([]d.Room, error) {
			return []d.Room{{ID: d.RoomID(roomID), Name: "Board Room"}}, nil
		},
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, nil
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	rec := httptest.NewRecorder()

	h.listRooms(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var got dto.ListRoomsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Len(t, got.Rooms, 1)
	assert.Equal(t, "Board Room", got.Rooms[0].Name)
}

func TestListSlots_InvalidDate(t *testing.T) {
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
		listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, nil
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+uuid.NewString()+"/slots/list?date=bad-date", nil)
	req = withURLParam(req, "roomId", uuid.NewString())
	rec := httptest.NewRecorder()

	h.listSlots(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListSlots_MissingDate(t *testing.T) {
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
		listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, nil
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
	})

	roomID := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID+"/slots/list", nil)
	req = withURLParam(req, "roomId", roomID)
	rec := httptest.NewRecorder()

	h.listSlots(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateSchedule_Conflict(t *testing.T) {
	roomID := uuid.New()
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
		listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, d.ErrScheduleAlreadyExists
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
	})

	body, err := json.Marshal(dto.CreateScheduleRequest{DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "18:00"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "roomId", roomID.String())
	rec := httptest.NewRecorder()

	h.createSchedule(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestCreateSchedule_RoomNotFound(t *testing.T) {
	roomID := uuid.New()
	h := newTestHandler(mockRoomUsecase{
		createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
		listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
		createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
			return d.Schedule{}, d.ErrRoomNotFound
		},
		listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
	})

	body, err := json.Marshal(dto.CreateScheduleRequest{DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "18:00"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID.String()+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "roomId", roomID.String())
	rec := httptest.NewRecorder()

	h.createSchedule(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
