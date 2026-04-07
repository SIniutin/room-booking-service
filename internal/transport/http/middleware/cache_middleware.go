package middleware

import (
	"bytes"
	"net/http"
	"time"

	"github.com/internships-backend/test-backend-SIniutin/internal/cache"
)

func CacheMiddleware(c cache.Cache, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			key := "route:" + r.URL.RequestURI()

			cached, err := c.Get(r.Context(), key)
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cache", "HIT")
				w.Write([]byte(cached))
				return
			}

			rec := &statusRecorder{ResponseWriter: w, body: &bytes.Buffer{}, status: 200}
			next.ServeHTTP(rec, r)

			if rec.status == http.StatusOK {
				c.Set(r.Context(), key, rec.body.String(), ttl)
			}
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
