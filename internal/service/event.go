package service

import (
	"encoding/json"
	"fmt"

	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type eventService struct {
	botAPI         bot.API
	logger         *logger.Logger
	handlerService HandlerService
}

var _ EventService = (*eventService)(nil)

// EventOptions represents input options for new instance of event service.
type EventOptions struct {
	BotAPI         bot.API
	Logger         *logger.Logger
	HandlerService HandlerService
}

// NewEvent returns new instance of event service.
func NewEvent(opts *EventOptions) *eventService {
	return &eventService{
		botAPI:         opts.BotAPI,
		logger:         opts.Logger,
		handlerService: opts.HandlerService,
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

			eventName := e.getEventNameFromMsg(&baseMessage)
			logger.Debug().Interface("eventName", eventName).Msg("got event from message")

			err = e.ReactOnEvent(eventName, update)
			if err != nil {
				logger.Error().Err(err).Msg("react on event")
			}

		case err := <-errs:
			logger.Error().Err(err).Msg("read updates")
		}
	}
}

// BotCommand represents the key used in a message to indicate that
// it contains a command for the bot to execute.
const botCommand = "bot_command"

func (e eventService) getEventNameFromMsg(msg *BaseMessage) event {
	if msg.Message.Text == botStartCommand && isBotCommand(msg.Message.Entities) {
		return startEvent
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

func (e eventService) ReactOnEvent(eventName event, messageData []byte) error {
	logger := e.logger

	switch eventName {
	case startEvent:
		err := e.handlerService.HandleEventStart(messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event start")
			return fmt.Errorf("handle event start: %w", err)
		}
	case unknownEvent:
		err := e.handlerService.HandleEventUnknown(messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event start")
			return fmt.Errorf("handle event start: %w", err)
		}
	default:
		logger.Info().Interface("eventName", eventName).Msg("didn't react on event")
		return nil
	}

	return nil
}
