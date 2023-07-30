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
	logger.Debug().Interface("opts", opts).Msg("got args")

	sendOptions := &bot.SendOptions{
		ChatID:  opts.ChatID,
		Message: opts.Message,
	}

	if opts.Type == keyboardTypeInline {
		sendOptions.InlineKeyboard = opts.Rows
	}
	if opts.Type == keyboardTypeRow {
		sendOptions.Keyboard = opts.Rows
	}
	logger.Info().Interface("sendOptions", sendOptions).Msg("built send options")

	err := k.botAPI.Send(sendOptions)
	if err != nil {
		logger.Error().Err(err).Msg("create keyboard using botAPI")
		return fmt.Errorf("create keyboard using botAPI: %w", err)
	}

	logger.Info().Msg("created keyboard")
	return nil
}
