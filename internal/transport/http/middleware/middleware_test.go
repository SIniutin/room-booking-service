package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SIniutin/room-booking-service/internal/domain"
	"github.com/SIniutin/room-booking-service/internal/metrics"
)

type testCache struct {
	get func(ctx context.Context, key string) (string, error)
	set func(ctx context.Context, key string, value any, ttl time.Duration) error
	del func(ctx context.Context, key string) error
}

func (c testCache) Get(ctx context.Context, key string) (string, error) { return c.get(ctx, key) }
func (c testCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.set(ctx, key, value, ttl)
}
func (c testCache) Del(ctx context.Context, key string) error { return c.del(ctx, key) }
func (c testCache) Close() error                              { return nil }

func newMetricsForTest() *metrics.Metrics {
	reg := prometheus.NewRegistry()
	m := &metrics.Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "http_requests_total_test", Help: "test"},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "http_request_duration_seconds_test", Help: "test"},
			[]string{"method", "path"},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "app_errors_total_test", Help: "test"},
			[]string{"layer", "op"},
		),
	}
	reg.MustRegister(m.RequestsTotal, m.RequestDuration, m.ErrorsTotal)
	return m
}

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	nextCalled := false
	handler := CORSMiddleware([]string{"http://localhost:3000"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, nextCalled)
	assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	nextCalled := false
	handler := CORSMiddleware([]string{"http://localhost:3000"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/rooms/list", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestCacheMiddleware_Hit(t *testing.T) {
	handler := CacheMiddleware(testCache{
		get: func(context.Context, string) (string, error) { return `{"ok":true}`, nil },
		set: func(context.Context, string, any, time.Duration) error { return nil },
		del: func(context.Context, string) error { return nil },
	}, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called on cache hit")
	}))

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "HIT", rec.Header().Get("X-Cache"))
	assert.JSONEq(t, `{"ok":true}`, rec.Body.String())
}

func TestCacheMiddleware_MissStoresSuccessfulResponse(t *testing.T) {
	setCalled := false
	handler := CacheMiddleware(testCache{
		get: func(context.Context, string) (string, error) { return "", assert.AnError },
		set: func(_ context.Context, key string, value any, ttl time.Duration) error {
			setCalled = true
			assert.Equal(t, "route:/rooms/list", key)
			assert.Equal(t, `{"rooms":[]}`, value)
			assert.Equal(t, time.Minute, ttl)
			return nil
		},
		del: func(context.Context, string) error { return nil },
	}, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"rooms":[]}`))
		require.NoError(t, err)
	}))

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, setCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMetricsMiddleware_RecordsRequest(t *testing.T) {
	m := newMetricsForTest()
	handler := MetricsMiddleware(m)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	metric := &dto.Metric{}
	require.NoError(t, m.RequestsTotal.WithLabelValues(http.MethodPost, "/login", "201").Write(metric))
	assert.Equal(t, 1.0, metric.GetCounter().GetValue())
}

func TestAuthMiddleware_SetsClaims(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":  "00000000-0000-0000-0000-000000000002",
		"role": "user",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString(privateKey)
	require.NoError(t, err)

	var got Claims
	handler := AuthMiddleware(&privateKey.PublicKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ClaimsKey).(Claims)
		require.True(t, ok)
		got = claims
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", got.UserID)
	assert.Equal(t, domain.RoleUser, got.Role)
}

func TestAdminOnlyMiddleware_Forbidden(t *testing.T) {
	handler := AdminOnlyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/rooms/create", nil)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsKey, Claims{UserID: "1", Role: domain.RoleUser}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestUserOnlyMiddleware_Forbidden(t *testing.T) {
	handler := UserOnlyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsKey, Claims{UserID: "1", Role: domain.RoleAdmin}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}
