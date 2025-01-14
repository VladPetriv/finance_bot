package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/google/uuid"
)

type handlerService struct {
	logger   *logger.Logger
	services Services
	stores   Stores
}

var _ HandlerService = (*handlerService)(nil)

// HandlerOptions represents input options for new instance of handler service.
type HandlerOptions struct {
	Logger   *logger.Logger
	Services Services
	Stores   Stores
}

// NewHandler returns new instance of handler service.
func NewHandler(opts *HandlerOptions) *handlerService {
	return &handlerService{
		logger:   opts.Logger,
		services: opts.Services,
		stores:   opts.Stores,
	}
}

func (h handlerService) HandleEventStart(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventStart").Logger()

	var nextStep models.FlowStep
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		state.Steps = append(state.Steps, nextStep)
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Any("updatedState", updatedState).Msg("updated state in store")
	}()

	username := msg.GetUsername()
	chatID := msg.GetChatID()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username: username,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}

	// Handle case when user already exists
	if user != nil {
		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  chatID,
			Message: fmt.Sprintf("Happy to see you again @%s!", username),
			Type:    keyboardTypeRow,
			Rows:    defaultKeyboardRows,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard")
			return fmt.Errorf("create keyboard: %w", err)
		}

		nextStep = models.EndFlowStep
		return nil
	}

	err = h.stores.User.Create(ctx, &models.User{
		ID:       uuid.NewString(),
		Username: username,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create user in store")
		return fmt.Errorf("create user in store: %w", err)
	}

	welcomeMessage := fmt.Sprintf("Hello, @%s!\nWelcome to @FinanceTracking_bot!", username)
	enterBalanceNameMessage := "Please enter the name of your initial balance!:"

	messagesToSend := []string{welcomeMessage, enterBalanceNameMessage}
	for _, message := range messagesToSend {
		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: chatID,
			Text:   message,
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}
	}

	nextStep = models.CreateInitialBalanceFlowStep
	return nil
}

func (h handlerService) HandleEventBack(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventBack").Logger()

	err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  msg.Message.Chat.ID,
		Message: "Please choose command to execute:",
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create keyboard")
		return fmt.Errorf("create keyboard: %w", err)
	}

	return nil
}

func (h handlerService) HandleEventUnknown(msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventUnknown").Logger()

	err := h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   "Didn't understand you!\nCould you please check available commands!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleError(ctx context.Context, receivedErr error, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleError").Logger()

	if errs.IsExpected(receivedErr) {
		err := h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: msg.GetChatID(),
			Text:   receivedErr.Error(),
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		logger.Info().Msg("handled expected error")
		return nil
	}

	err := h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: msg.GetChatID(),
		Text:   "Something went wrong!\nPlease try again later!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleWrappers(ctx context.Context, event models.Event, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleWrappers").Logger()
	logger.Debug().Any("msg", msg).Any("event", event).Msg("got args")

	var (
		rows    []bot.KeyboardRow
		message string
	)

	switch event {
	case models.BalanceEvent:

		rows = []bot.KeyboardRow{
			{
				Buttons: []string{models.BotCreateBalanceCommand, models.BotGetBalanceCommand},
			},
			{
				Buttons: []string{models.BotUpdateBalanceCommand, models.BotDeleteBalanceCommand},
			},
			{
				Buttons: []string{models.BotBackCommand},
			},
		}
		message = "Please choose balance command to execute:"
	case models.CategoryEvent:
		rows = []bot.KeyboardRow{
			{
				Buttons: []string{models.BotCreateCategoryCommand, models.BotListCategoriesCommand},
			},
			{
				Buttons: []string{models.BotUpdateCategoryCommand, models.BotDeleteCategoryCommand},
			},
			{
				Buttons: []string{models.BotBackCommand},
			},
		}
		message = "Please choose category command to execute:"
	case models.OperationEvent:
		rows = []bot.KeyboardRow{
			{
				Buttons: []string{models.BotCreateOperationCommand, models.BotGetOperationsHistory},
			},
			{
				Buttons: []string{models.BotDeleteOperationCommand},
			},
			{
				Buttons: []string{models.BotBackCommand},
			},
		}
		message = "Please choose operation command to execute:"
	default:
		return fmt.Errorf("unknown wrappers event: %s", event)
	}

	err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  msg.GetChatID(),
		Message: message,
		Type:    keyboardTypeRow,
		Rows:    rows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create keyboard")
		return fmt.Errorf("create keyboard: %w", err)
	}

	return nil
}

type named interface {
	GetName() string
}

const maxBalancesPerRow = 3

func getKeyboardRows[T named](data []T, includeRowWithBackButton bool) []bot.KeyboardRow {
	keyboardRows := make([]bot.KeyboardRow, 0)

	var currentRow bot.KeyboardRow
	for i, entry := range data {
		currentRow.Buttons = append(currentRow.Buttons, entry.GetName())

		// When row is full or we're at the last data item, append row
		if len(currentRow.Buttons) == maxBalancesPerRow || i == len(data)-1 {
			keyboardRows = append(keyboardRows, currentRow)
			currentRow = bot.KeyboardRow{} // Reset current row
		}
	}

	if includeRowWithBackButton {
		keyboardRows = append(keyboardRows, bot.KeyboardRow{
			Buttons: []string{models.BotBackCommand},
		})
	}

	return keyboardRows
}
