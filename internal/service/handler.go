package service

import (
	"encoding/json"
	"fmt"

	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type handlerService struct {
	logger          *logger.Logger
	messageService  MessageService
	keyboardService KeyboardService
}

var _ HandlerService = (*handlerService)(nil)

// HandlerOptions represents input options for new instance of handler service.
type HandlerOptions struct {
	Logger          *logger.Logger
	MessageService  MessageService
	KeyboardService KeyboardService
}

// NewHandler returns new instance of handler service.
func NewHandler(opts *HandlerOptions) *handlerService {
	return &handlerService{
		logger:          opts.Logger,
		messageService:  opts.MessageService,
		keyboardService: opts.KeyboardService,
	}
}

func (h handlerService) HandleEventStart(messageData []byte) error {
	logger := h.logger

	var msg HandleEventStartMessage

	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshall handle event start message")
	}
	logger.Debug().Interface("msg", msg).Msg("unmarshalled handle event start message")

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   fmt.Sprintf("Hello, @%s!\nWelcome to @FinanceTracking_bot!", msg.Message.From.Username),
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleEventStop(messageData []byte) error {
	return nil
}

func (h handlerService) HandleEventUnknown(messageData []byte) error {
	logger := h.logger

	var msg HandleEventUnknownMessage

	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshall handle event unknown message")
	}
	logger.Debug().Interface("msg", msg).Msg("unmarshalled handle event unknown message")

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   "Didn't understand you!\nCould you please check available commands!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
