package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/internships-backend/test-backend-SIniutin/internal/app"
	"github.com/internships-backend/test-backend-SIniutin/internal/config"
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
	defer logger.Sync()

	a, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("init failed", zap.Error(err))
	}

	if err := a.Run(); err != nil {
		logger.Fatal("app stopped", zap.Error(err))
	}
}
