package http

import (
	"crypto/rsa"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	m "github.com/internships-backend/test-backend-SIniutin/internal/transport/http/middleware"
)

func NewRouter(h *Handler, publicKey *rsa.PublicKey, allowedOrigins []string, metricsHandler http.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(m.CORSMiddleware(allowedOrigins))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(m.MetricsMiddleware(h.metrics))

	r.Get("/_info", h.info)
	r.Handle("/metrics", metricsHandler)
	r.Post("/dummyLogin", h.dummyLogin)
	r.Post("/register", h.register)
	r.Post("/login", h.login)

	auth := m.AuthMiddleware(publicKey)

	r.Route("/rooms", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth)
			r.Use(m.CacheMiddleware(h.cache, h.cacheTTL))
			r.Get("/list", h.listRooms)
			r.Get("/{roomId}/slots/list", h.listSlots)
		})

		// only admin
		r.Group(func(r chi.Router) {
			r.Use(auth)
			r.Use(m.AdminOnlyMiddleware)
			r.Post("/create", h.createRoom)
			r.Post("/{roomId}/schedule/create", h.createSchedule)
		})
	})

	r.Route("/bookings", func(r chi.Router) {
		// only user
		r.Group(func(r chi.Router) {
			r.Use(auth)
			r.Use(m.UserOnlyMiddleware)
			r.Post("/create", h.createBooking)
			r.Get("/my", h.listMyBookings)
			r.Post("/{bookingId}/cancel", h.cancelBooking)
		})

		// only admin
		r.Group(func(r chi.Router) {
			r.Use(auth)
			r.Use(m.AdminOnlyMiddleware)
			r.Get("/list", h.listBookings)
		})
	})

	return r
}
