package main

import (
	"context"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/app"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

func main() {
	cfg := config.Get()
	ctx := context.Background()

	logger := logger.New(cfg.Logger.LogLevel, cfg.Logger.LogFilename)

	app.Run(ctx, cfg, logger)
}
