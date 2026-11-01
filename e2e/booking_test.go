//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/SIniutin/room-booking-service/internal/cache"
	"github.com/SIniutin/room-booking-service/internal/dto"
	"github.com/SIniutin/room-booking-service/internal/metrics"
	pgRepo "github.com/SIniutin/room-booking-service/internal/repository/pg"
	transporthttp "github.com/SIniutin/room-booking-service/internal/transport/http"
	"github.com/SIniutin/room-booking-service/internal/usecase"
)

func isoWeekday(w time.Weekday) int {
	if w == time.Sunday {
		return 7
	}
	return int(w)
}

func setupServer(t *testing.T) *httptest.Server {
	t.Helper()
	ctx := context.Background()

	// ── Postgres ──────────────────────────────────────────────────
	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:17-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "testdb",
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = pgC.Terminate(ctx) })

	pgHost, _ := pgC.Host(ctx)
	pgPort, _ := pgC.MappedPort(ctx, "5432")
	pgDSN := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", pgHost, pgPort.Port())

	// ── Redis ─────────────────────────────────────────────────────
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start redis: %v", err)
	}
	t.Cleanup(func() { _ = redisC.Terminate(ctx) })

	redisHost, _ := redisC.Host(ctx)
	redisPort, _ := redisC.MappedPort(ctx, "6379")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort.Port())

	// ── Migrations ────────────────────────────────────────────────
	sqlDB, err := sql.Open("pgx", pgDSN)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer sqlDB.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose set dialect: %v", err)
	}
	if err := goose.Up(sqlDB, "../migrations"); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	// ── pgxpool ───────────────────────────────────────────────────
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)

	// ── Cache ─────────────────────────────────────────────────────
	c, err := cache.NewCache(ctx, redisAddr)
	if err != nil {
		t.Fatalf("cache.NewCache: %v", err)
	}

	// ── JWT keys (in-memory) ──────────────────────────────────────
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}

	// ── Wire up ───────────────────────────────────────────────────
	userRepo := pgRepo.NewUserRepo(pool)
	roomRepo := pgRepo.NewRoomRepo(pool)
	scheduleRepo := pgRepo.NewScheduleRepo(pool)
	slotRepo := pgRepo.NewSlotRepo(pool)
	bookingRepo := pgRepo.NewBookingRepo(pool)

	authUsecase := usecase.NewAuthUsecase(privateKey, userRepo, 24*time.Hour)
	slotGenerator := usecase.NewSlotGenerator(scheduleRepo, slotRepo)
	roomUsecase := usecase.NewRoomUsecase(roomRepo, scheduleRepo, slotRepo, slotGenerator)
	bookingUsecase := usecase.NewBookingUsecase(bookingRepo, slotRepo, c)

	registry := prometheus.NewRegistry()
	h := transporthttp.NewHandler(
		authUsecase, roomUsecase, bookingUsecase,
		c, zap.NewNop(), metrics.NewWithRegisterer(registry), 5*time.Minute,
	)
	router := transporthttp.NewRouter(h, &privateKey.PublicKey, []string{"http://localhost:3000"}, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

// doJSON executes an HTTP request with optional JSON body and Bearer token.
func doJSON(t *testing.T, srv *httptest.Server, method, path string, body any, token string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req, err := http.NewRequest(method, srv.URL+path, &buf)
	if err != nil {
		t.Fatalf("http.NewRequest %s %s: %v", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

// mustDecode decodes JSON response body into T, failing the test on error.
func mustDecode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return v
}

// registerAndLogin creates a user via /register then returns the JWT from /login.
func registerAndLogin(t *testing.T, srv *httptest.Server, email, password, role string) string {
	t.Helper()

	reg := doJSON(t, srv, "POST", "/register", map[string]string{
		"email": email, "password": password, "role": role,
	}, "")
	if reg.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: want 201, got %d", email, reg.StatusCode)
	}
	reg.Body.Close()

	login := doJSON(t, srv, "POST", "/login", map[string]string{
		"email": email, "password": password,
	}, "")
	if login.StatusCode != http.StatusOK {
		t.Fatalf("login %s: want 200, got %d", email, login.StatusCode)
	}
	return mustDecode[dto.LoginResponse](t, login).Token
}

// TestInfo checks that GET /_info always returns 200 OK.
func TestInfo(t *testing.T) {
	srv := setupServer(t)

	resp, err := http.Get(srv.URL + "/_info")
	if err != nil {
		t.Fatalf("GET /_info: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /_info: want 200, got %d", resp.StatusCode)
	}
}

// TestBookingFlow is the main E2E scenario:
//
//	register → create room → create schedule → list slots →
//	create booking → double-booking (409) → verify slot hidden →
//	list my bookings → cancel → idempotent cancel → slot visible again
func TestBookingFlow(t *testing.T) {
	srv := setupServer(t)

	adminToken := registerAndLogin(t, srv, "admin@test.com", "adminpass123", "admin")
	userToken := registerAndLogin(t, srv, "user@test.com", "userpass123", "user")

	// 1. Create room (admin)
	roomResp := doJSON(t, srv, "POST", "/rooms/create",
		map[string]string{"name": "Meeting Room A"}, adminToken)
	if roomResp.StatusCode != http.StatusCreated {
		t.Fatalf("create room: want 201, got %d", roomResp.StatusCode)
	}
	roomID := mustDecode[dto.CreateRoomResponse](t, roomResp).Room.ID

	// 2. Create schedule for tomorrow (admin)
	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	tomorrowStr := tomorrow.Format("2006-01-02")

	schedResp := doJSON(t, srv, "POST", "/rooms/"+roomID+"/schedule/create", map[string]any{
		"daysOfWeek": []int{isoWeekday(tomorrow.Weekday())},
		"startTime":  "09:00",
		"endTime":    "18:00",
	}, adminToken)
	if schedResp.StatusCode != http.StatusCreated {
		t.Fatalf("create schedule: want 201, got %d", schedResp.StatusCode)
	}
	schedResp.Body.Close()

	// 3. List slots — 09:00–18:00 = 18 slots of 30 min (user)
	slotsResp := doJSON(t, srv, "GET", "/rooms/"+roomID+"/slots/list?date="+tomorrowStr, nil, userToken)
	if slotsResp.StatusCode != http.StatusOK {
		t.Fatalf("list slots: want 200, got %d", slotsResp.StatusCode)
	}
	slotsBody := mustDecode[dto.ListSlotsResponse](t, slotsResp)
	if len(slotsBody.Slots) != 18 {
		t.Fatalf("list slots: want 18, got %d", len(slotsBody.Slots))
	}
	slotID := slotsBody.Slots[0].ID

	// 4. Book the first slot (user)
	bookResp := doJSON(t, srv, "POST", "/bookings/create", map[string]any{
		"slotId": slotID, "createConferenceLink": false,
	}, userToken)
	if bookResp.StatusCode != http.StatusCreated {
		t.Fatalf("create booking: want 201, got %d", bookResp.StatusCode)
	}
	booking := mustDecode[dto.CreateBookingResponse](t, bookResp)
	if booking.Booking.Status != "active" {
		t.Errorf("booking status: want active, got %s", booking.Booking.Status)
	}
	bookingID := booking.Booking.ID

	// 5. Double booking → 409 Conflict
	dupResp := doJSON(t, srv, "POST", "/bookings/create", map[string]any{
		"slotId": slotID, "createConferenceLink": false,
	}, userToken)
	if dupResp.StatusCode != http.StatusConflict {
		t.Errorf("double booking: want 409, got %d", dupResp.StatusCode)
	}
	dupResp.Body.Close()

	// 6. Booked slot must NOT appear in listing
	slots2Resp := doJSON(t, srv, "GET", "/rooms/"+roomID+"/slots/list?date="+tomorrowStr, nil, userToken)
	if slots2Resp.StatusCode != http.StatusOK {
		t.Fatalf("list slots (post-booking): want 200, got %d", slots2Resp.StatusCode)
	}
	for _, s := range mustDecode[dto.ListSlotsResponse](t, slots2Resp).Slots {
		if s.ID == slotID {
			t.Errorf("booked slot %s must not appear in available list", slotID)
		}
	}

	// 7. /bookings/my must include active booking
	myResp := doJSON(t, srv, "GET", "/bookings/my", nil, userToken)
	if myResp.StatusCode != http.StatusOK {
		t.Fatalf("list my bookings: want 200, got %d", myResp.StatusCode)
	}
	foundActive := false
	for _, b := range mustDecode[dto.ListMyBookingsResponse](t, myResp).Bookings {
		if b.ID == bookingID {
			foundActive = b.Status == "active"
			break
		}
	}
	if !foundActive {
		t.Errorf("active booking %s not found in /bookings/my", bookingID)
	}

	// 8. Cancel booking
	cancelResp := doJSON(t, srv, "POST", "/bookings/"+bookingID+"/cancel", nil, userToken)
	if cancelResp.StatusCode != http.StatusOK {
		t.Fatalf("cancel booking: want 200, got %d", cancelResp.StatusCode)
	}
	cancelBody := mustDecode[dto.CancelBookingResponse](t, cancelResp)
	if cancelBody.Booking.Status != "cancelled" {
		t.Errorf("cancel: want cancelled, got %s", cancelBody.Booking.Status)
	}

	// 9. Idempotent cancel → still 200
	idemp := doJSON(t, srv, "POST", "/bookings/"+bookingID+"/cancel", nil, userToken)
	if idemp.StatusCode != http.StatusOK {
		t.Errorf("idempotent cancel: want 200, got %d", idemp.StatusCode)
	}
	idemp.Body.Close()

	// 10. Slot must appear again after cancellation
	slots3Resp := doJSON(t, srv, "GET", "/rooms/"+roomID+"/slots/list?date="+tomorrowStr, nil, userToken)
	if slots3Resp.StatusCode != http.StatusOK {
		t.Fatalf("list slots (post-cancel): want 200, got %d", slots3Resp.StatusCode)
	}
	slotBack := false
	for _, s := range mustDecode[dto.ListSlotsResponse](t, slots3Resp).Slots {
		if s.ID == slotID {
			slotBack = true
			break
		}
	}
	if !slotBack {
		t.Errorf("slot %s must be available again after cancellation", slotID)
	}
}
