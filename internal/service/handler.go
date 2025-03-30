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
	return &handlerService{
		logger:   opts.Logger,
		services: opts.Services,
		apis:     opts.APIs,
		stores:   opts.Stores,
	}
}

func (h *handlerService) RegisterHandlers() {
	h.flowWithFlowStepsHandlers = map[models.Flow]map[models.FlowStep]flowStepHandlerFunc{
		// Flows with balances
		models.StartFlow: {
			models.CreateInitialBalanceFlowStep: h.handleCreateInitialBalanceFlowStep,
			// NOTE: We're using -ForUpdate methods, since the balance after first step is already created in the store.
			models.EnterBalanceAmountFlowStep:   h.handleEnterBalanceAmountFlowStepForUpdate,
			models.EnterBalanceCurrencyFlowStep: h.handleEnterBalanceCurrencyFlowStepForUpdate,
		},
		models.CreateBalanceFlow: {
			models.CreateBalanceFlowStep:        h.handleCreateBalanceFlowStep,
			models.EnterBalanceNameFlowStep:     h.handleEnterBalanceNameFlowStepForCreate,
			models.EnterBalanceAmountFlowStep:   h.handleEnterBalanceAmountFlowStepForCreate,
			models.EnterBalanceCurrencyFlowStep: h.handleEnterBalanceCurrencyFlowStepForCreate,
		},
		models.GetBalanceFlow: {
			models.GetBalanceFlowStep:                   h.handleGetBalanceFlowStep,
			models.ChooseMonthBalanceStatisticsFlowStep: h.handleChooseMonthBalanceStatisticsFlowStep,
			models.ChooseBalanceFlowStep:                h.handleChooseBalanceFlowStepForGetBalance,
		},
		models.UpdateBalanceFlow: {
			models.UpdateBalanceFlowStep:             h.handleUpdateBalanceFlowStep,
			models.ChooseBalanceFlowStep:             h.handleChooseBalanceFlowStepForUpdate,
			models.ChooseUpdateBalanceOptionFlowStep: h.handleChooseUpdateBalanceOptionFlowStep,
			models.EnterBalanceNameFlowStep:          h.handleEnterBalanceNameFlowStepForUpdate,
			models.EnterBalanceAmountFlowStep:        h.handleEnterBalanceAmountFlowStepForUpdate,
			models.EnterBalanceCurrencyFlowStep:      h.handleEnterBalanceCurrencyFlowStepForUpdate,
		},
		models.DeleteBalanceFlow: {
			models.DeleteBalanceFlowStep:          h.handleDeleteBalanceFlowStep,
			models.ConfirmBalanceDeletionFlowStep: h.handleConfirmBalanceDeletionFlowStep,
			models.ChooseBalanceFlowStep:          h.handleChooseBalanceFlowStepForDelete,
		},

		// Flows with categories
		models.CreateCategoryFlow: {
			models.CreateCategoryFlowStep:    h.handleCreateCategoryFlowStep,
			models.EnterCategoryNameFlowStep: h.handleEnterCategoryNameFlowStep,
		},
		models.ListCategoriesFlow: {
			models.ListCategoriesFlowStep: h.handleListCategoriesFlowStep,
		},
		models.UpdateCategoryFlow: {
			models.UpdateCategoryFlowStep:           h.handleUpdateCategoryFlowStep,
			models.ChooseCategoryFlowStep:           h.handleChooseCategoryFlowStepForUpdate,
			models.EnterUpdatedCategoryNameFlowStep: h.handleEnterUpdatedCategoryNameFlowStep,
		},
		models.DeleteCategoryFlow: {
			models.DeleteCategoryFlowStep: h.handleDeleteCategoryFlowStep,
			models.ChooseCategoryFlowStep: h.handleChooseCategoryFlowStepForDelete,
		},

		// Flows with operations
		models.CreateOperationFlow: {
			models.CreateOperationFlowStep:           h.handleCreateOperationFlowStep,
			models.ProcessOperationTypeFlowStep:      h.handleProcessOperationTypeFlowStep,
			models.ChooseBalanceFlowStep:             h.handleChooseBalanceFlowStepForCreatingOperation,
			models.ChooseBalanceFromFlowStep:         h.handleChooseBalanceFromFlowStep,
			models.ChooseBalanceToFlowStep:           h.handleChooseBalanceToFlowStep,
			models.EnterCurrencyExchangeRateFlowStep: h.handleEnterCurrencyExchangeRateFlowStep,
			models.ChooseCategoryFlowStep:            h.handleChooseCategoryFlowStep,
			models.EnterOperationDescriptionFlowStep: h.handleEnterOperationDescriptionFlowStep,
			models.EnterOperationAmountFlowStep:      h.handleEnterOperationAmountFlowStep,
		},
		models.GetOperationsHistoryFlow: {
			models.GetOperationsHistoryFlowStep:                 h.handleGetOperationsHistoryFlowStep,
			models.ChooseBalanceFlowStep:                        h.handleChooseBalanceFlowStepForGetOperationsHistory,
			models.ChooseTimePeriodForOperationsHistoryFlowStep: h.handleChooseTimePeriodForOperationsHistoryFlowStep,
		},
		models.UpdateOperationFlow: {
			models.UpdateOperationFlowStep:             h.handleUpdateOperationFlowStep,
			models.ChooseBalanceFlowStep:               h.handleChooseBalanceFlowStepForUpdateOperation,
			models.ChooseOperationToUpdateFlowStep:     h.handleChooseOperationToUpdateFlowStep,
			models.ChooseUpdateOperationOptionFlowStep: h.handleChooseUpdateOperationOptionFlowStep,
			models.EnterOperationAmountFlowStep:        h.handleEnterOperationAmountFlowStepForUpdate,
			models.EnterOperationDescriptionFlowStep:   h.handleEnterOperationDescriptionFlowStepForUpdate,
			models.ChooseCategoryFlowStep:              h.handleChooseCategoryFlowStepForOperationUpdate,
			models.EnterOperationDateFlowStep:          h.handleEnterOperationDateFlowStep,
		},
		models.DeleteOperationFlow: {
			models.DeleteOperationFlowStep:          h.handleDeleteOperationFlowStep,
			models.ChooseBalanceFlowStep:            h.handleChooseBalanceFlowStepForDeleteOperation,
			models.ChooseOperationToDeleteFlowStep:  h.handleChooseOperationToDeleteFlowStep,
			models.ConfirmOperationDeletionFlowStep: h.handleConfirmOperationDeletionFlowStep,
		},
		models.CreateOperationsThroughOneTimeInputFlow: {
			models.CreateOperationsThroughOneTimeInputFlowStep: h.handleCreateOperationsThroughOneTimeInputFlowStep,
			models.ChooseBalanceFlowStep:                       h.handleChooseBalanceFlowStepForOneTimeInputOperationCreate,
			models.ConfirmOperationDetailsFlowStep:             h.handleConfirmOperationDetailsFlowStepForOneTimeInputOperationCreate,
		},

		// Flows with balance subscriptions
		models.CreateBalanceSubscriptionFlow: {
			models.CreateBalanceSubscriptionFlowStep:              h.handleCreateBalanceSubscriptionFlowStep,
			models.ChooseBalanceFlowStep:                          h.handleChooseBalanceFlowStepForCreateBalanceSubscription,
			models.ChooseCategoryFlowStep:                         h.handleChooseCategoryFlowStepForCreateBalanceSubscription,
			models.EnterBalanceSubscriptionNameFlowStep:           h.handleEnterBalanceSubscriptionNameFlowStep,
			models.EnterBalanceSubscriptionAmountFlowStep:         h.handleEnterBalanceSubscriptionAmountFlowStep,
			models.ChooseBalanceSubscriptionFrequencyFlowStep:     h.handleChooseBalanceSubscriptionFrequencyFlowStep,
			models.EnterStartAtDateForBalanceSubscriptionFlowStep: h.handleEnterStartAtDateForBalanceSubscriptionFlowStep,
		},
		models.ListBalanceSubscriptionsFlow: {
			models.ListBalanceSubscriptionsFlowStep: h.handleListBalanceSubscriptionsFlowStep,
			models.ChooseBalanceFlowStep:            h.handleChooseBalanceFlowStepForListBalanceSubscriptions,
		},
	}
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

	userID := uuid.NewString()
	err = h.stores.User.Create(ctx, &models.User{
		ID:       userID,
		Username: username,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create user in store")
		return fmt.Errorf("create user in store: %w", err)
	}

	err = h.stores.User.CreateSettings(ctx, &models.UserSettings{
		ID:              uuid.NewString(),
		UserID:          userID,
		AIParserEnabled: false,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create user settings in store")
		return fmt.Errorf("create user settings in store: %w", err)
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
	case models.BalanceSubscriptionsEvent:
		rows = balanceSubscriptionKeyboardRows
		message = "Please choose balance subscription command to execute:"
	default:
		return fmt.Errorf("unknown wrappers event: %s", event)
	}

	return h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   msg.GetChatID(),
		Message:  message,
		Keyboard: rows,
	})
}

func (h handlerService) HandleAction(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleAction").Logger()

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

func (h handlerService) processHandler(ctx context.Context, state *models.State, message Message) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.processHandler").Logger()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username:        message.GetSenderName(),
		PreloadBalances: true,
		PreloadSettings: true,
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
		logger.Error().Any("flow", state.Flow).Msg("flow not found")
		return "", fmt.Errorf("flow not found")
	}

	stepHandler, ok := flowHandlers[currentStep]
	if !ok {
		logger.Error().Msg("step handler not found")
		return "", fmt.Errorf("step handler not found")
	}

	nextStep, err := stepHandler(ctx, flowProcessingOptions{
		user:          user,
		message:       message,
		state:         state,
		stateMetaData: state.Metedata,
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

type identifiable interface {
	GetID() string
	GetName() string
}

// convertModelToInlineKeyboardRowsWithPagination converts a slice of model into inline keyboard rows with pagination support.
// If there are more models than the per-message limit, it adds a "Show More" button.
func convertModelToInlineKeyboardRowsWithPagination[T identifiable](actualCount int, data []T, limitPerMessage int) []InlineKeyboardRow {
	inlineKeyboardRows := make([]InlineKeyboardRow, 0, len(data))
	for _, value := range data {
		inlineKeyboardRows = append(inlineKeyboardRows, InlineKeyboardRow{
			Buttons: []InlineKeyboardButton{
				{
					Text: value.GetName(),
					Data: value.GetID(),
				},
			},
		})
	}

	// Skip BotShowMoreOperationsForDeleteCommand if operations are within the operationsPerMessage limit.
	if actualCount <= limitPerMessage {
		return inlineKeyboardRows
	}

	inlineKeyboardRows = append(inlineKeyboardRows, []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotShowMoreCommand,
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

const currenciesPerKeyboardRow = 3

func (h handlerService) getCurrenciesKeyboardForBalance(ctx context.Context) ([]InlineKeyboardRow, error) {
	logger := h.logger.With().Str("name", "handlerService.getCurrenciesKeyboardForBalance").Logger()

	currencies, err := h.stores.Currency.List(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("list currencies from store")
		return nil, fmt.Errorf("list currencies from store: %w", err)
	}
	if len(currencies) == 0 {
		logger.Info().Msg("currencies not found")
		return nil, fmt.Errorf("no currencies found")
	}

	currenciesKeyboard := make([]InlineKeyboardRow, 0)

	var currentRow InlineKeyboardRow
	for index, currency := range currencies {
		currentRow.Buttons = append(currentRow.Buttons, InlineKeyboardButton{
			Text: currency.Name,
			Data: currency.ID,
		})

		if len(currentRow.Buttons) == currenciesPerKeyboardRow || index == len(currencies)-1 {
			currenciesKeyboard = append(currenciesKeyboard, currentRow)
			currentRow = InlineKeyboardRow{}
		}
	}

	return currenciesKeyboard, nil
}
