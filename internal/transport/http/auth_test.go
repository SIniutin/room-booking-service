package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

type mockAuthUsecase struct {
	register   func(ctx context.Context, email, password string, role d.UserRole) (d.User, error)
	login      func(ctx context.Context, email, password string) (d.Token, error)
	dummyLogin func(ctx context.Context, role d.UserRole) (d.Token, error)
}

func (m mockAuthUsecase) Register(ctx context.Context, email, password string, role d.UserRole) (d.User, error) {
	return m.register(ctx, email, password, role)
}
func (m mockAuthUsecase) Login(ctx context.Context, email, password string) (d.Token, error) {
	return m.login(ctx, email, password)
}
func (m mockAuthUsecase) DummyLogin(ctx context.Context, role d.UserRole) (d.Token, error) {
	return m.dummyLogin(ctx, role)
}

var _ usecase.AuthUsecase = mockAuthUsecase{}

func newAuthTestHandler(auth usecase.AuthUsecase) *Handler {
	return NewHandler(
		auth,
		mockRoomUsecase{
			createRoom: func(context.Context, string, *string, *int) (d.Room, error) { return d.Room{}, nil },
			listRooms:  func(context.Context) ([]d.Room, error) { return nil, nil },
			createSchedule: func(context.Context, d.RoomID, []int, string, string) (d.Schedule, error) {
				return d.Schedule{}, nil
			},
			listSlots: func(context.Context, d.RoomID, time.Time) ([]d.Slot, error) { return nil, nil },
		},
		stubBookingUsecase{},
		stubCache{},
		zap.NewNop(),
		metrics.NewWithRegisterer(prometheus.NewRegistry()),
		time.Minute,
	)
}

func TestDummyLogin_Success(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register: func(context.Context, string, string, d.UserRole) (d.User, error) { return d.User{}, nil },
		login:    func(context.Context, string, string) (d.Token, error) { return "", nil },
		dummyLogin: func(_ context.Context, role d.UserRole) (d.Token, error) {
			assert.Equal(t, d.RoleAdmin, role)
			return d.Token("jwt-token"), nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(`{"role":"admin"}`))
	rec := httptest.NewRecorder()

	h.dummyLogin(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var got dto.DummyLoginResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "jwt-token", got.Token)
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register: func(context.Context, string, string, d.UserRole) (d.User, error) { return d.User{}, nil },
		login:    func(context.Context, string, string) (d.Token, error) { return "", nil },
		dummyLogin: func(context.Context, d.UserRole) (d.Token, error) {
			return "", d.ErrRoleIsUndefined
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(`{"role":"guest"}`))
	rec := httptest.NewRecorder()

	h.dummyLogin(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDummyLogin_InvalidBody(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register:   func(context.Context, string, string, d.UserRole) (d.User, error) { return d.User{}, nil },
		login:      func(context.Context, string, string) (d.Token, error) { return "", nil },
		dummyLogin: func(context.Context, d.UserRole) (d.Token, error) { return "", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	h.dummyLogin(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_Conflict(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register: func(context.Context, string, string, d.UserRole) (d.User, error) {
			return d.User{}, d.ErrUserAlreadyExists
		},
		login:      func(context.Context, string, string) (d.Token, error) { return "", nil },
		dummyLogin: func(context.Context, d.UserRole) (d.Token, error) { return "", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"email":"a@b.com","password":"secret123","role":"user"}`))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	var got dto.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "USER_ALREADY_EXISTS", got.Error.Code)
}

func TestRegister_BadRequestOnMissingFields(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register: func(context.Context, string, string, d.UserRole) (d.User, error) {
			return d.User{}, errors.New("must not be called")
		},
		login:      func(context.Context, string, string) (d.Token, error) { return "", nil },
		dummyLogin: func(context.Context, d.UserRole) (d.Token, error) { return "", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"email":"a@b.com"}`))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogin_Unauthorized(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register: func(context.Context, string, string, d.UserRole) (d.User, error) { return d.User{}, nil },
		login: func(context.Context, string, string) (d.Token, error) {
			return "", d.ErrInvalidCredentials
		},
		dummyLogin: func(context.Context, d.UserRole) (d.Token, error) { return "", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"a@b.com","password":"wrong"}`))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	var got dto.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "UNAUTHORIZED", got.Error.Code)
}

func TestLogin_BadRequestOnDecode(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register:   func(context.Context, string, string, d.UserRole) (d.User, error) { return d.User{}, nil },
		login:      func(context.Context, string, string) (d.Token, error) { return "", errors.New("must not be called") },
		dummyLogin: func(context.Context, d.UserRole) (d.Token, error) { return "", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogin_BadRequestOnMissingFields(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register:   func(context.Context, string, string, d.UserRole) (d.User, error) { return d.User{}, nil },
		login:      func(context.Context, string, string) (d.Token, error) { return "", errors.New("must not be called") },
		dummyLogin: func(context.Context, d.UserRole) (d.Token, error) { return "", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"a@b.com"}`))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_Success(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register: func(_ context.Context, email, password string, role d.UserRole) (d.User, error) {
			parsed, _ := d.NewEmail(email)
			return d.User{ID: d.UserID(testUUID1), Email: *parsed, Role: role}, nil
		},
		login:      func(context.Context, string, string) (d.Token, error) { return "", nil },
		dummyLogin: func(context.Context, d.UserRole) (d.Token, error) { return "", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"email":"a@b.com","password":"secret123","role":"user"}`))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var got dto.RegisterResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "a@b.com", got.User.Email)
	assert.Equal(t, "user", got.User.Role)
}

func TestLogin_Success(t *testing.T) {
	h := newAuthTestHandler(mockAuthUsecase{
		register: func(context.Context, string, string, d.UserRole) (d.User, error) { return d.User{}, nil },
		login: func(context.Context, string, string) (d.Token, error) {
			return d.Token("signed-token"), nil
		},
		dummyLogin: func(context.Context, d.UserRole) (d.Token, error) { return "", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"a@b.com","password":"secret123"}`))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var got dto.LoginResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "signed-token", got.Token)
}

var testUUID1 = mustUUID("11111111-1111-1111-1111-111111111111")

func mustUUID(raw string) uuid.UUID {
	id, err := uuid.Parse(raw)
	if err != nil {
		panic(err)
	}
	return id
}

var _ cache.Cache = stubCache{}
