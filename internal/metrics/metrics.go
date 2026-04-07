package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ErrorsTotal     *prometheus.CounterVec
}

func New() *Metrics {
	return NewWithRegisterer(prometheus.DefaultRegisterer)
}

func NewWithRegisterer(registerer prometheus.Registerer) *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "total number of requests",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "app_errors_total",
				Help: "total number of errors",
			},
			[]string{"layer", "op"},
		),
	}

	registerer.MustRegister(
		m.RequestsTotal,
		m.RequestDuration,
		m.ErrorsTotal,
	)

	return m
}
