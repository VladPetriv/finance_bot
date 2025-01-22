package service

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h handlerService) HandleOperationCreate(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleOperationCreate").Logger()

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		if nextStep != "" {
			state.Steps = append(state.Steps, nextStep)
		}
		state.Metedata = stateMetaData
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Any("updatedState", updatedState).Msg("updated state in store")
	}()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username:        msg.GetSenderName(),
		PreloadBalances: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Info().Msg("user not found")
		return ErrUserNotFound
	}
	logger.Debug().Any("user", user).Msg("got user from store")

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on create operation flow")

	switch currentStep {
	case models.CreateOperationFlowStep:
		// Send an empty message with updated keyboard to avoid unexpected user behavior after clicking on previously generated keyboard.
		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  emptyMessage,
			Keyboard: rowKeyboardWithCancelButtonOnly,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard with welcome message")
			return fmt.Errorf("create keyboard with welcome message: %w", err)
		}
		err = h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose operation type:",
			InlineKeyboard: []InlineKeyboardRow{
				{
					Buttons: []InlineKeyboardButton{
						{
							Text: models.BotCreateIncomingOperationCommand,
						},
						{
							Text: models.BotCreateSpendingOperationCommand,
						},
						{
							Text: models.BotCreateTransferOperationCommand,
						},
					},
				},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("create inline keyboard")
			return fmt.Errorf("create inline keyboard: %w", err)
		}

		nextStep = models.ProcessOperationTypeFlowStep
	case models.ProcessOperationTypeFlowStep:
		step, err := h.handleProcessOperationTypeFlowStep(handleProcessOperationTypeFlowStepOptions{
			user:     user,
			metaData: stateMetaData,
			msg:      msg,
		})
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle process operation type flow step")
			return fmt.Errorf("handle process operation type flow step: %w", err)
		}

		nextStep = step
	case models.ChooseBalanceFlowStep:
		stateMetaData[balanceNameMetadataKey] = msg.GetText()
		categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
			UserID: user.ID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("list categories from store")
			return fmt.Errorf("list categories from store: %w", err)
		}
		if len(categories) == 0 {
			logger.Info().Msg("no categories found")
			return ErrCategoriesNotFound
		}
		logger.Debug().Any("categories", categories).Msg("got categories from store")

		err = h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Choose operation category:",
			Keyboard: getKeyboardRows(categories, 3, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseCategoryFlowStep
	case models.ChooseBalanceFromFlowStep:
		stateMetaData[balanceFromMetadataKey] = msg.GetText()

		userBalancesWithoutBalanceFrom := slices.DeleteFunc(user.Balances, func(balance models.Balance) bool {
			return balance.Name == msg.GetText()
		})

		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Choose balance to which transfer operation should be performed:",
			Keyboard: getKeyboardRows(userBalancesWithoutBalanceFrom, 3, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseBalanceToFlowStep
	case models.ChooseBalanceToFlowStep:
		stateMetaData[balanceToMetadataKey] = msg.GetText()

		balanceFrom := user.GetBalance(stateMetaData[balanceFromMetadataKey].(string))
		balanceTo := user.GetBalance(msg.GetText())

		if balanceFrom.Currency != balanceTo.Currency {
			parsedBalanceFromAmount, err := money.NewFromString(balanceFrom.Amount)
			if err != nil {
				logger.Error().Err(err).Msg("parse balance amount")
				return fmt.Errorf("parse balance amount: %w", err)
			}
			parsedBalanceFromAmount.Mul(money.NewFromInt(4))

			err = h.apis.Messenger.SendMessage(msg.GetChatID(), fmt.Sprintf(`⚠️ Different Currency Transfer ⚠️
Source Balance: %s
Currency: %s
Amount: %v %s

Destination Balance: %s
Currency: %s

To accurately convert your money, please provide the current exchange rate:

Formula: 1 %s = X %s
(How many %s you get for 1 %s)

Example:
- If 1 %s = 4 %s, enter: 4
- This means %v %s will be converted to %v %s

Please enter the current exchange rate:`,
				balanceFrom.Name,
				balanceFrom.Currency,
				balanceFrom.Amount,
				balanceFrom.Currency,
				balanceTo.Name,
				balanceTo.Currency,
				balanceFrom.Currency,
				balanceTo.Currency,
				balanceTo.Currency,
				balanceFrom.Currency,
				balanceFrom.Currency,
				balanceTo.Currency,
				balanceFrom.Amount,
				balanceFrom.Currency,
				parsedBalanceFromAmount.StringFixed(),
				balanceTo.Currency,
			))
			if err != nil {
				logger.Error().Err(err).Msg("send message")
				return fmt.Errorf("send message: %w", err)
			}

			nextStep = models.EnterCurrencyExchangeRateFlowStep
			break
		}

		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Enter operation amount:",
			Keyboard: rowKeyboardWithCancelButtonOnly,
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterOperationAmountFlowStep
	case models.EnterCurrencyExchangeRateFlowStep:
		exchangeRate, err := money.NewFromString(msg.GetText())
		if err != nil {
			logger.Error().Err(err).Msg("parse exchange rate")
			return ErrInvalidExchangeRateFormat
		}
		stateMetaData[exchangeRateMetadataKey] = exchangeRate.String()
		logger.Debug().Any("exchangeRate", exchangeRate).Msg("parsed exchange rate")

		err = h.apis.Messenger.SendMessage(msg.GetChatID(), fmt.Sprintf(
			"Enter operation amount(currency: %s): ",
			user.GetBalance(stateMetaData[balanceFromMetadataKey].(string)).Currency,
		))
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterOperationAmountFlowStep
	case models.ChooseCategoryFlowStep:
		stateMetaData[categoryTitleMetadataKey] = msg.GetText()

		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Enter operation description:",
			Keyboard: rowKeyboardWithCancelButtonOnly,
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterOperationDescriptionFlowStep
	case models.EnterOperationDescriptionFlowStep:
		stateMetaData[operationDescriptionMetadataKey] = msg.GetText()

		err := h.apis.Messenger.SendMessage(msg.GetChatID(), "Enter operation amount:")
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterOperationAmountFlowStep
	case models.EnterOperationAmountFlowStep:
		err := h.handleEnterOperationAmountFlowStep(ctx, handleEnterOperationAmountFlowStep{
			user:     user,
			metaData: stateMetaData,
			msg:      msg,
		})
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Msg(err.Error())
				return err
			}
			logger.Error().Err(err).Msg("handle enter operation amount flow step")
			return fmt.Errorf("handle enter operation amount flow step: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	return nil
}

type handleProcessOperationTypeFlowStepOptions struct {
	user     *models.User
	metaData map[string]any
	msg      Message
}

func (h handlerService) handleProcessOperationTypeFlowStep(opts handleProcessOperationTypeFlowStepOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleProcessOperationTypeFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationType := models.OperationCommandToOperationType[opts.msg.GetText()]
	opts.metaData[operationTypeMetadataKey] = operationType

	var (
		nextStep models.FlowStep
		message  string
	)

	switch operationType {
	case models.OperationTypeIncoming, models.OperationTypeSpending:
		message = "Choose balance:"
		nextStep = models.ChooseBalanceFlowStep
	case models.OperationTypeTransfer:
		message = "Choose balance from which money will be transferred:"
		nextStep = models.ChooseBalanceFromFlowStep
	}

	err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.msg.GetChatID(),
		Message:  message,
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
	if err != nil {
		logger.Error().Err(err).Msg("create row keyboard")
		return "", fmt.Errorf("create row keyboard: %w", err)
	}

	return nextStep, nil
}

type handleEnterOperationAmountFlowStep struct {
	user     *models.User
	metaData map[string]any
	msg      Message
}

func (h handlerService) handleEnterOperationAmountFlowStep(ctx context.Context, opts handleEnterOperationAmountFlowStep) error {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationAmountFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationAmount, err := money.NewFromString(opts.msg.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation amount")
		return ErrInvalidAmountFormat
	}
	logger.Debug().Any("operationAmount", operationAmount).Msg("parsed operation amount")

	operationType := models.OperationType(opts.metaData[operationTypeMetadataKey].(string))
	logger.Debug().Any("operationType", operationType).Msg("parsed operation type")

	switch operationType {
	case models.OperationTypeIncoming, models.OperationTypeSpending:
		err := h.processSpendingAndIncomingOperation(ctx, processSpendingAndIncomingOperationOptions{
			metaData:        opts.metaData,
			user:            opts.user,
			operationAmount: operationAmount,
			operationType:   operationType,
		})
		if err != nil {
			logger.Error().Err(err).Msg("process spending or incoming operation")
			return fmt.Errorf("process spending or incoming operation: %w", err)
		}

	case models.OperationTypeTransfer:
		err := h.processTransferOperation(ctx, processTransferOperationOptions{
			metaData:        opts.metaData,
			user:            opts.user,
			operationAmount: operationAmount,
		})
		if err != nil {
			logger.Error().Err(err).Msg("process transfer operation")
			return fmt.Errorf("process transfer operation: %w", err)
		}

	default:
		logger.Error().Any("operationType", operationType).Msg("invalid operation type")
		return fmt.Errorf("received unknown operation type: %s", operationType)
	}

	err = h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.msg.GetChatID(),
		Message:  "Operation created!",
		Keyboard: defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

type processSpendingAndIncomingOperationOptions struct {
	metaData        map[string]any
	user            *models.User
	operationAmount money.Money
	operationType   models.OperationType
}

func (h handlerService) processSpendingAndIncomingOperation(ctx context.Context, opts processSpendingAndIncomingOperationOptions) error {
	logger := h.logger.With().Str("name", "handlerService.processSpendingAndIncomingOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.metaData[balanceNameMetadataKey].(string))
	if balance == nil {
		logger.Info().Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balance", balance).Msg("got balance")

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		Title: opts.metaData[categoryTitleMetadataKey].(string),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return fmt.Errorf("get category from store: %w", err)
	}
	if category == nil {
		logger.Info().Msg("category not found")
		return ErrCategoryNotFound
	}
	logger.Debug().Any("category", category).Msg("got category from store")

	balanceAmount, err := money.NewFromString(balance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance amount")
		return fmt.Errorf("parse balance amount: %w", err)
	}
	logger.Debug().Any("balanceAmount", balanceAmount).Msg("parsed balance amount")

	switch opts.operationType {
	case models.OperationTypeIncoming:
		balanceAmount.Inc(opts.operationAmount)
		logger.Debug().Any("balanceAmount", balanceAmount).Msg("increased balance amount with incoming operation")
		balance.Amount = balanceAmount.StringFixed()

	case models.OperationTypeSpending:
		calculatedAmount := balanceAmount.Sub(opts.operationAmount)
		logger.Debug().Any("calculatedAmount", calculatedAmount).Msg("decreased balance amount with spending operation")
		balance.Amount = calculatedAmount.StringFixed()
	}

	operation := &models.Operation{
		ID:          uuid.NewString(),
		BalanceID:   balance.ID,
		CategoryID:  category.ID,
		Type:        opts.operationType,
		Amount:      opts.operationAmount.StringFixed(),
		Description: opts.metaData[operationDescriptionMetadataKey].(string),
		CreatedAt:   time.Now(),
	}
	logger.Debug().Any("operation", operation).Msg("build operation for create")

	err = h.stores.Operation.Create(ctx, operation)
	if err != nil {
		logger.Error().Err(err).Msg("create operation in store")
		return fmt.Errorf("create operation in store: %w", err)
	}

	err = h.stores.Balance.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	return nil
}

type processTransferOperationOptions struct {
	metaData        map[string]any
	user            *models.User
	operationAmount money.Money
}

func (h handlerService) processTransferOperation(ctx context.Context, opts processTransferOperationOptions) error {
	logger := h.logger.With().Str("name", "handlerService.processTransferOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceFrom := opts.user.GetBalance(opts.metaData[balanceFromMetadataKey].(string))
	if balanceFrom == nil {
		logger.Info().Msg("balance 'from' not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balanceFrom", balanceFrom).Msg("got balance from which money is transferred")

	balanceTo := opts.user.GetBalance(opts.metaData[balanceToMetadataKey].(string))
	if balanceTo == nil {
		logger.Info().Msg("balance 'to' not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balanceTo", balanceTo).Msg("got balance to which money is transferred")

	balanceFromAmount, err := money.NewFromString(balanceFrom.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance 'from' amount")
		return fmt.Errorf("parse balance 'from' amount: %w", err)
	}

	balanceToAmount, err := money.NewFromString(balanceTo.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance 'to' amount")
		return fmt.Errorf("parse balance 'to' amount: %w", err)
	}

	operationAmountIn, operationAmountOut := opts.operationAmount, opts.operationAmount

	exchangeRate, ok := opts.metaData[exchangeRateMetadataKey]
	if ok {
		parsedExchangeRate, err := money.NewFromString(exchangeRate.(string))
		if err != nil {
			logger.Error().Err(err).Msg("parse exchange rate")
			return fmt.Errorf("parse exchange rate: %w", err)
		}

		operationAmountIn.Mul(parsedExchangeRate)
		logger.Info().Any("operationAmountIn", operationAmountIn).Msg("updated operation amount, since exchange rate was provided")
	}

	calculatedAmount := balanceFromAmount.Sub(operationAmountOut)
	balanceFrom.Amount = calculatedAmount.StringFixed()

	balanceToAmount.Inc(operationAmountIn)
	balanceTo.Amount = balanceToAmount.StringFixed()

	operationsForCreate := []models.Operation{
		{
			ID:          uuid.NewString(),
			BalanceID:   balanceFrom.ID,
			CategoryID:  "",
			Type:        models.OperationTypeTransferOut,
			Amount:      operationAmountOut.StringFixed(),
			Description: fmt.Sprintf("Transfer: %s ➜ %s", balanceFrom.Name, balanceTo.Name),
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.NewString(),
			BalanceID:   balanceTo.ID,
			CategoryID:  "",
			Type:        models.OperationTypeTransferIn,
			Amount:      operationAmountIn.StringFixed(),
			Description: fmt.Sprintf("Received transfer from %s", balanceFrom.Name),
			CreatedAt:   time.Now(),
		},
	}
	for _, operation := range operationsForCreate {
		err := h.stores.Operation.Create(ctx, &operation)
		if err != nil {
			logger.Error().Err(err).Msg("create operation in store")
			return fmt.Errorf("create operation in store: %w", err)
		}
	}

	err = h.stores.Balance.Update(ctx, balanceTo)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	err = h.stores.Balance.Update(ctx, balanceFrom)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	return nil
}

func (h handlerService) HandleOperationHistory(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleOperationHistory").Logger()

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		if nextStep != "" {
			state.Steps = append(state.Steps, nextStep)
		}
		state.Metedata = stateMetaData
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Any("updatedState", updatedState).Msg("updated state in store")
	}()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username:        msg.GetSenderName(),
		PreloadBalances: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Info().Msg("user not found")
		return ErrUserNotFound
	}
	logger.Debug().Any("user", user).Msg("got user from store")

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on get operations history flow")

	switch currentStep {
	case models.GetOperationsHistoryFlowStep:
		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Choose balance to view operations history for:",
			Keyboard: getKeyboardRows(user.Balances, 3, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep
	case models.ChooseBalanceFlowStep:
		stateMetaData[balanceNameMetadataKey] = msg.GetText()

		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Please select a period for operation history!",
			Keyboard: []KeyboardRow{
				{
					Buttons: []string{
						string(models.CreationPeriodDay),
						string(models.CreationPeriodWeek),
						string(models.CreationPeriodMonth),
						string(models.CreationPeriodYear),
					},
				},
				{
					Buttons: []string{
						models.BotCancelCommand,
					},
				},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseTimePeriodForOperationsHistoryFlowStep
	case models.ChooseTimePeriodForOperationsHistoryFlowStep:
		err := h.handlerChooseTimePeriodForOperationsHistoryFlowStep(ctx, handlerChooseTimePeriodForOperationsHistoryFlowStepOptions{
			user:     user,
			metaData: stateMetaData,
			msg:      msg,
		})
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Msg(err.Error())
				return err
			}
			logger.Error().Err(err).Msg("handle choose time period for operations history flow step")
			return fmt.Errorf("handle choose time period for operations history flow step: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	return nil
}

type handlerChooseTimePeriodForOperationsHistoryFlowStepOptions struct {
	user     *models.User
	metaData map[string]any
	msg      Message
}

func (h handlerService) handlerChooseTimePeriodForOperationsHistoryFlowStep(ctx context.Context, opts handlerChooseTimePeriodForOperationsHistoryFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handlerChooseTimePeriodForOperationsHistoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.metaData[balanceNameMetadataKey].(string))
	if balance == nil {
		logger.Info().Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balance", balance).Msg("got balance")

	creationPeriod := models.GetCreationPeriodFromText(opts.msg.GetText())
	operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
		BalanceID:      balance.ID,
		CreationPeriod: creationPeriod,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get all operations from store")
		return fmt.Errorf("get all operations from store: %w", err)
	}
	if operations == nil {
		logger.Info().Msg("operations not found")
		return ErrOperationsNotFound
	}

	resultMessage := fmt.Sprintf("Balance Amount: %v%s\nPeriod: %v\n", balance.Amount, balance.Currency, *creationPeriod)

	for _, o := range operations {
		resultMessage += fmt.Sprintf(
			"\nOperation: %s\nDescription: %s\nCategory: %s\nAmount: %v%s\nCreation date: %v\n- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -",
			o.Type, o.Description, o.CategoryID, o.Amount, balance.Currency, o.CreatedAt.Format(time.ANSIC),
		)
	}

	err = h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.msg.GetChatID(),
		Message:  resultMessage,
		Keyboard: defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleOperationDelete(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleOperationDelete").Logger()

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		if nextStep != "" {
			state.Steps = append(state.Steps, nextStep)
		}
		state.Metedata = stateMetaData
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Any("updatedState", updatedState).Msg("updated state in store")
	}()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username:        msg.GetSenderName(),
		PreloadBalances: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Info().Msg("user not found")
		return ErrUserNotFound
	}
	logger.Debug().Any("user", user).Msg("got user from store")

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on create operation flow")

	switch currentStep {
	case models.DeleteOperationFlowStep:
		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Choose balance to delete operation from:",
			Keyboard: getKeyboardRows(user.Balances, 3, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep
	case models.ChooseBalanceFlowStep:
		stateMetaData[balanceNameMetadataKey] = msg.GetText()

		err := h.showCancelButton(msg.GetChatID())
		if err != nil {
			logger.Error().Err(err).Msg("show cancel button")
			return fmt.Errorf("show cancel button: %w", err)
		}

		err = h.sendListOfOperationsWithAbilityToPaginate(ctx, sendListOfOperationsWithAbilityToPaginateOptions{
			balanceID:     user.GetBalance(msg.GetText()).ID,
			chatID:        msg.GetChatID(),
			stateMetadata: stateMetaData,
		})
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Msg(err.Error())
				return err
			}
			logger.Error().Err(err).Msg("choose balance for delete operation flow step")
			return fmt.Errorf("choose balance for delete operation flow step: %w", err)
		}

		nextStep = models.ChooseOperationToDeleteFlowStep
	case models.ChooseOperationToDeleteFlowStep:
		step, err := h.chooseOperationToDeleteFlowStep(ctx, chooseOperationToDeleteFlowStepOptions{
			user:     user,
			msg:      msg,
			metaData: stateMetaData,
		})
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Msg(err.Error())
				return err
			}
			logger.Error().Err(err).Msg("handle choose operation to delete flow step")
			return fmt.Errorf("handle choose operation to delete flow step: %w", err)
		}

		nextStep = step

	case models.ConfirmOperationDeletionFlowStep:
		confirmOperationDeletion, err := strconv.ParseBool(msg.GetText())
		if err != nil {
			logger.Error().Err(err).Msg("parse callback data to bool")
			return fmt.Errorf("parse callback data to bool: %w", err)
		}

		if !confirmOperationDeletion {
			logger.Info().Msg("user did not confirm balance deletion")
			nextStep = models.EndFlowStep
			return h.notifyCancellationAndShowMenu(msg.GetChatID())
		}

		err = h.deleteOperation(ctx, deleteOperationOptions{
			user:        user,
			balanceName: stateMetaData[balanceNameMetadataKey].(string),
			operationID: stateMetaData[operationIDMetadataKey].(string),
		})
		if err != nil {
			logger.Error().Err(err).Msg("delete operation")
			return fmt.Errorf("delete operation: %w", err)
		}

		err = h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Operation deleted!",
			Keyboard: defaultKeyboardRows,
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	return nil
}

type chooseOperationToDeleteFlowStepOptions struct {
	user     *models.User
	msg      Message
	metaData map[string]any
}

func (h handlerService) chooseOperationToDeleteFlowStep(ctx context.Context, opts chooseOperationToDeleteFlowStepOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.chooseOperationToDeleteFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if opts.msg.GetText() == models.BotShowMoreOperationsForDeleteCommand {
		err := h.sendListOfOperationsWithAbilityToPaginate(ctx, sendListOfOperationsWithAbilityToPaginateOptions{
			balanceID:                      opts.user.GetBalance(opts.metaData[balanceNameMetadataKey].(string)).ID,
			chatID:                         opts.msg.GetChatID(),
			includeLastShowedOperationDate: true,
			stateMetadata:                  opts.metaData,
		})
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Msg(err.Error())
				return "", err
			}

			logger.Error().Err(err).Msg("send list of operations with ability to paginate")
			return "", fmt.Errorf("send list of operations with ability to paginate: %w", err)
		}

		return models.ChooseOperationToDeleteFlowStep, nil
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID: opts.msg.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return "", fmt.Errorf("get operation from store: %w", err)
	}
	if operation == nil {
		logger.Info().Msg("operation not found")
		return "", ErrOperationNotFound
	}

	opts.metaData[operationIDMetadataKey] = operation.ID

	err = h.sendMessageWithConfirmationInlineKeyboard(
		opts.msg.GetChatID(),
		operation.GetDeletionMessage(),
	)
	if err != nil {
		logger.Error().Err(err).Msg("send message with confirmation inline keyboard")
		return "", fmt.Errorf("send message with confirmation inline keyboard: %w", err)
	}

	return models.ConfirmOperationDeletionFlowStep, nil
}

type sendListOfOperationsWithAbilityToPaginateOptions struct {
	balanceID                      string
	chatID                         int
	includeLastShowedOperationDate bool
	stateMetadata                  map[string]any
}

const operationsPerMessage = 10

func (h handlerService) sendListOfOperationsWithAbilityToPaginate(ctx context.Context, opts sendListOfOperationsWithAbilityToPaginateOptions) error {
	logger := h.logger.With().Str("name", "handlerService.sendListOfOperationsWithAbilityToPaginate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	filter := ListOperationsFilter{
		BalanceID:           opts.balanceID,
		SortByCreatedAtDesc: true,
		Limit:               operationsPerMessage,
	}

	if opts.includeLastShowedOperationDate {
		lastOperationTime, ok := opts.stateMetadata[lastOperationDateMetadataKey].(primitive.DateTime)
		if ok {
			filter.CreatedAtLessThan = lastOperationTime.Time()
		}
	}

	operations, err := h.stores.Operation.List(ctx, filter)
	if err != nil {
		logger.Error().Err(err).Msg("list operations from store")
		return fmt.Errorf("list operations from store: %w", err)
	}
	if len(operations) == 0 {
		logger.Info().Any("balanceID", opts.balanceID).Msg("operations not found")
		return ErrOperationsNotFound
	}

	// Store the timestamp of the most recent operation in metadata.
	// This timestamp serves as a pagination cursor, enabling the retrieval
	// of subsequent operations in chronological order.
	opts.stateMetadata[lastOperationDateMetadataKey] = operations[len(operations)-1].CreatedAt

	err = h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.chatID,
		Message:        "Select operation to delete:",
		InlineKeyboard: convertOperationsToInlineKeyboardRowsWithPagination(operations, operationsPerMessage),
	})
	if err != nil {
		logger.Error().Err(err).Msg("create inline keyboard")
		return fmt.Errorf("create inline keyboard: %w", err)
	}

	return nil
}

type deleteOperationOptions struct {
	user        *models.User
	balanceName string
	operationID string
}

func (h handlerService) deleteOperation(ctx context.Context, opts deleteOperationOptions) error {
	logger := h.logger.With().Str("name", "handlerService.deleteOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID: opts.operationID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return fmt.Errorf("get operation from store: %w", err)
	}
	if operation == nil {
		logger.Info().Msg("operation not found")
		return ErrOperationNotFound
	}

	balance := opts.user.GetBalance(opts.balanceName)
	if balance == nil {
		logger.Error().Msg("balance not found")
		return ErrBalanceNotFound
	}

	switch operation.Type {
	case models.OperationTypeTransferIn, models.OperationTypeTransferOut:
		return h.deleteTransferOperation(ctx, operation, opts.user)
	case models.OperationTypeSpending, models.OperationTypeIncoming:

		err = h.deleteSpendingOrIncomeOperation(ctx, operation, balance)
		if err != nil {
			logger.Error().Err(err).Msgf("delete %s operation", operation.Type)
			return fmt.Errorf("delete %s operation: %w", operation.Type, err)
		}
	}

	return nil
}

// deleteTransferOperation handles the deletion of a transfer operation and its paired counterpart, adjusting the balances accordingly
func (h handlerService) deleteTransferOperation(ctx context.Context, initialOperation *models.Operation, user *models.User) error {
	logger := h.logger.With().Str("name", "handlerService.deleteTransferOperation").Logger()
	logger.Debug().Any("operation", initialOperation).Any("user", user).Msg("got args")

	filter := GetOperationFilter{
		BalanceIDs:   user.GetBalancesIDs(),
		CreateAtFrom: initialOperation.CreatedAt,
		CreateAtTo:   initialOperation.CreatedAt.Add(1 * time.Second),
	}

	// Determine the type of paired operation to look for
	switch initialOperation.Type {
	case models.OperationTypeTransferIn:
		filter.Type = models.OperationTypeTransferOut
	case models.OperationTypeTransferOut:
		filter.Type = models.OperationTypeTransferIn
	}

	pairedTransferOperation, err := h.stores.Operation.Get(ctx, filter)
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return fmt.Errorf("get operation from store: %w", err)
	}
	if pairedTransferOperation == nil {
		logger.Info().Msg("reversed transfer operation not found")
		return ErrOperationNotFound
	}

	pairedBalance := user.GetBalance(pairedTransferOperation.BalanceID)
	if pairedBalance == nil {
		logger.Info().Msg("paired balance not found")
		return ErrBalanceNotFound
	}

	initialBalance := user.GetBalance(initialOperation.BalanceID)
	if initialBalance == nil {
		logger.Info().Msg("initial balance not found")
		return ErrBalanceNotFound
	}

	operationAmount, err := money.NewFromString(initialOperation.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse operation amount")
		return fmt.Errorf("parse operation amount: %w", err)
	}

	initialBalanceAmount, err := money.NewFromString(initialBalance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance amount")
		return fmt.Errorf("parse balance amount: %w", err)
	}

	pairedBalanceAmount, err := money.NewFromString(pairedBalance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance amount")
		return fmt.Errorf("parse balance amount: %w", err)
	}

	switch initialOperation.Type {
	case models.OperationTypeTransferIn:
		initialBalance.Amount = initialBalanceAmount.Sub(operationAmount).StringFixed()
		pairedBalanceAmount.Inc(operationAmount)
		pairedBalance.Amount = pairedBalanceAmount.StringFixed()
	case models.OperationTypeTransferOut:
		initialBalanceAmount.Inc(operationAmount)
		initialBalance.Amount = initialBalanceAmount.StringFixed()
		pairedBalance.Amount = pairedBalanceAmount.Sub(operationAmount).StringFixed()
	}

	for _, operation := range []string{initialOperation.ID, pairedTransferOperation.ID} {
		err = h.stores.Operation.Delete(ctx, operation)
		if err != nil {
			logger.Error().Err(err).Msg("delete operation from store")
			return fmt.Errorf("delete operation from store: %w", err)
		}
	}

	for _, balance := range []*models.Balance{initialBalance, pairedBalance} {
		err = h.stores.Balance.Update(ctx, balance)
		if err != nil {
			logger.Error().Err(err).Msg("delete balance from store")
			return fmt.Errorf("delete balance from store: %w", err)
		}
	}

	return nil
}

// deleteSpendingOrIncomeOperation handles the deletion of a spending or income operation
// and updates the associated balance accordingly. For spending operations, it adds the
// amount back to the balance, and for income operations, it subtracts the amount from the balance.
func (h handlerService) deleteSpendingOrIncomeOperation(ctx context.Context, operation *models.Operation, balance *models.Balance) error {
	logger := h.logger.With().Str("name", "handlerService.deleteSpendingOrIncomeOperation").Logger()
	logger.Debug().Any("operation", operation).Any("balance", balance).Msg("got args")

	balanceAmount, err := money.NewFromString(balance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance amount")
		return fmt.Errorf("parse balance amount: %w", err)
	}

	operationAmount, err := money.NewFromString(operation.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse operation amount")
		return fmt.Errorf("parse operation amount: %w", err)
	}

	switch operation.Type {
	case models.OperationTypeSpending:
		balanceAmount.Inc(operationAmount)
		balance.Amount = balanceAmount.StringFixed()
	case models.OperationTypeIncoming:
		calculatedAmount := balanceAmount.Sub(operationAmount)
		balance.Amount = calculatedAmount.StringFixed()
	}

	err = h.stores.Balance.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	err = h.stores.Operation.Delete(ctx, operation.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete operation from store")
		return fmt.Errorf("delete operation from store: %w", err)
	}

	return nil
}
