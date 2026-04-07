package http

import (
	"time"

	"go.uber.org/zap"

	"github.com/internships-backend/test-backend-SIniutin/internal/cache"
	"github.com/internships-backend/test-backend-SIniutin/internal/metrics"
	"github.com/internships-backend/test-backend-SIniutin/internal/usecase"
)

type Handler struct {
	auth     usecase.AuthUsecase
	room     usecase.RoomUsecase
	booking  usecase.BookingUsecase
	cache    cache.Cache
	cacheTTL time.Duration
	log      *zap.Logger
	metrics  *metrics.Metrics
}

func NewHandler(
	auth usecase.AuthUsecase,
	room usecase.RoomUsecase,
	booking usecase.BookingUsecase,
	c cache.Cache,
	log *zap.Logger,
	m *metrics.Metrics,
	cacheTTL time.Duration,
) *Handler {
	return &Handler{
		auth:     auth,
		room:     room,
		booking:  booking,
		cache:    c,
		log:      log,
		metrics:  m,
		cacheTTL: cacheTTL,
	}
}
