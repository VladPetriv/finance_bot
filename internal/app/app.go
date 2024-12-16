package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

// Run is used to start the application.
func Run(ctx context.Context, cfg *config.Config, logger *logger.Logger) {
	b := bot.NewTelegramBot(cfg.Telegram.BotToken, cfg.Telegram.WebhookURL, cfg.Telegram.SeverAddress)

	botAPI, err := b.NewAPI()
	if err != nil {
		logger.Fatal().Err(err).Msg("create new bot api")
	}

	mongoDB, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("create new mongodb instance")
	}

	stores := service.Stores{
		Category:  store.NewCategory(mongoDB),
		User:      store.NewUser(mongoDB),
		Balance:   store.NewBalance(mongoDB),
		Operation: store.NewOperation(mongoDB),
	}

	messageService := service.NewMessage(botAPI, logger)
	keyboardService := service.NewKeyboard(botAPI, logger)
	categoryService := service.NewCategory(logger, stores.Category)
	userService := service.NewUser(logger, stores.User)
	balanceService := service.NewBalance(logger, stores.Balance)
	operationService := service.NewOperation(logger, stores.Operation, stores.Balance, stores.Category)

	services := service.Services{
		Message:   messageService,
		Keyboard:  keyboardService,
		Category:  categoryService,
		User:      userService,
		Balance:   balanceService,
		Operation: operationService,
	}

	handlerService := service.NewHandler(&service.HandlerOptions{
		Logger:   logger,
		Services: services,
		Stores:   stores,
	})

	eventService := service.NewEvent(&service.EventOptions{
		BotAPI:         botAPI,
		Logger:         logger,
		HandlerService: handlerService,
	})

	go eventService.Listen(ctx)
	logger.Info().Msg("application started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	err = botAPI.Close()
	if err != nil {
		logger.Error().Err(err).Msg("close telegram bot connection")
	}

	err = mongoDB.Close()
	if err != nil {
		logger.Error().Err(err).Msg("close mongodb connection")
	}

	logger.Info().Msg("application stopped")
}
