package service

import (
	"encoding/json"

	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type eventService struct {
	botAPI          bot.API
	logger          *logger.Logger
	messageService  MessageService
	keyboardService KeyboardService
}

var _ EventService = (*eventService)(nil)

// EventOptinos represents input optinos for new instance of event service.
type EventOptinos struct {
	BotAPI          bot.API
	Logger          *logger.Logger
	MessageService  MessageService
	KeyboardService KeyboardService
}

// NewEvent returns new instance of event service.
func NewEvent(opts *EventOptinos) *eventService {
	return &eventService{
		botAPI:          opts.BotAPI,
		logger:          opts.Logger,
		messageService:  opts.MessageService,
		keyboardService: opts.KeyboardService,
	}
}

func (e eventService) Listen(updates chan []byte, errs chan error) {
	logger := e.logger

	go e.botAPI.ReadUpdates(updates, errs)

	for {
		select {
		case update := <-updates:
			var baseMessage BaseMessage

			err := json.Unmarshal(update, &baseMessage)
			if err != nil {
				logger.Error().Err(err).Msg("unmarshalled update data")

				continue
			}
			logger.Debug().Interface("baseMessage", baseMessage).Msg("unmarshalled base message")

			eventName := e.getEventNameromMsg(&baseMessage)
			logger.Debug().Interface("eventName", eventName).Msg("got event from message")
		case err := <-errs:
			logger.Error().Err(err).Msg("read updates")
		}
	}
}

// BotCommand represents the key used in a message to indicate that
// it contains a command for the bot to execute.
const botCommand = "bot_command"

func (e eventService) getEventNameromMsg(msg *BaseMessage) event {
	if msg.Message.Text == botStartCommand && isBotCommand(msg.Message.Entities) {
		return startEvent
	}
	if msg.Message.Text == botStopCommand && isBotCommand(msg.Message.Entities) {
		return stopEvent
	}

	return unknownEvent
}

func isBotCommand(mesasgeEnitties []Entity) bool {
	if mesasgeEnitties == nil {
		return false
	}

	if mesasgeEnitties[0].Type == botCommand {
		return true
	}

	return false
}
