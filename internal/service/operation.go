package service

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

func (h *handlerService) handleCreateOperationsThroughOneTimeInputFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateOperationsThroughOneTimeInputFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: opts.user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list categories from store")
		return "", fmt.Errorf("list categories from store: %w", err)
	}
	if len(categories) == 0 {
		logger.Info().Msg("no categories found")
		return models.EndFlowStep, ErrCategoriesNotFound
	}

	prompt, err := models.BuildCreateOperationFromTextPrompt(opts.message.GetText(), categories)
	if err != nil {
		logger.Error().Err(err).Msg("build create operation from text prompt")
		return "", fmt.Errorf("build create operation from text prompt: %w", err)
	}

	response, err := h.apis.Prompter.Execute(ctx, prompt)
	if err != nil {
		logger.Error().Err(err).Msg("execute prompt through prompter")
		return "", fmt.Errorf("execute prompt through prompter: %w", err)
	}

	operationData, err := models.OperationDataFromPromptOutput(response)
	if err != nil {
		logger.Error().Err(err).Msg("parse operation data from prompt output")
		return "", fmt.Errorf("parse operation data from prompt output: %w", err)
	}
	logger.Debug().Any("operationData", operationData).Msg("parsed operation data from prompt output")

	parsedAmount, err := money.NewFromString(operationData.Amount)
	if err != nil {
		return "", ErrInvalidAmountFormat
	}

	var categoryTitle string
	for _, category := range categories {
		if category.ID == operationData.CategoryID {
			categoryTitle = category.Title
			break
		}
	}
	if categoryTitle == "" {
		return models.EndFlowStep, ErrCategoryNotFound
	}

	opts.stateMetaData[categoryTitleMetadataKey] = categoryTitle
	opts.stateMetaData[operationTypeMetadataKey] = operationData.Type
	opts.stateMetaData[operationAmountMetadataKey] = parsedAmount.StringFixed()
	opts.stateMetaData[operationDescriptionMetadataKey] = operationData.Description

	operationDetailsMessage := fmt.Sprintf(`Please confirm the following operation details:
Category: %s
Operation Type: %s
Amount: %s
Description: %s

Do you confirm this operation?`,
		categoryTitle, operationData.Type, parsedAmount.StringFixed(), operationData.Description,
	)

	return models.ConfirmOperationDetailsFlowStep, h.sendMessageWithConfirmationInlineKeyboard(opts.message.GetChatID(), operationDetailsMessage)
}

func (h *handlerService) handleConfirmOperationDetailsFlowStepForOneTimeInputOperationCreate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleConfirmOperationDetailsFlowStepForOneTimeInputOperationCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationDetailsConfirmed, err := strconv.ParseBool(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse confirmation flag")
		return "", fmt.Errorf("parse confirmation flag: %w", err)
	}
	if !operationDetailsConfirmed {
		return models.EndFlowStep, h.notifyCancellationAndShowMenu(opts.message.GetChatID())
	}

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose balance:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForOneTimeInputOperationCreate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForOneTimeInputOperationCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		return models.EndFlowStep, ErrBalanceNotFound
	}
	opts.stateMetaData[balanceNameMetadataKey] = balance.Name // and we'll avoid id

	parsedAmount, err := money.NewFromString(opts.stateMetaData[operationAmountMetadataKey].(string))
	if err != nil {
		logger.Error().Err(err).Msg("parse amount")
		return "", fmt.Errorf("parse amount: %w", err)
	}
	operationType := models.OperationType(opts.stateMetaData[operationTypeMetadataKey].(string))

	err = h.createSpendingOrIncomingOperation(ctx, createSpendingOrIncomingOperationOptions{
		user:            opts.user,
		metaData:        opts.stateMetaData,
		operationType:   operationType,
		operationAmount: parsedAmount,
	})
	if err != nil {
		logger.Error().Err(err).Msgf("create %s operation", operationType)
		return "", fmt.Errorf("create %s operation: %w", operationType, err)
	}

	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), "Operation successfully created!")
}

func (h handlerService) handleCreateOperationFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateOperationFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	chooseOperationTypeKeyboard := []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotCreateIncomingOperationCommand,
				},
				{
					Text: models.BotCreateSpendingOperationCommand,
				},
			},
		},
	}

	if len(opts.user.Balances) > 1 {
		chooseOperationTypeKeyboard[0].Buttons = append(chooseOperationTypeKeyboard[0].Buttons, InlineKeyboardButton{
			Text: models.BotCreateTransferOperationCommand,
		})
	}

	return models.ProcessOperationTypeFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose operation type:",
		InlineKeyboard: chooseOperationTypeKeyboard,
	})
}

func (h handlerService) handleProcessOperationTypeFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleProcessOperationTypeFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationType := models.OperationCommandToOperationType[opts.message.GetText()]
	opts.stateMetaData[operationTypeMetadataKey] = operationType

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

	return nextStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  message,
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForCreatingOperation(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForCreatingOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceNameMetadataKey] = opts.message.GetText()

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: opts.user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list categories from store")
		return "", fmt.Errorf("list categories from store: %w", err)
	}
	if len(categories) == 0 {
		logger.Info().Msg("no categories found")
		return "", ErrCategoriesNotFound
	}
	logger.Debug().Any("categories", categories).Msg("got categories from store")

	return models.ChooseCategoryFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose operation category:",
		Keyboard: getKeyboardRows(categories, 3, true),
	})
}

func (h handlerService) handleChooseBalanceFromFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFromFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceFromMetadataKey] = opts.message.GetText()

	userBalancesWithoutBalanceFrom := slices.DeleteFunc(opts.user.Balances, func(balance models.Balance) bool {
		return balance.Name == opts.message.GetText()
	})

	return models.ChooseBalanceToFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose balance to which transfer operation should be performed:",
		Keyboard: getKeyboardRows(userBalancesWithoutBalanceFrom, 3, true),
	})
}

func (h handlerService) handleChooseBalanceToFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceToFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceToMetadataKey] = opts.message.GetText()

	balanceFrom, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:            opts.stateMetaData[balanceFromMetadataKey].(string),
		PreloadCurrency: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return "", fmt.Errorf("get balance from store: %w", err)
	}
	if balanceFrom == nil {
		logger.Error().Msg("balance from not found")
		return "", fmt.Errorf("balance from not found")
	}

	balanceTo, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:            opts.message.GetText(),
		PreloadCurrency: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return "", fmt.Errorf("get balance from store: %w", err)
	}
	if balanceTo == nil {
		logger.Error().Msg("balance to not found")
		return "", fmt.Errorf("balance to not found")
	}

	if balanceFrom.GetCurrency().Code != balanceTo.GetCurrency().Code {
		parsedBalanceFromAmount, _ := money.NewFromString(balanceFrom.Amount)
		parsedBalanceFromAmount.Mul(money.NewFromInt(4))

		outputMessage := fmt.Sprintf(`⚠️ Different Currency Transfer ⚠️
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
			balanceFrom.GetCurrency().Symbol,
			balanceFrom.Amount,
			balanceFrom.GetCurrency().Symbol,
			balanceTo.Name,
			balanceTo.GetCurrency().Symbol,
			balanceFrom.GetCurrency().Symbol,
			balanceTo.GetCurrency().Symbol,
			balanceTo.GetCurrency().Symbol,
			balanceFrom.GetCurrency().Symbol,
			balanceFrom.GetCurrency().Symbol,
			balanceTo.GetCurrency().Symbol,
			balanceFrom.Amount,
			balanceFrom.GetCurrency().Symbol,
			parsedBalanceFromAmount.StringFixed(),
			balanceTo.GetCurrency().Symbol,
		)
		return models.EnterCurrencyExchangeRateFlowStep, h.showCancelButton(opts.message.GetChatID(), outputMessage)
	}

	return models.EnterOperationAmountFlowStep, h.showCancelButton(opts.message.GetChatID(), "Enter operation amount:")
}

func (h handlerService) handleEnterCurrencyExchangeRateFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterCurrencyExchangeRateFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	exchangeRate, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse exchange rate")
		return "", ErrInvalidExchangeRateFormat
	}

	opts.stateMetaData[exchangeRateMetadataKey] = exchangeRate.String()
	logger.Debug().Any("exchangeRate", exchangeRate).Msg("parsed exchange rate")

	balanceFrom, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:            opts.stateMetaData[balanceFromMetadataKey].(string),
		PreloadCurrency: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return "", fmt.Errorf("get balance form store: %w", err)
	}
	if balanceFrom == nil {
		logger.Error().Err(err).Msg("balance from not found")
		return "", fmt.Errorf("balance from not found")
	}

	return models.EnterOperationAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), fmt.Sprintf(
		"Enter operation amount(currency: %s): ",
		balanceFrom.GetCurrency().Name,
	))
}

func (h handlerService) handleChooseCategoryFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseCategoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[categoryTitleMetadataKey] = opts.message.GetText()
	return models.EnterOperationDescriptionFlowStep, h.showCancelButton(opts.message.GetChatID(), "Enter operation description:")
}

func (h handlerService) handleEnterOperationDescriptionFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationDescriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[operationDescriptionMetadataKey] = opts.message.GetText()
	return models.EnterOperationAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter operation amount:")
}

func (h handlerService) handleEnterOperationAmountFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationAmountFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationAmount, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation amount")
		return "", ErrInvalidAmountFormat
	}
	logger.Debug().Any("operationAmount", operationAmount).Msg("parsed operation amount")

	operationType := models.OperationType(opts.stateMetaData[operationTypeMetadataKey].(string))
	logger.Debug().Any("operationType", operationType).Msg("parsed operation type")

	switch operationType {
	case models.OperationTypeIncoming, models.OperationTypeSpending:
		err := h.createSpendingOrIncomingOperation(ctx, createSpendingOrIncomingOperationOptions{
			metaData:        opts.stateMetaData,
			user:            opts.user,
			operationAmount: operationAmount,
			operationType:   operationType,
		})
		if err != nil {
			logger.Error().Err(err).Msgf("create %s operation", operationType)
			return "", fmt.Errorf("process %s operation: %w", operationType, err)
		}
	case models.OperationTypeTransfer:
		err := h.createTransferOperation(ctx, createTransferOperationOptions{
			metaData:        opts.stateMetaData,
			user:            opts.user,
			operationAmount: operationAmount,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create transfer operation")
			return "", fmt.Errorf("create transfer operation: %w", err)
		}

	default:
		logger.Error().Any("operationType", operationType).Msg("invalid operation type")
		return "", fmt.Errorf("received unknown operation type: %s", operationType)
	}

	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), "Operation created!")
}

type createSpendingOrIncomingOperationOptions struct {
	metaData        map[string]any
	user            *models.User
	operationAmount money.Money
	operationType   models.OperationType
}

func (h handlerService) createSpendingOrIncomingOperation(ctx context.Context, opts createSpendingOrIncomingOperationOptions) error {
	logger := h.logger.With().Str("name", "handlerService.createSpendingOrIncomingOperation").Logger()
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

type createTransferOperationOptions struct {
	metaData        map[string]any
	user            *models.User
	operationAmount money.Money
}

func (h handlerService) createTransferOperation(ctx context.Context, opts createTransferOperationOptions) error {
	logger := h.logger.With().Str("name", "handlerService.createTransferOperation").Logger()
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

	for _, balance := range []*models.Balance{balanceFrom, balanceTo} {
		err = h.stores.Balance.Update(ctx, balance)
		if err != nil {
			logger.Error().Err(err).Msg("update balance in store")
			return fmt.Errorf("update balance in store: %w", err)
		}
	}

	return nil
}

func (h handlerService) handleGetOperationsHistoryFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleGetOperationsHistoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose balance to view operations history for:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForGetOperationsHistory(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForGetOperationsHistory").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceNameMetadataKey] = opts.message.GetText()

	return models.ChooseTimePeriodForOperationsHistoryFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:  opts.message.GetChatID(),
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
}

func (h handlerService) handleChooseTimePeriodForOperationsHistoryFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseTimePeriodForOperationsHistoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:            opts.stateMetaData[balanceNameMetadataKey].(string),
		PreloadCurrency: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return "", fmt.Errorf("get balance from store: %w", err)
	}
	if balance == nil {
		logger.Error().Msg("balance not found")
		return "", fmt.Errorf("balance not found")
	}
	logger.Debug().Any("balance", balance).Msg("got balance")

	creationPeriod := models.GetCreationPeriodFromText(opts.message.GetText())
	operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
		BalanceID:      balance.ID,
		CreationPeriod: creationPeriod,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get all operations from store")
		return "", fmt.Errorf("get all operations from store: %w", err)
	}
	if operations == nil {
		logger.Info().Msg("operations not found")
		return models.EndFlowStep, ErrOperationsNotFound
	}

	outputMessage := fmt.Sprintf("Balance Amount: %v%s\nPeriod: %v\n", balance.Amount, balance.GetCurrency().Symbol, creationPeriod)

	for _, o := range operations {
		outputMessage += fmt.Sprintf(
			"\nOperation: %s\nDescription: %s\nCategory: %s\nAmount: %v%s\nCreation date: %v\n- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -",
			o.Type, o.Description, o.CategoryID, o.Amount, balance.GetCurrency().Symbol, o.CreatedAt.Format(time.ANSIC),
		)
	}

	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), outputMessage)
}

func (h handlerService) handleDeleteOperationFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleDeleteOperationFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose balance to delete operation from:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForDeleteOperation(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForDeleteOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceNameMetadataKey] = opts.message.GetText()

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return models.ChooseOperationToDeleteFlowStep, h.sendListOfOperationsWithAbilityToPaginate(ctx, sendListOfOperationsWithAbilityToPaginateOptions{
		balanceID:     opts.user.GetBalance(opts.message.GetText()).ID,
		chatID:        opts.message.GetChatID(),
		stateMetadata: opts.stateMetaData,
	})
}

func (h handlerService) handleChooseOperationToDeleteFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseOperationToDeleteFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if opts.message.GetText() == models.BotShowMoreCommand {
		return models.ChooseOperationToDeleteFlowStep, h.sendListOfOperationsWithAbilityToPaginate(ctx, sendListOfOperationsWithAbilityToPaginateOptions{
			balanceID:                      opts.user.GetBalance(opts.stateMetaData[balanceNameMetadataKey].(string)).ID,
			chatID:                         opts.message.GetChatID(),
			stateMetadata:                  opts.stateMetaData,
			includeLastShowedOperationDate: true,
		})
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID: opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return "", fmt.Errorf("get operation from store: %w", err)
	}
	if operation == nil {
		logger.Info().Msg("operation not found")
		return "", ErrOperationNotFound
	}

	opts.stateMetaData[operationIDMetadataKey] = operation.ID

	return models.ConfirmOperationDeletionFlowStep, h.sendMessageWithConfirmationInlineKeyboard(
		opts.message.GetChatID(),
		operation.GetDeletionMessage(),
	)
}

func (h handlerService) handleConfirmOperationDeletionFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleConfirmOperationDeletionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	confirmOperationDeletion, err := strconv.ParseBool(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse callback data to bool")
		return "", fmt.Errorf("parse callback data to bool: %w", err)
	}

	if !confirmOperationDeletion {
		logger.Info().Msg("user did not confirm balance deletion")
		return models.EndFlowStep, h.notifyCancellationAndShowMenu(opts.message.GetChatID())
	}

	err = h.deleteOperation(ctx, deleteOperationOptions{
		user:        opts.user,
		balanceName: opts.stateMetaData[balanceNameMetadataKey].(string),
		operationID: opts.stateMetaData[operationIDMetadataKey].(string),
	})
	if err != nil {
		logger.Error().Err(err).Msg("delete operation")
		return "", fmt.Errorf("delete operation: %w", err)
	}

	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), "Operation deleted!")
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
		err := h.deleteSpendingOrIncomeOperation(ctx, operation, balance)
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

	pairedTransferOperation, err := h.findPairedTransferOperation(ctx, user, initialOperation)
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return fmt.Errorf("get operation from store: %w", err)
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

func (h handlerService) handleUpdateOperationFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateOperationFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose balance to update operation from:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForUpdateOperation(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForUpdateOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceNameMetadataKey] = opts.message.GetText()

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return models.ChooseOperationToUpdateFlowStep, h.sendListOfOperationsWithAbilityToPaginate(ctx, sendListOfOperationsWithAbilityToPaginateOptions{
		balanceID:     opts.user.GetBalance(opts.message.GetText()).ID,
		chatID:        opts.message.GetChatID(),
		stateMetadata: opts.stateMetaData,
	})
}

func (h handlerService) handleChooseOperationToUpdateFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseOperationToUpdateFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if opts.message.GetText() == models.BotShowMoreCommand {
		return models.ChooseOperationToUpdateFlowStep, h.sendListOfOperationsWithAbilityToPaginate(ctx, sendListOfOperationsWithAbilityToPaginateOptions{
			balanceID:                      opts.user.GetBalance(opts.stateMetaData[balanceNameMetadataKey].(string)).ID,
			chatID:                         opts.message.GetChatID(),
			stateMetadata:                  opts.stateMetaData,
			includeLastShowedOperationDate: true,
		})
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         opts.message.GetText(),
		BalanceIDs: opts.user.GetBalancesIDs(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return "", fmt.Errorf("get operation from store: %w", err)
	}
	if operation == nil {
		logger.Info().Msg("operation not found")
		return "", ErrOperationNotFound
	}

	opts.stateMetaData[operationIDMetadataKey] = operation.ID

	err = h.showCancelButton(opts.message.GetChatID(), operation.GetDetails())
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	var updateOperationOptionsKeyboard []InlineKeyboardRow
	switch operation.Type {
	case models.OperationTypeTransferIn, models.OperationTypeTransferOut:
		updateOperationOptionsKeyboard = updateOperationOptionsKeyboardForTransferOperations
	default:
		updateOperationOptionsKeyboard = updateOperationOptionsKeyboardForIncomingAndSpendingOperations
	}

	return models.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose update operation option:",
		InlineKeyboard: updateOperationOptionsKeyboard,
	})
}

func (h handlerService) handleChooseUpdateOperationOptionFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseUpdateOperationOptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	switch opts.message.GetText() {
	case models.BotUpdateOperationAmountCommand:
		return models.EnterOperationAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter updated operation amount:")
	case models.BotUpdateOperationDescriptionCommand:
		return models.EnterOperationDescriptionFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter updated operation description:")
	case models.BotUpdateOperationCategoryCommand:
		categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
			UserID: opts.user.ID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("list categories from store")
			return "", fmt.Errorf("list categories from store: %w", err)
		}
		if len(categories) == 0 {
			logger.Info().Msg("no categories found")
			return "", ErrCategoriesNotFound
		}
		logger.Debug().Any("categories", categories).Msg("got categories from store")

		operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
			ID:         opts.stateMetaData[operationIDMetadataKey].(string),
			BalanceIDs: opts.user.GetBalancesIDs(),
		})
		if err != nil {
			logger.Error().Err(err).Msg("get operation from store")
			return "", fmt.Errorf("get operation from store: %w", err)
		}
		if operation == nil {
			logger.Info().Msg("operation not found")
			return "", ErrOperationNotFound
		}

		categoriesWithoutAlreadyUsedCategory := slices.DeleteFunc(categories, func(category models.Category) bool {
			return category.ID == operation.CategoryID
		})
		if len(categoriesWithoutAlreadyUsedCategory) == 0 {
			return "", ErrNotEnoughCategories
		}

		return models.ChooseCategoryFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   opts.message.GetChatID(),
			Message:  "Choose updated operation category:",
			Keyboard: getKeyboardRows(categoriesWithoutAlreadyUsedCategory, 3, true),
		})
	case models.BotUpdateOperationDateCommand:
		return models.EnterOperationDateFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter updated operation date:\nPlease use the following format: DD/MM/YYYY HH:MM. Example: 01/01/2025 12:00")
	default:
		return "", fmt.Errorf("received unknown update operation option: %s", opts.message.GetText())
	}
}

func (h handlerService) handleEnterOperationAmountFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationAmountFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationAmount, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation amount")
		return "", ErrInvalidAmountFormat
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         opts.stateMetaData[operationIDMetadataKey].(string),
		BalanceIDs: opts.user.GetBalancesIDs(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return "", fmt.Errorf("get operation from store: %w", err)
	}
	if operation == nil {
		logger.Info().Msg("operation not found")
		return "", ErrOperationNotFound
	}

	switch operation.Type {
	case models.OperationTypeSpending, models.OperationTypeIncoming:
		balance := opts.user.GetBalance(operation.BalanceID)
		if balance == nil {
			logger.Info().Msg("balance not found")
			return "", ErrBalanceNotFound
		}
		logger.Debug().Any("balance", balance).Msg("got balance")

		err := h.updateOperationAmountForSpendingOrIncomeOperation(ctx, balance, operation, operationAmount)
		if err != nil {
			logger.Error().Err(err).Msgf("update operation amount for %s", operation.Type)
			return "", fmt.Errorf("update operation amount for %s: %w", operation.Type, err)
		}

		return models.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:         opts.message.GetChatID(),
			Message:        "Operation amount successfully updated!\nPlease choose other update operation option or finish action by canceling it!",
			InlineKeyboard: updateOperationOptionsKeyboardForIncomingAndSpendingOperations,
		})

	case models.OperationTypeTransferIn, models.OperationTypeTransferOut:
		err := h.updateOperationAmountForTransferOperation(ctx, opts.user, operation, operationAmount)
		if err != nil {
			logger.Error().Err(err).Msg("update operation amount for transfer operation")
			return "", fmt.Errorf("update operation amount for transfer operation: %w", err)
		}

		return models.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:         opts.message.GetChatID(),
			Message:        "Operation amount successfully updated!\nPlease choose other update operation option or finish action by canceling it!",
			InlineKeyboard: updateOperationOptionsKeyboardForTransferOperations,
		})
	}

	return "", nil
}

func (h handlerService) updateOperationAmountForSpendingOrIncomeOperation(ctx context.Context, balance *models.Balance, operation *models.Operation, updatedOperationAmount money.Money) error {
	logger := h.logger.With().Str("name", "handlerService.updateOperationAmountForSpendingOrIncomeOperation").Logger()
	logger.Debug().Any("operation", operation).Any("updatedOperationAmount", updatedOperationAmount).Any("balance", balance).Msg("got args")

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
	case models.OperationTypeIncoming:
		balanceAmountWithoutInitialOperationAmount := balanceAmount.Sub(operationAmount)
		balanceAmountWithoutInitialOperationAmount.Inc(updatedOperationAmount)
		balance.Amount = balanceAmountWithoutInitialOperationAmount.StringFixed()
	case models.OperationTypeSpending:
		balanceAmount.Inc(operationAmount)
		balanceAmountWithUpdatedOperationAmount := balanceAmount.Sub(updatedOperationAmount)
		balance.Amount = balanceAmountWithUpdatedOperationAmount.StringFixed()
	}

	err = h.stores.Balance.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	operation.Amount = updatedOperationAmount.StringFixed()
	err = h.stores.Operation.Update(ctx, operation.ID, operation)
	if err != nil {
		logger.Error().Err(err).Msg("update operation in store")
		return fmt.Errorf("update operation in store: %w", err)
	}

	return nil
}

func (h handlerService) updateOperationAmountForTransferOperation(ctx context.Context, user *models.User, initialOperation *models.Operation, updatedOperationAmount money.Money) error {
	logger := h.logger.With().Str("name", "handlerService.updateOperationAmountForTransferOperation").Logger()
	logger.Debug().Any("operation", initialOperation).Any("user", user).Any("updatedOperationAmount", updatedOperationAmount).Msg("got args")

	pairedTransferOperation, err := h.findPairedTransferOperation(ctx, user, initialOperation)
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return fmt.Errorf("get operation from store: %w", err)
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
		initialBalanceWithoutOperationAmount := initialBalanceAmount.Sub(operationAmount)
		initialBalanceWithoutOperationAmount.Inc(updatedOperationAmount)
		initialBalance.Amount = initialBalanceWithoutOperationAmount.StringFixed()

		pairedBalanceAmount.Inc(operationAmount)
		pairedBalanceAmountWithUpdatedOperationAmount := pairedBalanceAmount.Sub(updatedOperationAmount)
		pairedBalance.Amount = pairedBalanceAmountWithUpdatedOperationAmount.StringFixed()
	case models.OperationTypeTransferOut:
		initialBalanceAmount.Inc(operationAmount)
		initialBalanceAmountWithUpdatedOperationAmount := initialBalanceAmount.Sub(updatedOperationAmount)
		initialBalance.Amount = initialBalanceAmountWithUpdatedOperationAmount.StringFixed()

		pairedBalanceWithoutOperationAmount := pairedBalanceAmount.Sub(operationAmount)
		pairedBalanceWithoutOperationAmount.Inc(updatedOperationAmount)
		pairedBalance.Amount = pairedBalanceWithoutOperationAmount.StringFixed()
	}

	for _, operation := range []*models.Operation{initialOperation, pairedTransferOperation} {
		operation.Amount = updatedOperationAmount.StringFixed()
		err = h.stores.Operation.Update(ctx, operation.ID, operation)
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

func (h handlerService) handleEnterOperationDescriptionFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationDescriptionFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         opts.stateMetaData[operationIDMetadataKey].(string),
		BalanceIDs: opts.user.GetBalancesIDs(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return "", fmt.Errorf("get operation from store: %w", err)
	}
	if operation == nil {
		logger.Info().Msg("operation not found")
		return "", ErrOperationNotFound
	}

	operation.Description = opts.message.GetText()

	err = h.stores.Operation.Update(ctx, operation.ID, operation)
	if err != nil {
		logger.Error().Err(err).Msg("update operation in store")
		return "", fmt.Errorf("update operation in store: %w", err)
	}

	return models.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Operation description successfully updated!\nPlease choose other update operation option or finish action by canceling it!",
		InlineKeyboard: updateOperationOptionsKeyboardForIncomingAndSpendingOperations,
	})
}

func (h handlerService) handleChooseCategoryFlowStepForOperationUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseCategoryFlowStepForOperationUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		UserID: opts.user.ID,
		Title:  opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return "", fmt.Errorf("get category from store: %w", err)
	}
	if category == nil {
		logger.Info().Msg("category not found")
		return "", ErrCategoryNotFound
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         opts.stateMetaData[operationIDMetadataKey].(string),
		BalanceIDs: opts.user.GetBalancesIDs(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return "", fmt.Errorf("get operation from store: %w", err)
	}
	if operation == nil {
		logger.Info().Msg("operation not found")
		return "", ErrOperationNotFound
	}

	operation.CategoryID = category.ID

	err = h.stores.Operation.Update(ctx, operation.ID, operation)
	if err != nil {
		logger.Error().Err(err).Msg("update operation in store")
		return "", fmt.Errorf("update operation in store: %w", err)
	}

	err = h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return models.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Operation category successfully updated!\nPlease choose other update operation option or finish action by canceling it!",
		InlineKeyboard: updateOperationOptionsKeyboardForIncomingAndSpendingOperations,
	})
}

func (h handlerService) handleEnterOperationDateFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationDateFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         opts.stateMetaData[operationIDMetadataKey].(string),
		BalanceIDs: opts.user.GetBalancesIDs(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return "", fmt.Errorf("get operation from store: %w", err)
	}
	if operation == nil {
		logger.Info().Msg("operation not found")
		return "", ErrOperationNotFound
	}

	parsedOperationDate, err := time.Parse(defaultTimeFormat, opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation date")
		return "", ErrInvalidDateFormat
	}

	var outputKeyboard []InlineKeyboardRow

	switch operation.Type {
	case models.OperationTypeSpending, models.OperationTypeIncoming:
		operation.CreatedAt = parsedOperationDate

		err = h.stores.Operation.Update(ctx, operation.ID, operation)
		if err != nil {
			logger.Error().Err(err).Msg("update operation in store")
			return "", fmt.Errorf("update operation in store: %w", err)
		}

		outputKeyboard = updateOperationOptionsKeyboardForIncomingAndSpendingOperations
	case models.OperationTypeTransferIn, models.OperationTypeTransferOut:
		pairedOperation, err := h.findPairedTransferOperation(ctx, opts.user, operation)
		if err != nil {
			logger.Error().Err(err).Msg("get operation from store")
			return "", fmt.Errorf("get operation from store: %w", err)
		}
		if pairedOperation == nil {
			logger.Info().Msg("paired operation not found")
			return "", ErrOperationNotFound
		}

		for _, operation := range []*models.Operation{operation, pairedOperation} {
			operation.CreatedAt = parsedOperationDate
			err = h.stores.Operation.Update(ctx, operation.ID, operation)
			if err != nil {
				logger.Error().Err(err).Msg("update operation in store")
				return "", fmt.Errorf("update operation in store: %w", err)
			}
		}
		outputKeyboard = updateOperationOptionsKeyboardForTransferOperations
	}

	return models.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Operation category successfully updated!\nPlease choose other update operation option or finish action by canceling it!",
		InlineKeyboard: outputKeyboard,
	})
}

func (h handlerService) findPairedTransferOperation(ctx context.Context, user *models.User, initialOperation *models.Operation) (*models.Operation, error) {
	logger := h.logger.With().Str("name", "handlerService.findPairedTransferOperation").Logger()
	logger.Debug().Any("initialOperation", initialOperation).Any("user", user).Msg("got args")

	filter := GetOperationFilter{
		Amount:       initialOperation.Amount,
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
		return nil, fmt.Errorf("get operation from store: %w", err)
	}
	if pairedTransferOperation == nil {
		logger.Info().Msg("paired transfer operation not found")
		return nil, fmt.Errorf("paired transfer operation not found")
	}

	return pairedTransferOperation, nil
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
		BalanceID: opts.balanceID,
	}

	if opts.includeLastShowedOperationDate {
		lastOperationTime, ok := opts.stateMetadata[lastOperationDateMetadataKey].(string)
		if ok {
			parsedTime, err := time.Parse(time.RFC3339Nano, lastOperationTime)
			if err != nil {
				logger.Error().Err(err).Msg("parse last operation date")
				return fmt.Errorf("parse last operation date: %w", err)
			}
			filter.CreatedAtLessThan = parsedTime
		}
	}

	operationsCount, err := h.stores.Operation.Count(ctx, filter)
	if err != nil {
		logger.Error().Err(err).Msg("count operations")
		return fmt.Errorf("count operations: %w", err)
	}
	if operationsCount == 0 {
		logger.Info().Any("balanceID", opts.balanceID).Msg("operations not found")
		return ErrOperationsNotFound
	}

	filter.SortByCreatedAtDesc = true
	filter.Limit = operationsPerMessage
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
		InlineKeyboard: convertModelToInlineKeyboardRowsWithPagination(operationsCount, operations, operationsPerMessage),
	})
	if err != nil {
		logger.Error().Err(err).Msg("create inline keyboard")
		return fmt.Errorf("create inline keyboard: %w", err)
	}

	return nil
}
