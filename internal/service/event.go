package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type eventService struct {
	botAPI         bot.API
	logger         *logger.Logger
	handlerService HandlerService
}

var _ EventService = (*eventService)(nil)

// EventOptions represents an input options for creating new instance of event service.
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
			var botMessage botMessage

			err := json.Unmarshal(update, &botMessage)
			if err != nil {
				logger.Error().Err(err).Msg("unmarshalled update data")

				continue
			}
			logger.Debug().Interface("botMessage", botMessage).Msg("unmarshalled base message")

			//  Exceeded max input count, we need to reset all fields related to previous event
			if previousEventInputCount == previousEventMaxInputCount+1 {
				logger.Info().Msg("exceeded max input count, reset all related data to previous event")
				previousEventName = ""
				previousEventInputCount = 0
				previousEventMaxInputCount = 0
			}

			// No need to get new event name if previous one was not processed to the end
			if previousEventName == "" {
				eventName = e.getEventNameFromMsg(&botMessage)
				logger.Debug().Interface("eventName", eventName).Msg("got event from message")
			}

			eventMaxInputCount, ok := eventsWithInput[eventName]
			if ok {
				logger.Info().Msg("got event with input")
				previousEventName = eventName
				previousEventMaxInputCount = eventMaxInputCount
			}

			// Need to process all input for previous event
			if previousEventName != "" && previousEventInputCount <= previousEventMaxInputCount {
				logger.Info().Msg("increase process inputs for event")
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

func (e eventService) getEventNameFromMsg(msg *botMessage) event {
	if !strings.Contains(strings.Join(availableCommands, " "), msg.Message.Text) {
		return unknownEvent
	}
	if !strings.Contains(strings.Join(availableCommands, " "), msg.CallbackQuery.Data) {
		return unknownEvent
	}

	textToCheck := msg.Message.Text

	if msg.CallbackQuery.Data != "" {
		textToCheck = msg.CallbackQuery.Data
	}

	for _, c := range availableCommands {
		if strings.Contains(c, textToCheck) {
			if eventFromCommand, ok := commandToEvent[c]; ok {
				return eventFromCommand
			}
		}
	}

	return unknownEvent
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

	case updateBalanceEvent, updateBalanceAmountEvent, updateBalanceCurrencyEvent:
		err := e.handlerService.HandleEventUpdateBalance(ctx, eventName, messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event update balance")
			return fmt.Errorf("handle event update balance: %w", err)
		}

	case backEvent:
		err := e.handlerService.HandleEventBack(ctx, messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event back")
			return fmt.Errorf("handle event back: %w", err)
		}

	case getBalanceEvent:
		err := e.handlerService.HandleEventGetBalance(ctx, messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event get balance")
			return fmt.Errorf("handle event get balance: %w", err)
		}

	case createOperationEvent, createIncomingOperationEvent, createSpendingOperationEvent:
		err := e.handlerService.HandleEventOperationCreate(ctx, eventName, messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event create operation")
			return fmt.Errorf("handle event create operation: %w", err)
		}

	case updateOperationAmountEvent:
		err := e.handlerService.HandleEventUpdateOperationAmount(ctx, messageData)
		if err != nil {
			logger.Error().Err(err).Msg("handle event update operation amount event")
			return fmt.Errorf("handle event update operation amount event: %w", err)
		}

	default:
		logger.Warn().Interface("eventName", eventName).Msg("didn't react on event")
		return nil
	}

	return nil
}
