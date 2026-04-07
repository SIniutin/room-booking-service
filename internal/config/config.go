package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type HTTPConfig struct {
	Port           string        `yaml:"port"            env:"HTTP_PORT"             env-default:"8080"`
	ReadTimeout    time.Duration `yaml:"read_timeout"    env:"HTTP_READ_TIMEOUT"     env-default:"10s"`
	WriteTimeout   time.Duration `yaml:"write_timeout"   env:"HTTP_WRITE_TIMEOUT"    env-default:"10s"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"    env:"HTTP_IDLE_TIMEOUT"     env-default:"60s"`
	AllowedOrigins []string      `yaml:"allowed_origins" env:"HTTP_ALLOWED_ORIGINS"  env-default:"http://localhost:3000,http://127.0.0.1:3000,http://localhost:5173,http://127.0.0.1:5173" env-separator:","`
}

type PostgresConfig struct {
	DSN string `yaml:"dsn" env:"POSTGRES_DSN" env-required:"true"`
}

type RedisConfig struct {
	Addr     string        `yaml:"addr"      env:"REDIS_ADDR"      env-default:"localhost:6379"`
	CacheTTL time.Duration `yaml:"cache_ttl" env:"REDIS_CACHE_TTL" env-default:"30s"`
}

type JWTConfig struct {
	PrivateKeyPath string `yaml:"private_key_path" env:"JWT_PRIVATE_KEY_PATH" env-required:"true"`
	PublicKeyPath  string `yaml:"public_key_path"  env:"JWT_PUBLIC_KEY_PATH"  env-required:"true"`
	TTL            string `yaml:"ttl"              env:"JWT_TTL"              env-default:"24h"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("config.yml", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}
