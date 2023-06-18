package app

import (
	"context"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

// Run is used to start the application.
func Run(ctx context.Context, cfg *config.Config, logger *logger.Logger) {
	b := bot.NewTelegramBot(cfg.Telegram.BotToken)

	botAPI, err := b.NewAPI()
	if err != nil {
		logger.Fatal().Err(err).Msg("create new bot api")
	}

	mongoDB, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("create new mongodb instance")
	}

	stores := service.Stores{
		Category: store.NewCategory(mongoDB),
		User:     store.NewUserStore(mongoDB),
		Balance:  store.NewBalance(mongoDB),
	}

	messageService := service.NewMessage(botAPI, logger)
	keyboardService := service.NewKeyboard(botAPI, logger)
	categoryService := service.NewCategory(logger, stores.Category)
	userService := service.NewUser(logger, stores.User)
	balanceService := service.NewBalance(logger, stores.Balance)

	handlerService := service.NewHandler(&service.HandlerOptions{
		Logger:          logger,
		MessageService:  messageService,
		KeyboardService: keyboardService,
		CategoryService: categoryService,
		UserService:     userService,
		BalanceStore:    stores.Balance,
		BalanceService:  balanceService,
	})
	eventService := service.NewEvent(&service.EventOptions{
		BotAPI:         botAPI,
		Logger:         logger,
		HandlerService: handlerService,
	})

	errs := make(chan error)
	updates := make(chan []byte)

	eventService.Listen(ctx, updates, errs)
}
