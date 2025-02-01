package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/google/uuid"
)

type flowStepHandlerFunc func(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error)

type handlerService struct {
	logger   *logger.Logger
	services Services
	apis     APIs
	stores   Stores

	flowWithFlowStepsHandlers map[models.Flow]map[models.FlowStep]flowStepHandlerFunc
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
	handler := handlerService{
		logger:   opts.Logger,
		services: opts.Services,
		apis:     opts.APIs,
		stores:   opts.Stores,
	}

	handler.flowWithFlowStepsHandlers = map[models.Flow]map[models.FlowStep]flowStepHandlerFunc{
		// Flows with balances
		models.StartFlow: {
			models.CreateInitialBalanceFlowStep: handler.handleCreateBalanceFlowStep,
			models.EnterBalanceAmountFlowStep:   handler.handleEnterBalanceAmountFlowStep,
			models.EnterBalanceCurrencyFlowStep: handler.handleEnterBalanceCurrencyFlowStep,
		},
		models.CreateBalanceFlow: {
			models.CreateBalanceFlowStep:        handler.handleCreateBalanceFlowStep,
			models.EnterBalanceNameFlowStep:     handler.handleEnterBalanceNameFlowStep,
			models.EnterBalanceAmountFlowStep:   handler.handleEnterBalanceAmountFlowStep,
			models.EnterBalanceCurrencyFlowStep: handler.handleEnterBalanceCurrencyFlowStep,
		},
		models.GetBalanceFlow: {
			models.GetBalanceFlowStep:    handler.handleGetBalanceFlowStep,
			models.ChooseBalanceFlowStep: handler.handleChooseBalanceFlowStepForGetBalance,
		},
		models.UpdateBalanceFlow: {
			models.UpdateBalanceFlowStep:        handler.handleUpdateBalanceFlowStep,
			models.ChooseBalanceFlowStep:        handler.handleChooseBalanceFlowStepForUpdate,
			models.EnterBalanceNameFlowStep:     handler.handleEnterBalanceNameFlowStep,
			models.EnterBalanceAmountFlowStep:   handler.handleEnterBalanceAmountFlowStep,
			models.EnterBalanceCurrencyFlowStep: handler.handleEnterBalanceCurrencyFlowStep,
		},
		models.DeleteBalanceFlow: {
			models.DeleteBalanceFlowStep:          handler.handleDeleteBalanceFlowStep,
			models.ConfirmBalanceDeletionFlowStep: handler.handleConfirmBalanceDeletionFlowStep,
			models.ChooseBalanceFlowStep:          handler.handleChooseBalanceFlowStepForDelete,
		},

		// Flows with categories
		models.CreateCategoryFlow: {
			models.CreateCategoryFlowStep:    handler.handleCreateCategoryFlowStep,
			models.EnterCategoryNameFlowStep: handler.handleEnterCategoryNameFlowStep,
		},
		models.ListCategoriesFlow: {
			models.ListCategoriesFlowStep: handler.handleListCategoriesFlowStep,
		},
		models.UpdateCategoryFlow: {
			models.UpdateCategoryFlowStep:           handler.handleUpdateCategoryFlowStep,
			models.ChooseCategoryFlowStep:           handler.handleChooseCategoryFlowStepForUpdate,
			models.EnterUpdatedCategoryNameFlowStep: handler.handleEnterUpdatedCategoryNameFlowStep,
		},
		models.DeleteCategoryFlow: {
			models.DeleteCategoryFlowStep: handler.handleDeleteCategoryFlowStep,
			models.ChooseCategoryFlowStep: handler.handleChooseCategoryFlowStepForDelete,
		},
	}

	return &handler
}

func (h handlerService) HandleError(ctx context.Context, opts HandleErrorOptions) error {
	logger := h.logger.With().Str("name", "handlerService.HandleError").Logger()

	if errs.IsExpected(opts.Err) {
		logger.Info().Err(opts.Err).Msg("handled expected error")
		return h.apis.Messenger.SendMessage(opts.Msg.GetChatID(), opts.Err.Error())
	}

	message := "Something went wrong!\nPlease try again later!"

	if opts.SendDefaultKeyboard {
		return h.sendMessageWithDefaultKeyboard(opts.Msg.GetChatID(), message)
	}

	return h.apis.Messenger.SendMessage(opts.Msg.GetChatID(), message)
}

func (h handlerService) HandleUnknown(msg Message) error {
	return h.sendMessageWithDefaultKeyboard(msg.GetChatID(), "Didn't understand you!\nCould you please check available commands!")
}

func (h handlerService) HandleStart(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleStart").Logger()

	var nextStep models.FlowStep
	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}
	defer func() {
		h.updateState(ctx, updateStateOptions{
			updatedStep:  nextStep,
			initialState: state,
		})
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

func (h handlerService) HandleWrappers(ctx context.Context, event models.Event, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleWrappers").Logger()
	logger.Debug().Any("msg", msg).Any("event", event).Msg("got args")

	var (
		rows    []KeyboardRow
		message string
	)

	switch event {
	case models.BalanceEvent:
		rows = balanceKeyboardRows
		message = "Please choose balance command to execute:"
	case models.CategoryEvent:
		rows = categoryKeyboardRows
		message = "Please choose category command to execute:"
	case models.OperationEvent:
		rows = operationKeyboardRows
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

func (h handlerService) HandleBalanceCreate(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleBalanceCreate").Logger()

	var nextStep models.FlowStep
	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}
	defer func() {
		h.updateState(ctx, updateStateOptions{
			updatedStep:  nextStep,
			initialState: state,
		})
	}()

	nextStep, err = h.processHandler(ctx, state, msg)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return err
		}
		logger.Error().Err(err).Msg("process handler")
		return fmt.Errorf("process handler: %w", err)
	}

	return nil
}

func (h handlerService) HandleBalanceGet(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleBalanceGet").Logger()

	var nextStep models.FlowStep
	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}
	defer func() {
		h.updateState(ctx, updateStateOptions{
			updatedStep:  nextStep,
			initialState: state,
		})
	}()

	nextStep, err = h.processHandler(ctx, state, msg)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return err
		}
		logger.Error().Err(err).Msg("process handler")
		return fmt.Errorf("process handler: %w", err)
	}

	return nil
}

func (h handlerService) HandleBalanceUpdate(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleBalanceUpdate").Logger()

	var nextStep models.FlowStep
	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}
	defer func() {
		h.updateState(ctx, updateStateOptions{
			updatedStep:  nextStep,
			initialState: state,
		})
	}()

	nextStep, err = h.processHandler(ctx, state, msg)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return err
		}
		logger.Error().Err(err).Msg("process handler")
		return fmt.Errorf("process handler: %w", err)
	}

	return nil
}

func (h handlerService) HandleBalanceDelete(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleBalanceDelete").Logger()

	var nextStep models.FlowStep
	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}
	defer func() {
		h.updateState(ctx, updateStateOptions{
			updatedStep:  nextStep,
			initialState: state,
		})
	}()

	nextStep, err = h.processHandler(ctx, state, msg)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return err
		}
		logger.Error().Err(err).Msg("process handler")
		return fmt.Errorf("process handler: %w", err)
	}

	return nil
}

func (h handlerService) HandleCategoryCreate(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleCategoryCreate").Logger()

	var nextStep models.FlowStep
	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}
	defer func() {
		h.updateState(ctx, updateStateOptions{
			updatedStep:  nextStep,
			initialState: state,
		})
	}()

	nextStep, err = h.processHandler(ctx, state, msg)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return err
		}
		logger.Error().Err(err).Msg("process handler")
		return fmt.Errorf("process handler: %w", err)
	}

	return nil
}

func (h handlerService) HandleCategoryList(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleCategoryList").Logger()

	var nextStep models.FlowStep
	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}
	defer func() {
		h.updateState(ctx, updateStateOptions{
			updatedStep:     nextStep,
			initialState:    state,
			updatedMetadata: state.Metedata,
		})
	}()

	nextStep, err = h.processHandler(ctx, state, msg)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return err
		}
		logger.Error().Err(err).Msg("process handler")
		return fmt.Errorf("process handler: %w", err)
	}

	return nil
}

func (h handlerService) HandleCategoryUpdate(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleCategoryUpdate").Logger()

	var nextStep models.FlowStep
	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}
	defer func() {
		h.updateState(ctx, updateStateOptions{
			updatedStep:     nextStep,
			initialState:    state,
			updatedMetadata: state.Metedata,
		})
	}()

	nextStep, err = h.processHandler(ctx, state, msg)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return err
		}
		logger.Error().Err(err).Msg("process handler")
		return fmt.Errorf("process handler: %w", err)
	}

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on update category flow")

	switch currentStep {
	case models.ChooseCategoryFlowStep:
		state.Metedata[previousCategoryTitleMetadataKey] = msg.GetText()
	}

	return nil
}

func (h handlerService) HandleCategoryDelete(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleCategoryDelete").Logger()

	var nextStep models.FlowStep
	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}
	defer func() {
		h.updateState(ctx, updateStateOptions{
			updatedStep:     nextStep,
			initialState:    state,
			updatedMetadata: state.Metedata,
		})
	}()

	nextStep, err = h.processHandler(ctx, state, msg)
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return err
		}
		logger.Error().Err(err).Msg("process handler")
		return fmt.Errorf("process handler: %w", err)
	}

	return nil
}

func (h handlerService) processHandler(ctx context.Context, state *models.State, message Message) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.processHandler").Logger()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username:        message.GetSenderName(),
		PreloadBalances: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return "", fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Info().Msg("user not found")
		return "", ErrUserNotFound
	}
	logger.Debug().Any("user", user).Msg("got user from store")

	currentStep := state.GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step")

	flowHandlers, ok := h.flowWithFlowStepsHandlers[state.Flow]
	if !ok {
		logger.Error().Msg("flow not found")
		return "", fmt.Errorf("flow not found")
	}

	stepHandler, ok := flowHandlers[currentStep]
	if !ok {
		logger.Error().Msg("step handler not found")
		return "", fmt.Errorf("step handler not found")
	}

	nextStep, err := stepHandler(ctx, flowProcessingOptions{
		user:          user,
		stateMetaData: state.Metedata,
		message:       message,
	})
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Msg(err.Error())
			return nextStep, err
		}

		logger.Error().Err(err).Msg("handle flow step")
		return "", fmt.Errorf("handle %s flow step: %w", currentStep, err)
	}

	return nextStep, nil
}

func getStateFromContext(ctx context.Context) (*models.State, error) {
	state, ok := ctx.Value(contextFieldNameState).(*models.State)
	if !ok {
		return nil, fmt.Errorf("state not found in context")
	}

	return state, nil
}

type updateStateOptions struct {
	updatedStep     models.FlowStep
	updatedMetadata map[string]any
	initialState    *models.State
}

func (h handlerService) updateState(ctx context.Context, opts updateStateOptions) {
	logger := h.logger.With().Str("name", "handlerService.updateState").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if opts.updatedStep != "" {
		opts.initialState.Steps = append(opts.initialState.Steps, opts.updatedStep)
	}

	if len(opts.updatedMetadata) > 0 {
		opts.initialState.Metedata = opts.updatedMetadata
	}

	updatedState, err := h.stores.State.Update(ctx, opts.initialState)
	if err != nil {
		logger.Error().Err(err).Msg("update state in store")
		return
	}

	logger.Debug().Any("updatedState", updatedState).Msg("updated state in store")
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
