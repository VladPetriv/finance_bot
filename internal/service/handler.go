package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/google/uuid"
)

type handlerService struct {
	logger   *logger.Logger
	services Services
	apis     APIs
	stores   Stores
}

var _ HandlerService = (*handlerService)(nil)

// HandlerOptions represents input options for new instance of handler service.
type HandlerOptions struct {
	Logger   *logger.Logger
	Services Services
	APIs     APIs
	Stores   Stores
}

// NewHandler returns new instance of handler service.
func NewHandler(opts *HandlerOptions) *handlerService {
	return &handlerService{
		logger:   opts.Logger,
		services: opts.Services,
		apis:     opts.APIs,
		stores:   opts.Stores,
	}
}

func (h handlerService) HandleStart(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleStart").Logger()

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

	username := msg.GetSenderName()
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
		nextStep = models.EndFlowStep
		return h.sendMessageWithDefaultKeyboard(chatID, fmt.Sprintf("Happy to see you again @%s!", username))
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
		err := h.apis.Messenger.SendMessage(chatID, message)
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}
	}

	nextStep = models.CreateInitialBalanceFlowStep
	return nil
}

func (h handlerService) HandleCancel(ctx context.Context, msg Message) error {
	return h.sendMessageWithDefaultKeyboard(msg.GetChatID(), "Please choose command to execute:")
}

func (h handlerService) HandleUnknown(msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleUnknown").Logger()

	err := h.apis.Messenger.SendMessage(
		msg.GetChatID(),
		"Didn't understand you!\nCould you please check available commands!",
	)
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleError(ctx context.Context, opts HandleErrorOptions) error {
	logger := h.logger.With().Str("name", "handlerService.HandleError").Logger()

	if errs.IsExpected(opts.Err) {
		err := h.apis.Messenger.SendMessage(opts.Msg.GetChatID(), opts.Err.Error())
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		logger.Info().Msg("handled expected error")
		return nil
	}

	message := "Something went wrong!\nPlease try again later!"

	if opts.SendDefaultKeyboard {
		return h.sendMessageWithDefaultKeyboard(opts.Msg.GetChatID(), message)
	}

	return h.apis.Messenger.SendMessage(opts.Msg.GetChatID(), "Something went wrong!\nPlease try again later!")
}

func (h handlerService) HandleWrappers(ctx context.Context, event models.Event, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleWrappers").Logger()
	logger.Debug().Any("msg", msg).Any("event", event).Msg("got args")

	var (
		rows    []KeyboardRow
		message string
	)

	switch event {
	case models.BalanceEvent:
		rows = []KeyboardRow{
			{
				Buttons: []string{models.BotCreateBalanceCommand, models.BotGetBalanceCommand},
			},
			{
				Buttons: []string{models.BotUpdateBalanceCommand, models.BotDeleteBalanceCommand},
			},
			{
				Buttons: []string{models.BotCancelCommand},
			},
		}
		message = "Please choose balance command to execute:"
	case models.CategoryEvent:
		rows = []KeyboardRow{
			{
				Buttons: []string{models.BotCreateCategoryCommand, models.BotListCategoriesCommand},
			},
			{
				Buttons: []string{models.BotUpdateCategoryCommand, models.BotDeleteCategoryCommand},
			},
			{
				Buttons: []string{models.BotCancelCommand},
			},
		}
		message = "Please choose category command to execute:"
	case models.OperationEvent:
		rows = []KeyboardRow{
			{
				Buttons: []string{models.BotCreateOperationCommand, models.BotGetOperationsHistory},
			},
			{
				Buttons: []string{models.BotDeleteOperationCommand},
			},
			{
				Buttons: []string{models.BotCancelCommand},
			},
		}
		message = "Please choose operation command to execute:"
	default:
		return fmt.Errorf("unknown wrappers event: %s", event)
	}

	return h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   msg.GetChatID(),
		Message:  message,
		Keyboard: rows,
	})
}

// sendMessageWithConfirmationInlineKeyboard sends a message to the specified chat with Yes/No inline keyboard buttons.
func (h handlerService) sendMessageWithConfirmationInlineKeyboard(chatID int, message string) error {
	return h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:  chatID,
		Message: message,
		InlineKeyboard: []InlineKeyboardRow{
			{
				Buttons: []InlineKeyboardButton{
					{
						Text: "Yes",
						Data: "true",
					},
					{
						Text: "No",
						Data: "false",
					},
				},
			},
		},
	})
}

// notifyCancellationAndShowMenu sends a cancellation message and displays the main menu.
// It informs the user that their current action was cancelled and presents available commands
// through the default keyboard interface.
func (h handlerService) notifyCancellationAndShowMenu(chatID int) error {
	return h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   chatID,
		Message:  "Action cancelled!\nPlease choose new command to execute:",
		Keyboard: defaultKeyboardRows,
	})
}

const emptyMessage = "ㅤ"

// showCancelButton displays a single "Cancel" button in the chat interface,
// replacing any previous keyboard and sends a message if provided. This prevents users from interacting with
// outdated keyboard buttons that may still be visible from previous messages.
func (h handlerService) showCancelButton(chatID int, message string) error {
	if message == "" {
		message = emptyMessage
	}

	return h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   chatID,
		Message:  message,
		Keyboard: rowKeyboardWithCancelButtonOnly,
	})
}

type named interface {
	GetName() string
}

func getKeyboardRows[T named](data []T, elementLimitPerRow int, includeRowWithCancelButton bool) []KeyboardRow {
	keyboardRows := make([]KeyboardRow, 0)

	var currentRow KeyboardRow
	for i, entry := range data {
		currentRow.Buttons = append(currentRow.Buttons, entry.GetName())

		// When row is full or we're at the last data item, append row
		if len(currentRow.Buttons) == elementLimitPerRow || i == len(data)-1 {
			keyboardRows = append(keyboardRows, currentRow)
			currentRow = KeyboardRow{} // Reset current row
		}
	}

	if includeRowWithCancelButton {
		keyboardRows = append(keyboardRows, KeyboardRow{
			Buttons: []string{models.BotCancelCommand},
		})
	}

	return keyboardRows
}

// convertOperationsToInlineKeyboardRowsWithPagination converts a slice of operations into inline keyboard rows with pagination support.
// If there are more operations than the per-message limit, it adds a "Show More" button.
func convertOperationsToInlineKeyboardRowsWithPagination(operations []models.Operation, limitPerMessage int) []InlineKeyboardRow {
	inlineKeyboardRows := make([]InlineKeyboardRow, 0, len(operations))
	for _, operation := range operations {
		inlineKeyboardRows = append(inlineKeyboardRows, InlineKeyboardRow{
			Buttons: []InlineKeyboardButton{
				{
					Text: operation.GetName(),
					Data: operation.ID,
				},
			},
		})
	}

	// Skip BotShowMoreOperationsForDeleteCommand if operations are within the operationsPerMessage limit.
	if len(operations) <= limitPerMessage {
		return inlineKeyboardRows
	}

	inlineKeyboardRows = append(inlineKeyboardRows, []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotShowMoreOperationsForDeleteCommand,
				},
			},
		},
	}...)

	return inlineKeyboardRows
}

// sendMessageWithDefaultKeyboard sends a message to the specified chat with the default keyboard interface.
func (h handlerService) sendMessageWithDefaultKeyboard(chatID int, message string) error {
	return h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   chatID,
		Message:  message,
		Keyboard: defaultKeyboardRows,
	})
}
