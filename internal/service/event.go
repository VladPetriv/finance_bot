package service

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type eventService struct {
	botAPI         bot.API
	logger         *logger.Logger
	handlerService HandlerService
	stateService   StateService
}

var _ EventService = (*eventService)(nil)

// EventOptions represents an input options for creating new instance of event service.
type EventOptions struct {
	BotAPI         bot.API
	Logger         *logger.Logger
	HandlerService HandlerService
	StateService   StateService
}

// NewEvent returns new instance of event service.
func NewEvent(opts *EventOptions) *eventService {
	return &eventService{
		botAPI:         opts.BotAPI,
		logger:         opts.Logger,
		handlerService: opts.HandlerService,
		stateService:   opts.StateService,
	}
}

func (e eventService) Listen(ctx context.Context) {
	logger := e.logger.With().Str("name", "eventService.Listen").Logger()

	updatesCH := make(chan []byte)
	errorsCH := make(chan error)

	go e.botAPI.ReadUpdates(updatesCH, errorsCH)

	for {
		func() {
			var msg botMessage
			defer func() {
				if r := recover(); r != nil {
					logger.Error().
						Any("panic", r).
						Str("stack", string(debug.Stack())).
						Msg("recovered from panic while processing bot update")
				}

				if msg.GetUsername() != "" {
					err := e.stateService.DeleteState(ctx, msg)
					if err != nil {
						logger.Error().Err(err).Msg("delete state")
					}
				}

				handleErr := e.handlerService.HandleError(ctx, HandleErrorOptions{
					Err:                 fmt.Errorf("internal error"),
					Msg:                 msg,
					SendDefaultKeyboard: true,
				})
				if handleErr != nil {
					logger.Error().Err(handleErr).Msg("handle error")
				}
			}()

			select {
			case update := <-updatesCH:

				err := json.Unmarshal(update, &msg)
				if err != nil {
					logger.Error().Err(err).Msg("unmarshal incoming update data")

					return
				}
				logger.Debug().Any("msg", msg).Msg("unmarshalled incoming update data")

				stateOutput, err := e.stateService.HandleState(ctx, msg)
				if err != nil {
					logger.Error().Err(err).Msg("handle state")

					return
				}
				logger.Debug().Any("stateOutput", stateOutput).Msg("handled request state")

				ctx = context.WithValue(ctx, contextFieldNameState, stateOutput.State)
				err = e.ReactOnEvent(ctx, stateOutput.Event, msg)
				if err != nil {
					logger.Error().Err(err).Msg("react on event")

					handleErr := e.handlerService.HandleError(ctx, HandleErrorOptions{
						Err: err,
						Msg: msg,
					})
					if handleErr != nil {
						logger.Error().Err(handleErr).Msg("handle error")
					}
				}
			case err := <-errorsCH:
				logger.Error().Err(err).Msg("read updates")
			}
		}()
	}
}

func getEventFromMsg(msg *botMessage) models.Event {
	if !strings.Contains(strings.Join(models.AvailableCommands, " "), msg.Message.Text) {
		return models.UnknownEvent
	}
	if !strings.Contains(strings.Join(models.AvailableCommands, " "), msg.CallbackQuery.Data) {
		return models.UnknownEvent
	}

	textToCheck := msg.Message.Text

	if msg.CallbackQuery.Data != "" {
		textToCheck = msg.CallbackQuery.Data
	}

	for _, c := range models.AvailableCommands {
		if strings.Contains(c, textToCheck) {
			if eventFromCommand, ok := models.CommandToEvent[c]; ok {
				return eventFromCommand
			}
		}
	}

	return models.UnknownEvent
}

func (e eventService) ReactOnEvent(ctx context.Context, event models.Event, msg botMessage) error {
	logger := e.logger.With().Str("name", "eventService.ReactOnEvent").Logger()

	switch event {
	case models.StartEvent:
		err := e.handlerService.HandleEventStart(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event start")
			return fmt.Errorf("handle event start: %w", err)
		}

	case models.CreateBalanceEvent:
		err := e.handlerService.HandleEventBalanceCreated(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event balance created")
			return fmt.Errorf("handle event balance created: %w", err)
		}

	case models.UpdateBalanceEvent:
		err := e.handlerService.HandleEventBalanceUpdated(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event balance created")
			return fmt.Errorf("handle event balance created: %w", err)
		}

	case models.GetBalanceEvent:
		err := e.handlerService.HandleEventGetBalance(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event get balance")
			return fmt.Errorf("handle event get balance: %w", err)
		}

	case models.DeleteBalanceEvent:
		err := e.handlerService.HandleEventBalanceDeleted(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event balance deleted")
			return fmt.Errorf("handle event balance deleted: %w", err)
		}

	case models.CreateCategoryEvent:
		err := e.handlerService.HandleEventCategoryCreated(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event category created")
			return fmt.Errorf("handle event category created: %w", err)
		}

	case models.ListCategoriesEvent:
		err := e.handlerService.HandleEventListCategories(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event list categories")
			return fmt.Errorf("handle event list categories: %w", err)
		}

	case models.UpdateCategoryEvent:
		err := e.handlerService.HandleEventCategoryUpdated(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}
			logger.Error().Err(err).Msg("handle event category updated")
			return fmt.Errorf("handle event category updated: %w", err)
		}

	case models.DeleteCategoryEvent:
		err := e.handlerService.HandleEventCategoryDeleted(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}
			logger.Error().Err(err).Msg("handle event category deleted")
			return fmt.Errorf("handle event category deleted: %w", err)
		}

	case models.CreateOperationEvent:
		err := e.handlerService.HandleEventOperationCreated(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event operation created")
			return fmt.Errorf("handle event operation created: %w", err)
		}

	case models.GetOperationsHistoryEvent:
		err := e.handlerService.HandleEventGetOperationsHistory(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event get operations history")
			return fmt.Errorf("handle event get operations history: %w", err)
		}

	case models.UnknownEvent:
		err := e.handlerService.HandleEventUnknown(msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event unknown")
			return fmt.Errorf("handle event event unknown: %w", err)
		}

	case models.BackEvent:
		err := e.handlerService.HandleEventBack(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event back")
			return fmt.Errorf("handle event back: %w", err)
		}

	default:
		logger.Error().Any("event", event).Msg("receive unexpected event")
		return fmt.Errorf("receive unexpected event: %v", event)
	}

	return nil
}
