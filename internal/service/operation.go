package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

func (h handlerService) HandleEventOperationCreated(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventOperationCreated").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	logger.Debug().Any("stateMetaData", stateMetaData).Msg("got state metadata")

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
		Username:        msg.GetUsername(),
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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step")

	switch currentStep {
	case models.CreateOperationFlowStep:
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose operation type:",
			Type:    keyboardTypeInline,
			Rows: []bot.KeyboardRow{
				{
					Buttons: []string{models.BotCreateIncomingOperationCommand, models.BotCreateSpendingOperationCommand, models.BotCreateTransferOperationCommand},
				},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("create inline keyboard")
			return fmt.Errorf("create inline keyboard: %w", err)
		}

		nextStep = models.ProcessOprationTypeFlowStep
	case models.ProcessOprationTypeFlowStep:
		step, err := h.handleProcessOprationTypeFlowStep(handleProcessOprationTypeFlowStepOptions{
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
		stateMetaData["balanceName"] = msg.Message.Text
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

		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose operation category:",
			Type:    keyboardTypeRow,
			Rows:    getKeyboardRows(categories, false),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseCategoryFlowStep
	case models.ChooseBalanceFromFlowStep:
		stateMetaData["balanceFrom"] = msg.Message.Text

		userBalancesWithoutBalanceFrom := slices.DeleteFunc(user.Balances, func(balance models.Balance) bool {
			return balance.Name == msg.Message.Text
		})

		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose balance to which transfer operation should be performed:",
			Type:    keyboardTypeRow,
			Rows:    getKeyboardRows(userBalancesWithoutBalanceFrom, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseBalanceToFlowStep
	case models.ChooseBalanceToFlowStep:
		stateMetaData["balanceTo"] = msg.Message.Text

		balanceFrom := user.GetBalance(stateMetaData["balanceFrom"].(string))
		balanceTo := user.GetBalance(msg.Message.Text)

		if balanceFrom.Currency != balanceTo.Currency {
			parsedBalanceFromAmount, err := money.NewFromString(balanceFrom.Amount)
			if err != nil {
				logger.Error().Err(err).Msg("parse balance amount")
				return fmt.Errorf("parse balance amount: %w", err)
			}
			parsedBalanceFromAmount.Mul(money.NewFromInt(4))

			err = h.services.Message.SendMessage(&SendMessageOptions{
				ChatID: msg.GetChatID(),
				Text: fmt.Sprintf(`⚠️ Different Currency Transfer ⚠️

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
				),
			})
			if err != nil {
				logger.Error().Err(err).Msg("send message")
				return fmt.Errorf("send message: %w", err)
			}

			nextStep = models.EnterCurrencyExchangeRateFlowStep
			break
		}

		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: msg.GetChatID(),
			Text:   "Enter operation amount:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterOperationAmountFlowStep
	case models.EnterCurrencyExchangeRateFlowStep:
		exchangeRate, err := money.NewFromString(msg.Message.Text)
		if err != nil {
			logger.Error().Err(err).Msg("parse exchange rate")
			return ErrInvalidExchangeRateFormat
		}
		logger.Debug().Any("exchangeRate", exchangeRate).Msg("parsed exchange rate")

		stateMetaData["exchangeRate"] = exchangeRate.String()

		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: msg.GetChatID(),
			Text: fmt.Sprintf(
				"Enter operation amount(currency: %s): ",
				user.GetBalance(stateMetaData["balanceFrom"].(string)).Currency,
			),
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterOperationAmountFlowStep
	case models.ChooseCategoryFlowStep:
		stateMetaData["categoryTitle"] = msg.Message.Text

		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: msg.GetChatID(),
			Text:   "Enter operation description:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterOperationDescriptionFlowStep
	case models.EnterOperationDescriptionFlowStep:
		stateMetaData["operationDescription"] = msg.Message.Text

		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: msg.GetChatID(),
			Text:   "Enter operation amount:",
		})
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

	logger.Info().Msg("handled event operation created")
	return nil
}

type handleProcessOprationTypeFlowStepOptions struct {
	user     *models.User
	metaData map[string]any
	msg      botMessage
}

func (h handlerService) handleProcessOprationTypeFlowStep(opts handleProcessOprationTypeFlowStepOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleProcessOprationTypeFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationType := models.OperationCommandToOperationType[opts.msg.CallbackQuery.Data]
	opts.metaData["operationType"] = operationType

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

	err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  opts.msg.GetChatID(),
		Message: message,
		Type:    keyboardTypeRow,
		Rows:    getKeyboardRows(opts.user.Balances, true),
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
	msg      botMessage
}

func (h handlerService) handleEnterOperationAmountFlowStep(ctx context.Context, opts handleEnterOperationAmountFlowStep) error {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationAmountFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationAmount, err := money.NewFromString(opts.msg.Message.Text)
	if err != nil {
		logger.Error().Err(err).Msg("parse operation amount")
		return ErrInvalidAmountFormat
	}
	logger.Debug().Any("operationAmount", operationAmount).Msg("parsed operation amount")

	operationType := models.OperationType(opts.metaData["operationType"].(string))
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

	err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  opts.msg.GetChatID(),
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
		Message: "Operation created!",
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

	balance := opts.user.GetBalance(opts.metaData["balanceName"].(string))
	if balance == nil {
		logger.Info().Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balance", balance).Msg("got balance")

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		Title: opts.metaData["categoryTitle"].(string),
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
		Description: opts.metaData["operationDescription"].(string),
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

	balanceFrom := opts.user.GetBalance(opts.metaData["balanceFrom"].(string))
	if balanceFrom == nil {
		logger.Info().Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balanceFrom", balanceFrom).Msg("got balance from which money is transferred")

	balanceTo := opts.user.GetBalance(opts.metaData["balanceTo"].(string))
	if balanceTo == nil {
		logger.Info().Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balanceTo", balanceTo).Msg("got balance to which money is transferred")

	balanceFromAmount, err := money.NewFromString(balanceFrom.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance amount")
		return fmt.Errorf("parse balance amount: %w", err)
	}

	balanceToAmount, err := money.NewFromString(balanceTo.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance amount")
		return fmt.Errorf("parse balance amount: %w", err)
	}

	operationAmountIn, operationAmountOut := opts.operationAmount, opts.operationAmount

	exchangeRate, ok := opts.metaData["exchangeRate"]
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

func (h handlerService) HandleEventGetOperationsHistory(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventGetOperationsHistory").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	logger.Debug().Any("stateMetaData", stateMetaData).Msg("got state metadata")

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
		Username:        msg.GetUsername(),
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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step")

	switch currentStep {
	case models.GetOperationsHistoryFlowStep:
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose balance to view operations history for:",
			Type:    keyboardTypeRow,
			Rows:    getKeyboardRows(user.Balances, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep
	case models.ChooseBalanceFlowStep:
		stateMetaData["balanceName"] = msg.Message.Text
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Please select a period for operation history!",
			Type:    keyboardTypeRow,
			Rows: []bot.KeyboardRow{
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
						models.BotBackCommand,
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

	logger.Info().Msg("handled get operations history event")
	return nil
}

type handlerChooseTimePeriodForOperationsHistoryFlowStepOptions struct {
	user     *models.User
	metaData map[string]any
	msg      botMessage
}

func (h handlerService) handlerChooseTimePeriodForOperationsHistoryFlowStep(ctx context.Context, opts handlerChooseTimePeriodForOperationsHistoryFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handlerChooseTimePeriodForOperationsHistoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.metaData["balanceName"].(string))
	if balance == nil {
		logger.Info().Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balance", balance).Msg("got balance")

	creationPeriod := models.GetCreationPeriodFromText(opts.msg.Message.Text)
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

	err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  opts.msg.GetChatID(),
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
		Message: resultMessage,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
