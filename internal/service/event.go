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

	var eventName, previousEventName event
	var previousEventInputCount, previousEventMaxInputCount int

	for {
		select {
		case update := <-updates:
			var msg botMessage

			err := json.Unmarshal(update, &msg)
			if err != nil {
				logger.Error().Err(err).Msg("unmarshal incoming update data")

				continue
			}
			logger.Debug().Interface("msg", msg).Msg("unmarshalled incoming update data")

			//  Exceeded max input count, we need to reset all fields related to previous event
			if previousEventInputCount == previousEventMaxInputCount+1 {
				logger.Info().Msg("exceeded max input count, reset all related data to previous event")
				previousEventName = ""
				previousEventInputCount = 0
				previousEventMaxInputCount = 0
			}

			// No need to get new event name if previous one was not processed to the end
			if previousEventName == "" {
				eventName = e.getEventNameFromMsg(&msg)
				logger.Info().Interface("eventName", eventName).Msg("got event from message")
			}

			eventMaxInputCount, ok := eventsWithInput[eventName]
			if ok {
				logger.Info().Interface("eventName", eventName).Msg("got event with input")
				previousEventName = eventName
				previousEventMaxInputCount = eventMaxInputCount
			}

			// Need to process all input for previous event
			if previousEventName != "" && previousEventInputCount <= previousEventMaxInputCount {
				logger.Info().Msg("increase input count for previous event")
				eventName = previousEventName
				previousEventInputCount++
			}

			err = e.ReactOnEvent(ctx, eventName, msg)
			if err != nil {
				logger.Error().Err(err).Msg("react on event")

				handleErr := e.handlerService.HandleError(ctx, msg)
				if handleErr != nil {
					logger.Error().Err(err).Msg("react on event")
				}
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

func (e eventService) ReactOnEvent(ctx context.Context, eventName event, msg botMessage) error {
	logger := e.logger
	logger.Debug().
		Interface("eventName", eventName).
		Interface("msg", msg).
		Msg("got args")

	switch eventName {
	case startEvent:
		err := e.handlerService.HandleEventStart(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event start")
			return fmt.Errorf("handle event start: %w", err)
		}

	case unknownEvent:
		err := e.handlerService.HandleEventUnknown(msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event unknown")
			return fmt.Errorf("handle event event unknown: %w", err)
		}

	case backEvent:
		err := e.handlerService.HandleEventBack(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event back")
			return fmt.Errorf("handle event back: %w", err)
		}

	case createCategoryEvent:
		err := e.handlerService.HandleEventCategoryCreate(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event category create")
			return fmt.Errorf("handle event category create: %w", err)
		}

	case listCategoryEvent:
		err := e.handlerService.HandleEventListCategories(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event list categories")
			return fmt.Errorf("handle event list categories: %w", err)
		}

	case updateBalanceEvent, updateBalanceAmountEvent, updateBalanceCurrencyEvent:
		err := e.handlerService.HandleEventUpdateBalance(ctx, eventName, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event update balance")
			return fmt.Errorf("handle event update balance: %w", err)
		}

	case getBalanceEvent:
		err := e.handlerService.HandleEventGetBalance(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event get balance")
			return fmt.Errorf("handle event get balance: %w", err)
		}

	case createOperationEvent, createIncomingOperationEvent, createSpendingOperationEvent:
		err := e.handlerService.HandleEventOperationCreate(ctx, eventName, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event operation create")
			return fmt.Errorf("handle event operation create: %w", err)
		}

	case updateOperationAmountEvent:
		err := e.handlerService.HandleEventUpdateOperationAmount(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event update operation amount")
			return fmt.Errorf("handle event update operation amount: %w", err)
		}

	case getOperationsHistoryEvent:
		err := e.handlerService.HandleEventGetOperationsHistory(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event get operations history")
			return fmt.Errorf("handle event get operations history: %w", err)
		}

	default:
		logger.Warn().Interface("eventName", eventName).Msg("receive unexpected event")
		return fmt.Errorf("receive unexpected event: %v", eventName)
	}

	return nil
}
