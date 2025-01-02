package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

// Run is used to start the application.
func Run(ctx context.Context, cfg *config.Config, logger *logger.Logger) {
	fmt.Printf("cfg: %+v\n", cfg)

	// b := bot.NewTelegramBot(cfg.Telegram.BotToken, cfg.Telegram.WebhookURL, cfg.Telegram.SeverAddress)

	// botAPI, err := b.NewAPI()
	// if err != nil {
	// 	logger.Fatal().Err(err).Msg("create new bot api")
	// }

	mongoDB, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("create new mongodb instance")
	}

	// stores := service.Stores{
	// 	Category:  store.NewCategory(mongoDB),
	// 	User:      store.NewUser(mongoDB),
	// 	Balance:   store.NewBalance(mongoDB),
	// 	Operation: store.NewOperation(mongoDB),
	// 	State:     store.NewState(mongoDB),
	// }

	// messageService := service.NewMessage(botAPI, logger)
	// keyboardService := service.NewKeyboard(botAPI, logger)
	// stateService := service.NewState(&service.StateOptions{
	// 	Logger: logger,
	// 	Stores: stores,
	// })

	// services := service.Services{
	// 	Message:  messageService,
	// 	Keyboard: keyboardService,
	// 	State: stateService,
	// }

	// handlerService := service.NewHandler(&service.HandlerOptions{
	// 	Logger:   logger,
	// 	Services: services,
	// 	Stores:   stores,
	// })

	// eventService := service.NewEvent(&service.EventOptions{
	// 	// BotAPI:         botAPI,
	// 	Logger:         logger,
	// 	HandlerService: handlerService,
	// 	StateService:   stateService,
	// })

	// go eventService.Listen(ctx)

	// Setup health check server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	server := &http.Server{
		Addr:    ":8080", // You might want to make this configurable
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

	// err = botAPI.Close()
	// if err != nil {
	// 	logger.Error().Err(err).Msg("close telegram bot connection")
	// }

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
