package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
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
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	var nextStep models.FlowStep
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		state.Steps = append(state.Steps, nextStep)
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Interface("updatedState", updatedState).Msg("updated state in store")
	}()

	username := msg.GetUsername()
	chatID := msg.GetChatID()
	logger.Debug().
		Str("username", username).
		Int64("chatID", chatID).
		Msg("got username and chat id from incoming message")

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
			logger.Error().Err(err).Msg("create keyboard with welcome message")
			return fmt.Errorf("create keyboard with welcome message: %w", err)
		}

		nextStep = models.EndFlowStep

		logger.Info().Msg("user already exists")
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

	logger.Info().Msg("handled event start")
	return nil
}

func (h handlerService) HandleEventBack(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  msg.Message.Chat.ID,
		Message: "Please choose command to execute:",
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create row keyboard")
		return fmt.Errorf("create row keyboard: %w", err)
	}

	logger.Info().Msg("handled event back")
	return nil
}

func (h handlerService) HandleEventUnknown(msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	err := h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   "Didn't understand you!\nCould you please check available commands!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("handled event back")
	return nil
}

func (h handlerService) HandleError(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	err := h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: msg.GetChatID(),
		Text:   "Something went wrong!\nPlease try again later!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("handled error")
	return nil
}

type named interface {
	GetName() string
}

const maxBalancesPerRow = 3

func convertSliceToKeyboardRows[T named](data []T) []bot.KeyboardRow {
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

	return keyboardRows
}
