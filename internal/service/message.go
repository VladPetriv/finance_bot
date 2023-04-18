package service

import (
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type messageService struct {
	botAPI bot.API
	logger *logger.Logger
}

var _ MessageService = (*messageService)(nil)

// NewMessage returns new instance of message service.
func NewMessage(botAPI bot.API, logger *logger.Logger) *messageService {
	return &messageService{
		botAPI: botAPI,
		logger: logger,
	}
}

func (m messageService) SendMessage(chatID int64, message string) error {
	return nil
}
