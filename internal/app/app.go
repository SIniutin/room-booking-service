package app

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/internships-backend/test-backend-SIniutin/internal/cache"
	"github.com/internships-backend/test-backend-SIniutin/internal/config"
	"github.com/internships-backend/test-backend-SIniutin/internal/metrics"
	"github.com/internships-backend/test-backend-SIniutin/internal/repository/pg"
	chimux "github.com/internships-backend/test-backend-SIniutin/internal/transport/http"
	"github.com/internships-backend/test-backend-SIniutin/internal/usecase"
)

type App struct {
	cfg            *config.Config
	logger         *zap.Logger
	server         *http.Server
	slotGenerator  *usecase.SlotGenerator
	db             *pgxpool.Pool
	sqlDB          *sqlDBCloser
	cache          cache.Cache
	stopBackground context.CancelFunc
}

type sqlDBCloser struct {
	closeFn func() error
}

func (c *sqlDBCloser) Close() error {
	if c == nil || c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

func New(cfg *config.Config, logger *zap.Logger) (*App, error) {
	// postgres
	db, err := pgxpool.New(context.Background(), cfg.Postgres.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	// migrations
	sqlDB, err := goose.OpenDBWithDriver("pgx", cfg.Postgres.DSN)
	if err != nil {
		return nil, fmt.Errorf("goose open: %w", err)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return nil, fmt.Errorf("goose up: %w", err)
	}

	// redis / cache
	rdb, err := cache.NewCache(context.Background(), cfg.Redis.Addr)
	if err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}

	// jwt keys
	privateKey, err := loadPrivateKey(cfg.JWT.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}
	publicKey, err := loadPublicKey(cfg.JWT.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load public key: %w", err)
	}

	ttl, err := time.ParseDuration(cfg.JWT.TTL)
	if err != nil {
		return nil, fmt.Errorf("parse jwt ttl: %w", err)
	}

	// metrics
	registry := prometheus.NewRegistry()
	m := metrics.NewWithRegisterer(registry)

	// repositories
	userRepo := pg.NewUserRepo(db)
	roomRepo := pg.NewRoomRepo(db)
	scheduleRepo := pg.NewScheduleRepo(db)
	slotRepo := pg.NewSlotRepo(db)
	bookingRepo := pg.NewBookingRepo(db)

	// usecases
	authUsecase := usecase.NewAuthUsecase(privateKey, userRepo, ttl)
	slotGenerator := usecase.NewSlotGenerator(scheduleRepo, slotRepo)
	roomUsecase := usecase.NewRoomUsecase(roomRepo, scheduleRepo, slotRepo, slotGenerator)
	bookingUsecase := usecase.NewBookingUsecase(bookingRepo, slotRepo, rdb)

	// handler + router
	handler := chimux.NewHandler(authUsecase, roomUsecase, bookingUsecase, rdb, logger, m, cfg.Redis.CacheTTL)
	router := chimux.NewRouter(handler, publicKey, cfg.HTTP.AllowedOrigins, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &App{
		cfg:            cfg,
		logger:         logger,
		server:         server,
		slotGenerator:  slotGenerator,
		db:             db,
		sqlDB:          &sqlDBCloser{closeFn: sqlDB.Close},
		cache:          rdb,
		stopBackground: nil,
	}, nil
}

func (a *App) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	bgCtx, cancelBackground := context.WithCancel(context.Background())
	a.stopBackground = cancelBackground

	go func() {
		a.logger.Info("server started", zap.String("addr", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("server error", zap.Error(err))
		}
	}()

	go a.runSlotGenerator(bgCtx)

	<-quit
	a.logger.Info("shutting down...")

	cancelBackground()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var shutdownErr error
	if err := a.server.Shutdown(ctx); err != nil {
		shutdownErr = errors.Join(shutdownErr, err)
	}
	if err := a.cache.Close(); err != nil {
		shutdownErr = errors.Join(shutdownErr, err)
	}
	if err := a.sqlDB.Close(); err != nil {
		shutdownErr = errors.Join(shutdownErr, err)
	}
	a.db.Close()

	return shutdownErr
}

func (a *App) runSlotGenerator(ctx context.Context) {
	if err := a.slotGenerator.GenerateAll(ctx, time.Now().UTC(), 7); err != nil {
		a.logger.Error("slot generator: initial run failed", zap.Error(err))
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			from := time.Now().UTC().AddDate(0, 0, 7)
			if err := a.slotGenerator.GenerateAll(ctx, from, 1); err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("slot generator: daily run failed", zap.Error(err))
			}
		}
	}
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaKey, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA public key")
		}
		return rsaPub, nil
	}

	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub, nil
}
