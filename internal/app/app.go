package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/VladPetriv/finance_bot/config"
	currencybeacon "github.com/VladPetriv/finance_bot/internal/api/currency_beacon"
	"github.com/VladPetriv/finance_bot/internal/api/gemini"
	"github.com/VladPetriv/finance_bot/internal/api/telegram"
	"github.com/VladPetriv/finance_bot/internal/migrations"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

// Run is used to start the application.
func Run(ctx context.Context, cfg *config.Config, logger *logger.Logger) {
	telegram, err := telegram.New(telegram.Options{
		Token:         cfg.Telegram.BotToken,
		UpdatesType:   cfg.Telegram.UpdatesType,
		ServerAddress: cfg.Telegram.SeverAddress,
		WebhookURL:    cfg.Telegram.WebhookURL,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("create new telegram api")
	}

	gemini, err := gemini.New(ctx, cfg.Gemini.APIKey, cfg.Gemini.Model)
	if err != nil {
		logger.Fatal().Err(err).Msg("create new gemini api")
	}

	apis := service.APIs{
		Messenger:         telegram,
		Prompter:          gemini,
		CurrencyExchanger: currencybeacon.New(cfg.CurrencyBeacon.APIEndpoint, cfg.CurrencyBeacon.APIKey),
	}

	postgres, err := database.NewPostgreSQL(database.PostgreSQLOptions{
		User:     cfg.PostgreSQL.User,
		Password: cfg.PostgreSQL.Password,
		Database: cfg.PostgreSQL.Database,
		Host:     cfg.PostgreSQL.Host,
		Port:     cfg.PostgreSQL.Port,
		SSLMode:  cfg.PostgreSQL.SSLMode,
		URL:      cfg.PostgreSQL.URL,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("create new postgres instance")
	}

	err = postgres.Ping()
	if err != nil {
		logger.Fatal().Err(err).Msg("connection with postgres is not established")
	}

	err = migrations.MigrateDB(logger, postgres.DB, cfg.PostgreSQL.Database, migrations.Migrations)
	if err != nil {
		logger.Fatal().Err(err).Msg("migrate database")
	}

	stores := service.Stores{
		Category:            store.NewCategory(postgres),
		User:                store.NewUser(postgres),
		Balance:             store.NewBalance(postgres),
		BalanceSubscription: store.NewBalanceSubscription(postgres),
		Operation:           store.NewOperation(postgres),
		State:               store.NewState(postgres),
		Currency:            store.NewCurrency(postgres),
	}

	currencyService := service.NewCurrency(logger, apis, stores)

	services := service.Services{
		State: service.NewState(&service.StateOptions{
			Logger: logger,
			Stores: stores,
			APIs:   apis,
		}),
		Currency:                  currencyService,
		BalanceSubscriptionEngine: service.NewBalanceSubscriptionEngine(logger, stores, apis),
	}

	handlerService := service.NewHandler(&service.HandlerOptions{
		Logger:   logger,
		Services: services,
		APIs:     apis,
		Stores:   stores,
	})
	handlerService.RegisterHandlers()
	services.Handler = handlerService

	eventService := service.NewEvent(&service.EventOptions{
		Logger:   logger,
		APIs:     apis,
		Services: services,
	})

	err = currencyService.InitCurrencies(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("init currencies")
	}

	go eventService.Listen(ctx)
	go services.BalanceSubscriptionEngine.CreateOperations(ctx)

	// Setup health check server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status": "ok"}`))
		if err != nil {
			logger.Error().Err(err).Msg("error writing response")
		}
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info().Msg("starting health check server on :8080")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("health check server error")
		}
	}()

	logger.Info().Msg("application started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	err = telegram.Close()
	if err != nil {
		logger.Error().Err(err).Msg("close telegram bot connection")
	}

	err = gemini.Close()
	if err != nil {
		logger.Error().Err(err).Msg("close gemini connection")
	}

	err = server.Shutdown(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("error shutting down health check server")
	}

	err = postgres.Close()
	if err != nil {
		logger.Error().Err(err).Msg("close postgres connection")
	}

	logger.Info().Msg("application stopped")
}
