package service

import (
	"fmt"

	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type keyboardService struct {
	botAPI bot.API
	logger *logger.Logger
}

var _ (KeyboardService) = (*keyboardService)(nil)

// NewKeyboard returns new instance of keyboard service.
func NewKeyboard(botAPI bot.API, logger *logger.Logger) *keyboardService {
	return &keyboardService{
		botAPI: botAPI,
		logger: logger,
	}
}

func (k keyboardService) CreateKeyboard(opts *CreateKeyboardOptions) error {
	logger := k.logger

	sendOptions := &bot.SendOptions{
		ChatID: opts.ChatID,
	}

	if opts.Type == keyboardTypeInline {
		sendOptions.InlineKeyboard = opts.Rows
	}
	if opts.Type == keyboardTypeRow {
		sendOptions.Keyboard = opts.Rows
	}
	logger.Info().Interface("sendOptons", sendOptions).Msg("built send options")

	err := k.botAPI.Send(sendOptions)
	if err != nil {
		logger.Error().Err(err).Msg("create keyboard")
		return fmt.Errorf("create keyboard: %w", err)
	}

	return nil
}
