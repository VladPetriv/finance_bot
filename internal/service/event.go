package service

import (
	"context"
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

func (e eventService) Listen(ctx context.Context, updates chan []byte, errs chan error) {
	logger := e.logger

	go e.botAPI.ReadUpdates(updates, errs)

	var eventName event
	var previousEventName event
	var previousEventInputCount int
	var previousEventMaxInputCount int

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

			//  Exceeded max input count, we need to reset all fields related to previous event
			if previousEventInputCount == previousEventMaxInputCount+1 {
				previousEventName = ""
				previousEventInputCount = 0
				previousEventMaxInputCount = 0
			}

			// No need to get new event name if previous one was not processed to the end
			if previousEventName == "" {
				eventName = e.getEventNameFromMsg(&baseMessage)
				logger.Debug().Interface("eventName", eventName).Msg("got event from message")
			}

			eventMaxInputCount, ok := eventsWithInput[eventName]
			if ok {
				previousEventName = eventName
				previousEventMaxInputCount = eventMaxInputCount
			}

			// Need to process all input for previous event
			if previousEventName != "" && previousEventInputCount <= previousEventMaxInputCount {
				eventName = previousEventName
				previousEventInputCount++
			}

			err = e.ReactOnEvent(ctx, eventName, update)
			if err != nil {
				logger.Error().Err(err).Msg("react on event")
			}
		case err := <-errs:
			logger.Error().Err(err).Msg("read updates")
		}
	}
}

func (e eventService) getEventNameFromMsg(msg *BaseMessage) event {
	if len(msg.Message.Entities) == 0 {
		return unknownEvent
	}

	// Got not a bot command
	if !msg.Message.Entities[0].IsBotCommand() {
		return ""
	}

	switch msg.Message.Text {
	case botStartCommand:
		return startEvent
	case botCreateCategoryCommand:
		return createCategoryEvent
	case botListCategoriesCommand:
		return listCategoryEvent
	case botUpdateBalanceCommand:
		return updateBalanceEvent
	case botGetBalanceCommand:
		return getBalanceEvent
	default:
		return unknownEvent
	}
}

func (e eventService) ReactOnEvent(ctx context.Context, eventName event, messageData []byte) error {
	logger := e.logger

	switch eventName {
	case startEvent:
		err := e.handlerService.HandleEventStart(ctx, messageData)
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

	case createCategoryEvent:
		err := e.handlerService.HandleEventCategoryCreate(ctx, messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event create category")
			return fmt.Errorf("handle event create category: %w", err)
		}

	case listCategoryEvent:
		err := e.handlerService.HandleEventListCategories(ctx, messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event list categories")
			return fmt.Errorf("handle event list categories: %w", err)
		}

	case updateBalanceEvent:
		err := e.handlerService.HandleEventUpdateBalance(ctx, messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event update balance")
			return fmt.Errorf("handle event update balance: %w", err)
		}

	case getBalanceEvent:
		err := e.handlerService.HandleEventGetBalance(ctx, messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event get balance")
			return fmt.Errorf("handle event get balance: %w", err)
		}

	default:
		logger.Warn().Interface("eventName", eventName).Msg("didn't react on event")
		return nil
	}

	return nil
}
