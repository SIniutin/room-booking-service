package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/SIniutin/room-booking-service/internal/app"
	"github.com/SIniutin/room-booking-service/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	a, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("init failed", zap.Error(err))
	}

	if err := a.Run(); err != nil {
		logger.Fatal("app stopped", zap.Error(err))
	}
}
