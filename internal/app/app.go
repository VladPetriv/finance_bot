package app

import (
	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

// Run is used to start the application.
func Run(cfg *config.Config, logger *logger.Logger) {
	b := bot.NewTelegramgBot(cfg.Telegram.BotToken)

	botAPI, err := b.NewAPI()
	if err != nil {
		logger.Fatal().Err(err).Msg("create new bot api")
	}

	messageService := service.NewMessage(botAPI, logger)
	keyboardService := service.NewKeyboard(botAPI, logger)
	eventService := service.NewEvent(&service.EventOptinos{
		BotAPI:          botAPI,
		Logger:          logger,
		MessageService:  messageService,
		KeyboardService: keyboardService,
	})

	services := service.Services{
		MessageService:  messageService,
		KeyboardService: keyboardService,
		EventService:    eventService,
	}

	errs := make(chan error)
	updates := make(chan []byte)

	services.EventService.Listen(updates, errs)
}
