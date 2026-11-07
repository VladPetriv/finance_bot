package main

import (
	"context"
	_ "time/tzdata"

	_ "github.com/lib/pq"
	"github.com/vl-ptrv/finance_bot/config"
	"github.com/vl-ptrv/finance_bot/internal/app"
	"github.com/vl-ptrv/finance_bot/pkg/logger"
)

func main() {
	cfg := config.Get()
	ctx := context.Background()

	logger := logger.New(logger.LoggergerOptions{
		LogLevel:        cfg.Logger.LogLevel,
		LogFile:         cfg.Logger.LogFilename,
		PrettyLogOutput: cfg.Logger.PrettyLogOutput,
	})

	app.Run(ctx, cfg, logger)
}
