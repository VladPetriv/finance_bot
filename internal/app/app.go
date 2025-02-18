package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/VladPetriv/finance_bot/config"
	currencybeacon "github.com/VladPetriv/finance_bot/internal/api/currency_beacon"
	"github.com/VladPetriv/finance_bot/internal/api/telegram"
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

	apis := service.APIs{
		Messenger:         telegram,
		CurrencyExchanger: currencybeacon.New(cfg.CurrencyBeacon.APIEndpoint, cfg.CurrencyBeacon.APIKey),
	}

	mongoDB, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("create new mongodb instance")
	}

	err = mongoDB.Ping(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("connection with mondodb is not established")
	}

	stores := service.Stores{
		Category:  store.NewCategory(mongoDB),
		User:      store.NewUser(mongoDB),
		Balance:   store.NewBalance(mongoDB),
		Operation: store.NewOperation(mongoDB),
		State:     store.NewState(mongoDB),
		Currency:  store.NewCurrency(mongoDB),
	}

	currencyService := service.NewCurrency(logger, apis, stores)

	stateService := service.NewState(&service.StateOptions{
		Logger: logger,
		Stores: stores,
		APIs:   apis,
	})

	services := service.Services{
		State:    stateService,
		Currency: currencyService,
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

	err = server.Shutdown(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("error shutting down health check server")
	}

	err = mongoDB.Close()
	if err != nil {
		logger.Error().Err(err).Msg("close mongodb connection")
	}

	logger.Info().Msg("application stopped")
}
