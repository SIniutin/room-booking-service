package http

import (
	"time"

	"go.uber.org/zap"

	"github.com/SIniutin/room-booking-service/internal/cache"
	"github.com/SIniutin/room-booking-service/internal/metrics"
	"github.com/SIniutin/room-booking-service/internal/usecase"
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
