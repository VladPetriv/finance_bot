package service

import (
	"fmt"

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

func (m messageService) SendMessage(opts *SendMessageOptions) error {
	logger := m.logger

	err := m.botAPI.Send(&bot.SendOptions{
		ChatID:  opts.ChantID,
		Message: opts.Text,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("successfully sent message")
	return nil
}
