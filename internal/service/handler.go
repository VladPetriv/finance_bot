package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/google/uuid"
)

type flowStepHandlerFunc func(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error)

type handlerService struct {
	logger   *logger.Logger
	services Services
	apis     APIs
	stores   Stores

	flowWithFlowStepsHandlers map[model.Flow]map[model.FlowStep]flowStepHandlerFunc
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
	h.flowWithFlowStepsHandlers = map[model.Flow]map[model.FlowStep]flowStepHandlerFunc{
		// Flows with balances
		model.StartFlow: {
			model.CreateInitialBalanceFlowStep: h.handleCreateInitialBalanceFlowStep,
			// NOTE: We're using -ForUpdate methods, since the balance after first step is already created in the store.
			model.EnterBalanceAmountFlowStep:   h.handleEnterBalanceAmountFlowStepForUpdate,
			model.EnterBalanceCurrencyFlowStep: h.handleEnterBalanceCurrencyFlowStepForUpdate,
		},
		model.CreateBalanceFlow: {
			model.CreateBalanceFlowStep:        h.handleCreateBalanceFlowStep,
			model.EnterBalanceNameFlowStep:     h.handleEnterBalanceNameFlowStepForCreate,
			model.EnterBalanceAmountFlowStep:   h.handleEnterBalanceAmountFlowStepForCreate,
			model.EnterBalanceCurrencyFlowStep: h.handleEnterBalanceCurrencyFlowStepForCreate,
		},
		model.GetBalanceFlow: {
			model.GetBalanceFlowStep:                   h.handleGetBalanceFlowStep,
			model.ChooseMonthBalanceStatisticsFlowStep: h.handleChooseMonthBalanceStatisticsFlowStep,
			model.ChooseBalanceFlowStep:                h.handleChooseBalanceFlowStepForGetBalance,
		},
		model.UpdateBalanceFlow: {
			model.UpdateBalanceFlowStep:             h.handleUpdateBalanceFlowStep,
			model.ChooseBalanceFlowStep:             h.handleChooseBalanceFlowStepForUpdate,
			model.ChooseUpdateBalanceOptionFlowStep: h.handleChooseUpdateBalanceOptionFlowStep,
			model.EnterBalanceNameFlowStep:          h.handleEnterBalanceNameFlowStepForUpdate,
			model.EnterBalanceAmountFlowStep:        h.handleEnterBalanceAmountFlowStepForUpdate,
			model.EnterBalanceCurrencyFlowStep:      h.handleEnterBalanceCurrencyFlowStepForUpdate,
		},
		model.DeleteBalanceFlow: {
			model.DeleteBalanceFlowStep:          h.handleDeleteBalanceFlowStep,
			model.ConfirmBalanceDeletionFlowStep: h.handleConfirmBalanceDeletionFlowStep,
			model.ChooseBalanceFlowStep:          h.handleChooseBalanceFlowStepForDelete,
		},

		// Flows with categories
		model.CreateCategoryFlow: {
			model.CreateCategoryFlowStep:    h.handleCreateCategoryFlowStep,
			model.EnterCategoryNameFlowStep: h.handleEnterCategoryNameFlowStep,
		},
		model.ListCategoriesFlow: {
			model.ListCategoriesFlowStep: h.handleListCategoriesFlowStep,
		},
		model.UpdateCategoryFlow: {
			model.UpdateCategoryFlowStep:           h.handleUpdateCategoryFlowStep,
			model.ChooseCategoryFlowStep:           h.handleChooseCategoryFlowStepForUpdate,
			model.EnterUpdatedCategoryNameFlowStep: h.handleEnterUpdatedCategoryNameFlowStep,
		},
		model.DeleteCategoryFlow: {
			model.DeleteCategoryFlowStep: h.handleDeleteCategoryFlowStep,
			model.ChooseCategoryFlowStep: h.handleChooseCategoryFlowStepForDelete,
		},

		// Flows with operations
		model.CreateOperationFlow: {
			model.CreateOperationFlowStep:           h.handleCreateOperationFlowStep,
			model.ProcessOperationTypeFlowStep:      h.handleProcessOperationTypeFlowStep,
			model.ChooseBalanceFlowStep:             h.handleChooseBalanceFlowStepForCreatingOperation,
			model.ChooseBalanceFromFlowStep:         h.handleChooseBalanceFromFlowStep,
			model.ChooseBalanceToFlowStep:           h.handleChooseBalanceToFlowStep,
			model.EnterCurrencyExchangeRateFlowStep: h.handleEnterCurrencyExchangeRateFlowStep,
			model.ChooseCategoryFlowStep:            h.handleChooseCategoryFlowStep,
			model.EnterOperationDescriptionFlowStep: h.handleEnterOperationDescriptionFlowStep,
			model.EnterOperationAmountFlowStep:      h.handleEnterOperationAmountFlowStep,
		},
		model.GetOperationsHistoryFlow: {
			model.GetOperationsHistoryFlowStep:                 h.handleGetOperationsHistoryFlowStep,
			model.ChooseBalanceFlowStep:                        h.handleChooseBalanceFlowStepForGetOperationsHistory,
			model.ChooseTimePeriodForOperationsHistoryFlowStep: h.handleChooseTimePeriodForOperationsHistoryFlowStep,
		},
		model.UpdateOperationFlow: {
			model.UpdateOperationFlowStep:             h.handleUpdateOperationFlowStep,
			model.ChooseBalanceFlowStep:               h.handleChooseBalanceFlowStepForUpdateOperation,
			model.ChooseOperationToUpdateFlowStep:     h.handleChooseOperationToUpdateFlowStep,
			model.ChooseUpdateOperationOptionFlowStep: h.handleChooseUpdateOperationOptionFlowStep,
			model.EnterOperationAmountFlowStep:        h.handleEnterOperationAmountFlowStepForUpdate,
			model.EnterOperationDescriptionFlowStep:   h.handleEnterOperationDescriptionFlowStepForUpdate,
			model.ChooseCategoryFlowStep:              h.handleChooseCategoryFlowStepForOperationUpdate,
			model.EnterOperationDateFlowStep:          h.handleEnterOperationDateFlowStep,
		},
		model.DeleteOperationFlow: {
			model.DeleteOperationFlowStep:          h.handleDeleteOperationFlowStep,
			model.ChooseBalanceFlowStep:            h.handleChooseBalanceFlowStepForDeleteOperation,
			model.ChooseOperationToDeleteFlowStep:  h.handleChooseOperationToDeleteFlowStep,
			model.ConfirmOperationDeletionFlowStep: h.handleConfirmOperationDeletionFlowStep,
		},
		model.CreateOperationsThroughOneTimeInputFlow: {
			model.CreateOperationsThroughOneTimeInputFlowStep: h.handleCreateOperationsThroughOneTimeInputFlowStep,
			model.ChooseBalanceFlowStep:                       h.handleChooseBalanceFlowStepForOneTimeInputOperationCreate,
			model.ConfirmOperationDetailsFlowStep:             h.handleConfirmOperationDetailsFlowStepForOneTimeInputOperationCreate,
		},

		// Flows with balance subscriptions
		model.CreateBalanceSubscriptionFlow: {
			model.CreateBalanceSubscriptionFlowStep:              h.handleCreateBalanceSubscriptionFlowStep,
			model.ChooseBalanceFlowStep:                          h.handleChooseBalanceFlowStepForCreateBalanceSubscription,
			model.ChooseCategoryFlowStep:                         h.handleChooseCategoryFlowStepForCreateBalanceSubscription,
			model.EnterBalanceSubscriptionNameFlowStep:           h.handleEnterBalanceSubscriptionNameFlowStep,
			model.EnterBalanceSubscriptionAmountFlowStep:         h.handleEnterBalanceSubscriptionAmountFlowStep,
			model.ChooseBalanceSubscriptionFrequencyFlowStep:     h.handleChooseBalanceSubscriptionFrequencyFlowStep,
			model.EnterStartAtDateForBalanceSubscriptionFlowStep: h.handleEnterStartAtDateForBalanceSubscriptionFlowStep,
		},
		model.ListBalanceSubscriptionFlow: {
			model.ListBalanceSubscriptionFlowStep: h.handleListBalanceSubscriptionFlowStep,
			model.ChooseBalanceFlowStep:           h.handleChooseBalanceFlowStepForListBalanceSubscriptions,
		},
		model.UpdateBalanceSubscriptionFlow: {
			model.UpdateBalanceSubscriptionFlowStep:             h.handleUpdateBalanceSubscriptionFlowStep,
			model.ChooseBalanceFlowStep:                         h.handleChooseBalanceFlowStepForUpdateBalanceSubscription,
			model.ChooseBalanceSubscriptionToUpdateFlowStep:     h.handleChooseBalanceSubscriptionToUpdateFlowStep,
			model.ChooseUpdateBalanceSubscriptionOptionFlowStep: h.handleChooseUpdateBalanceSubscriptionOptionFlowStep,
			model.EnterBalanceSubscriptionNameFlowStep:          h.handleEnterBalanceSubscriptionNameFlowStepForUpdate,
			model.EnterBalanceSubscriptionAmountFlowStep:        h.handleEnterBalanceSubscriptionAmountFlowStepForUpdate,
			model.ChooseCategoryFlowStep:                        h.handleChooseCategoryFlowStepForBalanceSubscriptionUpdate,
			model.ChooseBalanceSubscriptionFrequencyFlowStep:    h.handleChooseBalanceSubscriptionFrequencyFlowStepForUpdate,
		},
		model.DeleteBalanceSubscriptionFlow: {
			model.DeleteBalanceSubscriptionFlowStep:         h.handleDeleteBalanceSubscriptionFlowStep,
			model.ChooseBalanceFlowStep:                     h.handleChooseBalanceFlowStepForBalanceSubscriptionDelete,
			model.ChooseBalanceSubscriptionToDeleteFlowStep: h.handleChooseBalanceSubscriptionToDeleteFlowStep,
			model.ConfirmDeleteBalanceSubscriptionFlowStep:  h.handleConfirmDeleteBalanceSubscriptionFlowStep,
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

	var nextStep model.FlowStep
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
		nextStep = model.EndFlowStep
		return h.sendMessageWithDefaultKeyboard(chatID, fmt.Sprintf("Happy to see you again @%s!", username))
	}

	userID := uuid.NewString()
	err = h.stores.User.Create(ctx, &model.User{
		ID:       userID,
		ChatID:   chatID,
		Username: username,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create user in store")
		return fmt.Errorf("create user in store: %w", err)
	}

	err = h.stores.User.CreateSettings(ctx, &model.UserSettings{
		ID:                              uuid.NewString(),
		UserID:                          userID,
		AIParserEnabled:                 false,
		NotifyAboutSubscriptionPayments: true,
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

	nextStep = model.CreateInitialBalanceFlowStep
	return nil
}

func (h handlerService) HandleCancel(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleCancel").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

	state, err := getStateFromContext(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get state from context")
		return fmt.Errorf("get state from context: %w", err)
	}

	previousBaseFlow, ok := state.Metedata[baseFlowKey].(model.Flow)
	if !ok {
		logger.Warn().Msg("no base flow found in metadata, showing default menu")
		return h.sendMessageWithDefaultKeyboard(msg.GetChatID(), "Please choose command to execute:")
	}

	flowConfigs := map[model.Flow]struct {
		rows    []KeyboardRow
		message string
	}{
		model.BalanceFlow: {
			rows:    balanceKeyboardRows,
			message: "Action cancelled!\nPlease choose balance command to execute:",
		},
		model.CategoryFlow: {
			rows:    categoryKeyboardRows,
			message: "Action cancelled!\nPlease choose category command to execute:",
		},
		model.OperationFlow: {
			rows:    operationKeyboardRows,
			message: "Action cancelled!\nPlease choose operation command to execute:",
		},
		model.BalanceSubscriptionFlow: {
			rows:    balanceSubscriptionKeyboardRows,
			message: "Action cancelled!\nPlease choose balance subscription command to execute:",
		},
	}

	config, exists := flowConfigs[previousBaseFlow]
	if !exists {
		logger.Error().Str("flow", string(previousBaseFlow)).Msg("unknown flow")
		return fmt.Errorf("unknown flow: %s", previousBaseFlow)
	}

	return h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   msg.GetChatID(),
		Message:  config.message,
		Keyboard: config.rows,
	})
}

func (h handlerService) HandleBack(ctx context.Context, msg Message) error {
	return h.sendMessageWithDefaultKeyboard(msg.GetChatID(), "Please choose command to execute:")
}

func (h handlerService) HandleWrappers(ctx context.Context, event model.Event, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleWrappers").Logger()
	logger.Debug().Any("msg", msg).Any("event", event).Msg("got args")

	var (
		rows    []KeyboardRow
		message string
	)

	switch event {
	case model.BalanceEvent:
		rows = balanceKeyboardRows
		message = "Please choose balance command to execute:"
	case model.CategoryEvent:
		rows = categoryKeyboardRows
		message = "Please choose category command to execute:"
	case model.OperationEvent:
		rows = operationKeyboardRows
		message = "Please choose operation command to execute:"
	case model.BalanceSubscriptionEvent:
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

	var nextStep model.FlowStep
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

func (h handlerService) processHandler(ctx context.Context, state *model.State, message Message) (model.FlowStep, error) {
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

func getStateFromContext(ctx context.Context) (*model.State, error) {
	state, ok := ctx.Value(contextFieldNameState).(*model.State)
	if !ok {
		return nil, fmt.Errorf("state not found in context")
	}

	return state, nil
}

type updateStateOptions struct {
	updatedStep     model.FlowStep
	updatedMetadata map[string]any
	initialState    *model.State
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

// notifyCancellationAndShowKeyboard sends a cancellation message and displays the provided menu.
// It informs the user that their current action was cancelled and presents available commands
func (h handlerService) notifyCancellationAndShowKeyboard(message Message, keyboardRows []KeyboardRow) error {
	return h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          message.GetChatID(),
		MessageID:       message.GetMessageID(),
		UpdatedMessage:  "Action cancelled!\nPlease choose new command to execute:",
		UpdatedKeyboard: keyboardRows,
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

// sendMessageWithDefaultKeyboard sends a message to the specified chat with the default keyboard interface.
// If message is empty, it sends an empty message.
func (h handlerService) sendMessageWithDefaultKeyboard(chatID int, message string) error {
	if message == "" {
		message = emptyMessage
	}

	return h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   chatID,
		Message:  message,
		Keyboard: defaultKeyboardRows,
	})
}
