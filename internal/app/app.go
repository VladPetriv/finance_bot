package app

import (
	"fmt"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

// Run is used to start the application.
func Run(cfg *config.Config, logger *logger.Logger) {
	bot := bot.NewTelegramgBot(cfg.Telegram.BotToken)

	botAPI, err := bot.NewAPI()
	if err != nil {
		logger.Fatal().Err(err).Msg("create new bot api")
	}

	services := service.Services{
		MessageService: service.NewMessage(botAPI, logger),
		EventService:   service.NewEvent(botAPI, logger),
	}

	fmt.Printf("services: %v\n", services)
}
