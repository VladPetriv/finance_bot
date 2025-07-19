package service

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type eventService struct {
	logger   *logger.Logger
	apis     APIs
	services Services
}

var _ EventService = (*eventService)(nil)

// EventOptions represents an input options for creating new instance of event service.
type EventOptions struct {
	Logger   *logger.Logger
	APIs     APIs
	Services Services
}

// NewEvent returns new instance of event service.
func NewEvent(opts *EventOptions) *eventService {
	return &eventService{
		logger:   opts.Logger,
		apis:     opts.APIs,
		services: opts.Services,
	}
}

func (e eventService) Listen(ctx context.Context) {
	logger := e.logger.With().Str("name", "eventService.Listen").Logger()

	updatesCH := make(chan Message)
	errorsCH := make(chan error)

	go func() {
		defer close(updatesCH)
		defer close(errorsCH)
		e.apis.Messenger.ReadUpdates(updatesCH, errorsCH)
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("shutting down Listen")
			return
		case msg := <-updatesCH:
			e.handleMessage(ctx, msg)
		case err := <-errorsCH:
			logger.Error().Err(err).Msg("read updates")
		}
	}
}

func (e eventService) handleMessage(ctx context.Context, msg Message) {
	logger := e.logger.With().Str("name", "eventService.handleMessage").Logger()

	defer func() {
		if r := recover(); r != nil {
			e.handlePanic(ctx, msg, r)
		}
	}()

	stateOutput, err := e.services.State.HandleState(ctx, msg)
	if err != nil {
		logger.Error().Err(err).Msg("handle state")
		return
	}
	if stateOutput == nil {
		return
	}
	logger.Debug().Any("stateOutput", stateOutput).Msg("handled request state")

	msgCtx := context.WithValue(ctx, contextFieldNameState, stateOutput.State)
	err = e.ReactOnEvent(msgCtx, stateOutput.Event, msg)
	if err != nil {
		logger.Error().Err(err).Msg("react on event")
		err := e.services.Handler.HandleError(msgCtx, HandleErrorOptions{
			Err: err,
			Msg: msg,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle error")
		}
	}
}

func (e eventService) handlePanic(ctx context.Context, msg Message, r any) {
	logger := e.logger.With().Str("name", "eventService.handlePanic").Logger()
	logger.Error().
		Any("panic", r).
		Str("stack", string(debug.Stack())).
		Msg("recovered from panic")

	if msg.GetSenderName() != "" {
		err := e.services.State.DeleteState(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("delete state")
		}
	}

	err := e.services.Handler.HandleError(ctx, HandleErrorOptions{
		Err:                 fmt.Errorf("internal error"),
		Msg:                 msg,
		SendDefaultKeyboard: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("handle error")
	}
}

func getEventFromMsg(user *models.User, msg Message) models.Event {
	aiParserEnabled := user != nil && user.Settings != nil && user.Settings.AIParserEnabled
	inputIsNotACommand := !strings.Contains(strings.Join(models.AvailableCommands, " "), msg.GetText())

	if aiParserEnabled && inputIsNotACommand {
		return models.CreateOperationsThroughOneTimeInputEvent
	}

	for _, command := range models.AvailableCommands {
		if command == msg.GetText() {
			if eventFromCommand, ok := models.CommandToEvent[command]; ok {
				return eventFromCommand
			}
		}
	}

	return models.UnknownEvent
}

func (e eventService) ReactOnEvent(ctx context.Context, event models.Event, msg Message) error {
	logger := e.logger.With().Str("name", "eventService.ReactOnEvent").Logger()

	switch event {
	case models.StartEvent:
		err := e.services.Handler.HandleStart(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event start")
			return fmt.Errorf("handle event start: %w", err)
		}

	case models.UnknownEvent:
		err := e.services.Handler.HandleUnknown(msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle event unknown")
			return fmt.Errorf("handle event event unknown: %w", err)
		}

	case models.CancelEvent:
		err := e.services.Handler.HandleCancel(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event cancel")
			return fmt.Errorf("handle event cancel: %w", err)
		}

	case models.BackEvent:
		err := e.services.Handler.HandleBack(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle event back")
			return fmt.Errorf("handle event back: %w", err)
		}

	case models.BalanceEvent, models.CategoryEvent, models.OperationEvent, models.BalanceSubscriptionEvent:
		err := e.services.Handler.HandleWrappers(ctx, event, msg)
		if err != nil {
			logger.Error().Err(err).Msg("handle wrappers")
			return fmt.Errorf("handle wrappers: %w", err)
		}

	case models.CreateBalanceEvent, models.GetBalanceEvent, models.UpdateBalanceEvent, models.DeleteBalanceEvent,
		models.CreateCategoryEvent, models.ListCategoriesEvent, models.UpdateCategoryEvent, models.DeleteCategoryEvent,
		models.CreateOperationEvent, models.GetOperationsHistoryEvent, models.DeleteOperationEvent, models.UpdateOperationEvent,
		models.CreateBalanceSubscriptionEvent, models.ListBalanceSubscriptionEvent, models.UpdateBalanceSubscriptionEvent, models.DeleteBalanceSubscriptionEvent,
		models.CreateOperationsThroughOneTimeInputEvent:
		err := e.services.Handler.HandleAction(ctx, msg)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}
			logger.Error().Err(err).Msg("handle action")
			return fmt.Errorf("handle action: %w", err)
		}

	default:
		logger.Error().Any("event", event).Msg("receive unexpected event")
		return fmt.Errorf("receive unexpected event: %v", event)
	}

	return nil
}
