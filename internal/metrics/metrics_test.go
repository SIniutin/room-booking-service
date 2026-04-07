package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWithRegisterer(t *testing.T) {
	registry := prometheus.NewRegistry()

	m := NewWithRegisterer(registry)

	assert.NotNil(t, m)
	assert.NotNil(t, m.RequestsTotal)
	assert.NotNil(t, m.RequestDuration)
	assert.NotNil(t, m.ErrorsTotal)

	m.RequestsTotal.WithLabelValues("GET", "/rooms/list", "200").Inc()
	m.RequestDuration.WithLabelValues("GET", "/rooms/list").Observe(0.1)
	m.ErrorsTotal.WithLabelValues("http", "listRooms").Inc()

	metric := &dto.Metric{}
	require.NoError(t, m.RequestsTotal.WithLabelValues("GET", "/rooms/list", "200").Write(metric))
	assert.Equal(t, 1.0, metric.GetCounter().GetValue())
}
