package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "/tmp/private.pem")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "/tmp/public.pem")
	t.Setenv("HTTP_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")

	_, err := os.Stat("config.yml")
	require.Error(t, err)

	cfg, loadErr := Load()

	require.NoError(t, loadErr)
	assert.Equal(t, "postgres://test", cfg.Postgres.DSN)
	assert.Equal(t, "redis:6379", cfg.Redis.Addr)
	assert.Equal(t, "/tmp/private.pem", cfg.JWT.PrivateKeyPath)
	assert.Equal(t, []string{"http://localhost:3000", "http://localhost:5173"}, cfg.HTTP.AllowedOrigins)
}
