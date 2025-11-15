package service

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

func (h *handlerService) handleCreateOperationsThroughOneTimeInputFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
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
		return model.EndFlowStep, ErrCategoriesNotFound
	}

	prompt, err := model.BuildCreateOperationFromTextPrompt(opts.message.GetText(), categories)
	if err != nil {
		logger.Error().Err(err).Msg("build create operation from text prompt")
		return "", fmt.Errorf("build create operation from text prompt: %w", err)
	}

	response, err := h.apis.Prompter.Execute(ctx, prompt)
	if err != nil {
		logger.Error().Err(err).Msg("execute prompt through prompter")
		return "", fmt.Errorf("execute prompt through prompter: %w", err)
	}

	operationData, err := model.OperationDataFromPromptOutput(response)
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
		return model.EndFlowStep, ErrCategoryNotFound
	}

	opts.stateMetaData.Add(model.CategoryTitleMetadataKey, categoryTitle)
	opts.stateMetaData.Add(model.OperationTypeMetadataKey, operationData.Type)
	opts.stateMetaData.Add(model.OperationAmountMetadataKey, parsedAmount.StringFixed())
	opts.stateMetaData.Add(model.OperationDescriptionMetadataKey, operationData.Description)

	operationDetailsMessage := fmt.Sprintf(`Please confirm the following operation details:
Category: %s
Operation Type: %s
Amount: %s
Description: %s

Do you confirm this operation?`,
		categoryTitle, operationData.Type, parsedAmount.StringFixed(), operationData.Description,
	)

	return model.ConfirmOperationDetailsFlowStep, h.sendMessageWithConfirmationInlineKeyboard(opts.message.GetChatID(), operationDetailsMessage)
}

func (h *handlerService) handleConfirmOperationDetailsFlowStepForOneTimeInputOperationCreate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleConfirmOperationDetailsFlowStepForOneTimeInputOperationCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationDetailsConfirmed, err := strconv.ParseBool(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse confirmation flag")
		return "", fmt.Errorf("parse confirmation flag: %w", err)
	}
	if !operationDetailsConfirmed {
		return model.EndFlowStep, h.notifyCancellationAndShowKeyboard(opts.message, defaultKeyboardRows)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose balance:",
		UpdatedInlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForOneTimeInputOperationCreate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForOneTimeInputOperationCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		return model.EndFlowStep, ErrBalanceNotFound
	}
	opts.stateMetaData.Add(model.BalanceNameMetadataKey, balance.Name)

	operationAmount, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationAmountMetadataKey)
	if !ok {
		logger.Error().Msg("operation amount not found in metadata")
		return "", fmt.Errorf("operation amount not found in metadata")
	}

	operationType, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationTypeMetadataKey)
	if !ok {
		logger.Error().Msg("operation type not found in metadata")
		return "", fmt.Errorf("operation type not found in metadata")
	}

	parsedAmount, err := money.NewFromString(operationAmount)
	if err != nil {
		logger.Error().Err(err).Msg("parse amount")
		return "", fmt.Errorf("parse amount: %w", err)
	}

	parsedOperationType := model.OperationType(operationType)

	operation, err := h.createSpendingOrIncomingOperation(ctx, createSpendingOrIncomingOperationOptions{
		user:            opts.user,
		metaData:        opts.stateMetaData,
		operationType:   parsedOperationType,
		operationAmount: parsedAmount,
	})
	if err != nil {
		logger.Error().Err(err).Msgf("create %s operation", operationType)
		return "", fmt.Errorf("create %s operation: %w", operationType, err)
	}

	return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		InlineMessageID: opts.message.GetInlineMessageID(),
		UpdatedKeyboard: defaultKeyboardRows,
		UpdatedMessage:  fmt.Sprintf("Operation successfully created!\n\n%s", operation.GetDetails()),
	})
}

func (h handlerService) handleCreateOperationFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
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
					Text: model.BotCreateIncomingOperationCommand,
				},
				{
					Text: model.BotCreateSpendingOperationCommand,
				},
			},
		},
	}

	if len(opts.user.Balances) > 1 {
		chooseOperationTypeKeyboard[0].Buttons = append(chooseOperationTypeKeyboard[0].Buttons, InlineKeyboardButton{
			Text: model.BotCreateTransferOperationCommand,
		})
	}

	return model.ProcessOperationTypeFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose operation type:",
		InlineKeyboard: chooseOperationTypeKeyboard,
	})
}

func (h handlerService) handleProcessOperationTypeFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleProcessOperationTypeFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationType := model.OperationCommandToOperationType[opts.message.GetText()]
	opts.stateMetaData.Add(model.OperationTypeMetadataKey, operationType)

	var (
		nextStep model.FlowStep
		message  string
	)
	switch operationType {
	case model.OperationTypeIncoming, model.OperationTypeSpending:
		message = "Choose balance:"
		nextStep = model.ChooseBalanceFlowStep
	case model.OperationTypeTransfer:
		message = "Choose balance *from which* money will be transferred:"
		nextStep = model.ChooseBalanceFromFlowStep
	}

	return nextStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage:          message,
		UpdatedInlineKeyboard:   getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForCreatingOperation(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForCreatingOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.BalanceNameMetadataKey, opts.message.GetText())

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

	return model.ChooseCategoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose operation category:",
		UpdatedInlineKeyboard: getInlineKeyboardRows(categories, 3),
	})
}

func (h handlerService) handleChooseBalanceFromFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFromFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.BalanceFromMetadataKey, opts.message.GetText())

	userBalancesWithoutBalanceFrom := slices.DeleteFunc(opts.user.Balances, func(balance model.Balance) bool {
		return balance.Name == opts.message.GetText()
	})

	return model.ChooseBalanceToFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage:          "Choose balance *to which* transfer operation should be performed:",
		UpdatedInlineKeyboard:   getInlineKeyboardRows(userBalancesWithoutBalanceFrom, 2),
	})
}

func (h handlerService) handleChooseBalanceToFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceToFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.BalanceToMetadataKey, opts.message.GetText())

	balanceFromName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceFromMetadataKey)
	if !ok {
		logger.Error().Msg("balance from not found in metadata")
		return "", fmt.Errorf("balance from not found in metadata")
	}

	balanceFrom, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:            balanceFromName,
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
		outputMessage := model.BuildCurrencyConversionMessage(balanceFrom, balanceTo)
		return model.EnterCurrencyExchangeRateFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:          opts.message.GetChatID(),
			MessageID:       opts.message.GetMessageID(),
			InlineMessageID: opts.message.GetInlineMessageID(),
			UpdatedMessage:  outputMessage,
		})
	}

	return model.EnterOperationAmountFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		InlineMessageID: opts.message.GetInlineMessageID(),
		UpdatedMessage:  "Enter operation amount:",
	})
}

func (h handlerService) handleEnterCurrencyExchangeRateFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterCurrencyExchangeRateFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	exchangeRate, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse exchange rate")
		return "", ErrInvalidExchangeRateFormat
	}

	opts.stateMetaData.Add(model.ExchangeRateMetadataKey, exchangeRate.String())
	logger.Debug().Any("exchangeRate", exchangeRate).Msg("parsed exchange rate")

	balanceFromName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceFromMetadataKey)
	if !ok {
		logger.Error().Msg("balance from not found in metadata")
		return "", fmt.Errorf("balance from not found in metadata")
	}

	balanceFrom, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:            balanceFromName,
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

	return model.EnterOperationAmountFlowStep, h.apis.Messenger.SendMessage(
		opts.message.GetChatID(),
		fmt.Sprintf(
			"Enter operation amount(currency: %s): ",
			balanceFrom.GetCurrency().Name,
		),
	)
}

func (h handlerService) handleChooseCategoryFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseCategoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.CategoryTitleMetadataKey, opts.message.GetText())
	return model.EnterOperationDescriptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		InlineMessageID: opts.message.GetInlineMessageID(),
		UpdatedMessage:  "Enter operation description:",
	})
}

func (h handlerService) handleEnterOperationDescriptionFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationDescriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.OperationDescriptionMetadataKey, opts.message.GetText())
	return model.EnterOperationAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter operation amount:")
}

func (h handlerService) handleEnterOperationAmountFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationAmountFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationAmount, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation amount")
		return "", ErrInvalidAmountFormat
	}
	logger.Debug().Any("operationAmount", operationAmount).Msg("parsed operation amount")

	operationType, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationTypeMetadataKey)
	if !ok {
		logger.Error().Msg("operation type not found in metadata")
		return "", fmt.Errorf("operation type not found in metadata")
	}

	parsedOperationType := model.OperationType(operationType)
	logger.Debug().Any("operationType", operationType).Msg("parsed operation type")

	var outputMessage string
	switch parsedOperationType {
	case model.OperationTypeIncoming, model.OperationTypeSpending:
		operation, err := h.createSpendingOrIncomingOperation(ctx, createSpendingOrIncomingOperationOptions{
			metaData:        opts.stateMetaData,
			user:            opts.user,
			operationAmount: operationAmount,
			operationType:   parsedOperationType,
		})
		if err != nil {
			logger.Error().Err(err).Msgf("create %s operation", operationType)
			return "", fmt.Errorf("process %s operation: %w", operationType, err)
		}

		outputMessage = fmt.Sprintf("Operation successfully created!\n\n%s", operation.GetDetails())

	case model.OperationTypeTransfer:
		operationOut, operationIn, err := h.createTransferOperation(ctx, createTransferOperationOptions{
			metaData:        opts.stateMetaData,
			user:            opts.user,
			operationAmount: operationAmount,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create transfer operation")
			return "", fmt.Errorf("create transfer operation: %w", err)
		}

		outputMessage = fmt.Sprintf("Operation successfully created!\n\nOperation Out:\n%s\n\nOperation In:\n%s", operationOut.GetDetails(), operationIn.GetDetails())

	default:
		logger.Error().Any("operationType", operationType).Msg("invalid operation type")
		return "", fmt.Errorf("received unknown operation type: %s", operationType)
	}

	return model.EndFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  outputMessage,
		Keyboard: operationKeyboardRows,
	})
}

type createSpendingOrIncomingOperationOptions struct {
	metaData        model.Metadata
	user            *model.User
	operationAmount money.Money
	operationType   model.OperationType
}

func (h handlerService) createSpendingOrIncomingOperation(ctx context.Context, opts createSpendingOrIncomingOperationOptions) (*model.Operation, error) {
	logger := h.logger.With().Str("name", "handlerService.createSpendingOrIncomingOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceName, ok := model.GetTypedFromMetadata[string](opts.metaData, model.BalanceNameMetadataKey)
	if !ok {
		logger.Error().Msg("balance name not found in metadata")
		return nil, fmt.Errorf("balance name not found in metadata")
	}

	categoryTitle, ok := model.GetTypedFromMetadata[string](opts.metaData, model.CategoryTitleMetadataKey)
	if !ok {
		logger.Error().Msg("category title not found in metadata")
		return nil, fmt.Errorf("category title not found in metadata")
	}

	operationDescription, ok := model.GetTypedFromMetadata[string](opts.metaData, model.OperationDescriptionMetadataKey)
	if !ok {
		logger.Error().Msg("operation description not found in metadata")
		return nil, fmt.Errorf("operation description not found in metadata")
	}

	balance := opts.user.GetBalance(balanceName)
	if balance == nil {
		logger.Info().Msg("balance not found")
		return nil, ErrBalanceNotFound
	}

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		Title: categoryTitle,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return nil, fmt.Errorf("get category from store: %w", err)
	}
	if category == nil {
		logger.Info().Msg("category not found")
		return nil, ErrCategoryNotFound
	}
	logger.Debug().Any("category", category).Msg("got category from store")

	balanceAmount, err := money.NewFromString(balance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance amount")
		return nil, fmt.Errorf("parse balance amount: %w", err)
	}
	logger.Debug().Any("balanceAmount", balanceAmount).Msg("parsed balance amount")

	switch opts.operationType {
	case model.OperationTypeIncoming:
		calculateIncomingOperation(&balanceAmount, opts.operationAmount)
	case model.OperationTypeSpending:
		calculateSpendingOperation(&balanceAmount, opts.operationAmount)
	}

	balance.Amount = balanceAmount.StringFixed()

	operation := &model.Operation{
		ID:          uuid.NewString(),
		BalanceID:   balance.ID,
		CategoryID:  category.ID,
		Type:        opts.operationType,
		Amount:      opts.operationAmount.StringFixed(),
		Description: operationDescription,
		CreatedAt:   time.Now(),
	}
	logger.Debug().Any("operation", operation).Msg("build operation for create")

	err = h.stores.Operation.Create(ctx, operation)
	if err != nil {
		logger.Error().Err(err).Msg("create operation in store")
		return nil, fmt.Errorf("create operation in store: %w", err)
	}

	err = h.stores.Balance.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return nil, fmt.Errorf("update balance in store: %w", err)
	}

	return operation, nil
}

type createTransferOperationOptions struct {
	metaData        model.Metadata
	user            *model.User
	operationAmount money.Money
}

func (h handlerService) createTransferOperation(ctx context.Context, opts createTransferOperationOptions) (*model.Operation, *model.Operation, error) {
	logger := h.logger.With().Str("name", "handlerService.createTransferOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceFromName, ok := model.GetTypedFromMetadata[string](opts.metaData, model.BalanceFromMetadataKey)
	if !ok {
		logger.Error().Msg("balance from not found in metadata")
		return nil, nil, fmt.Errorf("balance from not found in metadata")
	}

	balanceToName, ok := model.GetTypedFromMetadata[string](opts.metaData, model.BalanceToMetadataKey)
	if !ok {
		logger.Error().Msg("balance to not found in metadata")
		return nil, nil, fmt.Errorf("balance to not found in metadata")
	}

	balanceFrom := opts.user.GetBalance(balanceFromName)
	if balanceFrom == nil {
		logger.Info().Msg("balance 'from' not found")
		return nil, nil, ErrBalanceNotFound
	}
	logger.Debug().Any("balanceFrom", balanceFrom).Msg("got balance from which money is transferred")

	balanceTo := opts.user.GetBalance(balanceToName)
	if balanceTo == nil {
		logger.Info().Msg("balance 'to' not found")
		return nil, nil, ErrBalanceNotFound
	}
	logger.Debug().Any("balanceTo", balanceTo).Msg("got balance to which money is transferred")

	operationIDOut, operationIDIn := uuid.NewString(), uuid.NewString()
	operationOut := model.Operation{
		ID:                operationIDOut,
		BalanceID:         balanceFrom.ID,
		CategoryID:        "",
		ParentOperationID: operationIDIn,
		Type:              model.OperationTypeTransferOut,
		Amount:            opts.operationAmount.StringFixed(),
		Description:       fmt.Sprintf("Transfer: %s ➜ %s", balanceFrom.Name, balanceTo.Name),
		CreatedAt:         time.Now(),
	}
	operationIn := model.Operation{
		ID:                operationIDIn,
		BalanceID:         balanceTo.ID,
		CategoryID:        "",
		ParentOperationID: operationIDOut,
		Type:              model.OperationTypeTransferIn,
		Amount:            opts.operationAmount.StringFixed(),
		Description:       fmt.Sprintf("Received transfer from %s", balanceFrom.Name),
		CreatedAt:         time.Now(),
	}

	balanceAmountFrom, _ := money.NewFromString(balanceFrom.Amount)
	balanceAmountTo, _ := money.NewFromString(balanceTo.Amount)

	calculateOptions := calculateTransferOperationOptions{
		operationType:   operationIn.Type,
		balanceFrom:     &balanceAmountFrom,
		balanceTo:       &balanceAmountTo,
		operationAmount: opts.operationAmount,
	}

	exchangeRate, ok := model.GetTypedFromMetadata[string](opts.metaData, model.ExchangeRateMetadataKey)
	if ok {
		parsedExchangeRate, _ := money.NewFromString(exchangeRate)
		operationIn.ExchangeRate = parsedExchangeRate.String()
		operationOut.ExchangeRate = parsedExchangeRate.String()
		calculateOptions.exchangeRate = &parsedExchangeRate

		operationAmountIn := opts.operationAmount
		operationAmountIn.Mul(parsedExchangeRate)
		operationIn.Amount = operationAmountIn.StringFixed()
	}

	calculateTransferOperation(calculateOptions)

	for _, operation := range []model.Operation{operationIn, operationOut} {
		err := h.stores.Operation.Create(ctx, &operation)
		if err != nil {
			logger.Error().Err(err).Msg("create operation in store")
			return nil, nil, fmt.Errorf("create operation in store: %w", err)
		}
	}

	balanceFrom.Amount = balanceAmountFrom.StringFixed()
	balanceTo.Amount = balanceAmountTo.StringFixed()

	for _, balance := range []*model.Balance{balanceFrom, balanceTo} {
		err := h.stores.Balance.Update(ctx, balance)
		if err != nil {
			logger.Error().Err(err).Msg("update balance in store")
			return nil, nil, fmt.Errorf("update balance in store: %w", err)
		}
	}

	return &operationOut, &operationIn, nil
}

func (h handlerService) handleGetOperationsHistoryFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleGetOperationsHistoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose balance to view operations history for:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForGetOperationsHistory(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForGetOperationsHistory").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.BalanceNameMetadataKey, opts.message.GetText())
	opts.stateMetaData.Add(model.PageMetadataKey, firstPage)

	return model.ChooseTimePeriodForOperationsHistoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Please select a period for operation history!",
		UpdatedInlineKeyboard: operationHistoryPeriodKeyboard,
	})
}

func (h handlerService) handleChooseTimePeriodForOperationsHistoryFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseTimePeriodForOperationsHistoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceNameMetadataKey)
	if !ok {
		logger.Error().Msg("balance name not found in metadata")
		return "", fmt.Errorf("balance name not found in metadata")
	}

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:            balanceName,
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

	messageText := opts.message.GetText()
	if isPaginationNeeded(messageText) {
		creationPeriod, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationCreationPeriodMetadataKey)
		if !ok {
			logger.Error().Msg("creation period not found in metadata")
			return "", fmt.Errorf("creation period not found in metadata")
		}

		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData.Add(model.PageMetadataKey, nextPage)
		parsedCreationPeriod := model.CreationPeriod(creationPeriod)

		message, keyboard, err := h.getOperationsHistoryKeyboard(
			ctx,
			getOperationsHistoryKeyboardOptions{
				balance:        balance,
				creationPeriod: parsedCreationPeriod,
				page:           nextPage,
			},
		)
		if err != nil {
			logger.Error().Err(err).Msg("get operations keyboard")
			return "", fmt.Errorf("get operations keyboard: %w", err)
		}

		return model.ChooseTimePeriodForOperationsHistoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedInlineKeyboard:   keyboard,
			UpdatedMessage:          message,
		})
	}

	creationPeriod := model.GetCreationPeriodFromText(messageText)
	opts.stateMetaData.Add(model.OperationCreationPeriodMetadataKey, creationPeriod)

	message, keyboard, err := h.getOperationsHistoryKeyboard(
		ctx,
		getOperationsHistoryKeyboardOptions{
			balance:        balance,
			creationPeriod: creationPeriod,
			page:           firstPage,
		},
	)
	if err != nil {
		logger.Error().Err(err).Msg("get operations keyboard")
		return "", fmt.Errorf("get operations keyboard: %w", err)
	}

	return model.ChooseTimePeriodForOperationsHistoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		UpdatedMessage:          message,
		FormatMessageInMarkDown: true,
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		UpdatedInlineKeyboard:   keyboard,
	})
}

func (h handlerService) handleDeleteOperationFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleDeleteOperationFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose balance to delete operation from:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForDeleteOperation(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForDeleteOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.BalanceNameMetadataKey, opts.message.GetText())
	opts.stateMetaData.Add(model.PageMetadataKey, firstPage)

	keyboard, err := h.getOperationsKeyboard(ctx, getOperationsKeyboardOptions{
		balanceID: opts.user.GetBalance(opts.message.GetText()).ID,
		page:      firstPage,
	})
	if err != nil {
		return "", fmt.Errorf("get operations keyboard: %w", err)
	}

	return model.ChooseOperationToDeleteFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose operation to delete:",
		UpdatedInlineKeyboard: keyboard,
	})
}

func (h handlerService) handleChooseOperationToDeleteFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseOperationToDeleteFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	messageText := opts.message.GetText()
	if isPaginationNeeded(messageText) {
		balanceName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceNameMetadataKey)
		if !ok {
			logger.Error().Msg("balance name not found in metadata")
			return "", fmt.Errorf("balance name not found in metadata")
		}

		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData.Add(model.PageMetadataKey, nextPage)

		keyboard, err := h.getOperationsKeyboard(ctx, getOperationsKeyboardOptions{
			balanceID: opts.user.GetBalance(balanceName).ID,
			page:      nextPage,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get operations keyboard")
			return "", fmt.Errorf("get operations keyboard: %w", err)
		}

		return model.ChooseOperationToDeleteFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        "Choose operation to delete:",
			UpdatedInlineKeyboard: keyboard,
		})
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID: messageText,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return "", fmt.Errorf("get operation from store: %w", err)
	}
	if operation == nil {
		logger.Info().Msg("operation not found")
		return "", ErrOperationNotFound
	}

	opts.stateMetaData.Add(model.OperationIDMetadataKey, operation.ID)
	return model.ConfirmOperationDeletionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedInlineKeyboard: confirmationInlineKeyboardRows,
		UpdatedMessage:        operation.GetDeletionMessage(),
	})
}

func (h handlerService) handleConfirmOperationDeletionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleConfirmOperationDeletionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	confirmOperationDeletion, err := strconv.ParseBool(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse callback data to bool")
		return "", fmt.Errorf("parse callback data to bool: %w", err)
	}

	if !confirmOperationDeletion {
		logger.Info().Msg("user did not confirm balance deletion")
		return model.EndFlowStep, h.notifyCancellationAndShowKeyboard(opts.message, operationKeyboardRows)
	}

	balanceName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceNameMetadataKey)
	if !ok {
		logger.Error().Msg("balance name not found in metadata")
		return "", fmt.Errorf("balance name not found in metadata")
	}

	operationID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationIDMetadataKey)
	if !ok {
		logger.Error().Msg("operation id not found in metadata")
		return "", fmt.Errorf("operation id not found in metadata")
	}

	err = h.deleteOperation(ctx, deleteOperationOptions{
		user:        opts.user,
		balanceName: balanceName,
		operationID: operationID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("delete operation")
		return "", fmt.Errorf("delete operation: %w", err)
	}

	return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		UpdatedKeyboard: operationKeyboardRows,
		UpdatedMessage:  "Operation deleted!",
	})
}

type deleteOperationOptions struct {
	user        *model.User
	balanceName string
	operationID string
}

func (h handlerService) deleteOperation(ctx context.Context, opts deleteOperationOptions) error {
	logger := h.logger.With().Str("name", "handlerService.deleteOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	initialOperation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID: opts.operationID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return fmt.Errorf("get operation from store: %w", err)
	}
	if initialOperation == nil {
		logger.Info().Msg("operation not found")
		return ErrOperationNotFound
	}

	switch initialOperation.Type {
	case model.OperationTypeTransferIn, model.OperationTypeTransferOut:
		pairedOperation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
			ID: initialOperation.ParentOperationID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get operation from store")
			return fmt.Errorf("get operation from store: %w", err)
		}
		if pairedOperation == nil {
			logger.Info().Msg("paired operation not found")
			return ErrOperationNotFound
		}

		return h.deleteTransferOperation(ctx, initialOperation, pairedOperation, opts.user)
	case model.OperationTypeSpending, model.OperationTypeIncoming:
		balance := opts.user.GetBalance(opts.balanceName)
		if balance == nil {
			logger.Error().Msg("balance not found")
			return ErrBalanceNotFound
		}

		err := h.deleteSpendingOrIncomeOperation(ctx, initialOperation, balance)
		if err != nil {
			logger.Error().Err(err).Msgf("delete %s operation", initialOperation.Type)
			return fmt.Errorf("delete %s operation: %w", initialOperation.Type, err)
		}
	}

	return nil
}

// deleteTransferOperation handles the deletion of a transfer operation and its paired counterpart, adjusting the balances accordingly
func (h handlerService) deleteTransferOperation(ctx context.Context, initialOperation, pairedOperation *model.Operation, user *model.User) error {
	logger := h.logger.With().Str("name", "handlerService.deleteTransferOperation").Logger()
	logger.Debug().Any("operation", initialOperation).Any("user", user).Msg("got args")

	initialBalance := user.GetBalance(initialOperation.BalanceID)
	if initialBalance == nil {
		logger.Info().Msg("initial balance not found")
		return ErrBalanceNotFound
	}

	pairedBalance := user.GetBalance(pairedOperation.BalanceID)
	if pairedBalance == nil {
		logger.Info().Msg("paired balance not found")
		return ErrBalanceNotFound
	}

	initialBalanceAmount, _ := money.NewFromString(initialBalance.Amount)
	pairedBalanceAmount, _ := money.NewFromString(pairedBalance.Amount)

	initialOperationAmount, _ := money.NewFromString(initialOperation.Amount)
	pairedOperationAmount, _ := money.NewFromString(pairedOperation.Amount)

	var calculateOptions calculateTransferOperationOptions

	switch initialOperation.Type {
	case model.OperationTypeTransferIn:
		calculateOptions.balanceFrom = &pairedBalanceAmount
		calculateOptions.balanceTo = &initialBalanceAmount

		calculateOptions.transferAmountIn = &initialOperationAmount
		calculateOptions.transferAmountOut = &pairedOperationAmount
	case model.OperationTypeTransferOut:
		calculateOptions.balanceFrom = &initialBalanceAmount
		calculateOptions.balanceTo = &pairedBalanceAmount

		calculateOptions.transferAmountIn = &pairedOperationAmount
		calculateOptions.transferAmountOut = &initialOperationAmount
	}

	calculateDeletedTransferOperation(calculateOptions)

	for _, operation := range []string{initialOperation.ID, pairedOperation.ID} {
		err := h.stores.Operation.Delete(ctx, operation)
		if err != nil {
			logger.Error().Err(err).Msg("delete operation from store")
			return fmt.Errorf("delete operation from store: %w", err)
		}
	}

	initialBalance.Amount = initialBalanceAmount.StringFixed()
	pairedBalance.Amount = pairedBalanceAmount.StringFixed()

	for _, balance := range []*model.Balance{initialBalance, pairedBalance} {
		err := h.stores.Balance.Update(ctx, balance)
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
func (h handlerService) deleteSpendingOrIncomeOperation(ctx context.Context, operation *model.Operation, balance *model.Balance) error {
	logger := h.logger.With().Str("name", "handlerService.deleteSpendingOrIncomeOperation").Logger()
	logger.Debug().Any("operation", operation).Any("balance", balance).Msg("got args")

	balanceAmount, _ := money.NewFromString(balance.Amount)
	operationAmount, _ := money.NewFromString(operation.Amount)

	switch operation.Type {
	case model.OperationTypeSpending:
		calculateDeletedSpendingOperation(&balanceAmount, operationAmount)
		balance.Amount = balanceAmount.StringFixed()
	case model.OperationTypeIncoming:
		calculateDeletedIncomingOperation(&balanceAmount, operationAmount)
		balance.Amount = balanceAmount.StringFixed()
	}

	err := h.stores.Balance.Update(ctx, balance)
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

func (h handlerService) handleUpdateOperationFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateOperationFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose balance to update operation from:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForUpdateOperation(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForUpdateOperation").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.BalanceNameMetadataKey, opts.message.GetText())
	opts.stateMetaData.Add(model.PageMetadataKey, firstPage)

	keyboard, err := h.getOperationsKeyboard(ctx, getOperationsKeyboardOptions{
		balanceID: opts.user.GetBalance(opts.message.GetText()).ID,
		page:      firstPage,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operations keyboard")
		return "", fmt.Errorf("get operations keyboard: %w", err)
	}

	return model.ChooseOperationToUpdateFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose operation to update:",
		UpdatedInlineKeyboard: keyboard,
	})
}

func (h handlerService) handleChooseOperationToUpdateFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseOperationToUpdateFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	messageText := opts.message.GetText()
	if isPaginationNeeded(messageText) {
		balanceName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceNameMetadataKey)
		if !ok {
			logger.Error().Msg("balance name not found in metadata")
			return "", fmt.Errorf("balance name not found in metadata")
		}

		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData.Add(model.PageMetadataKey, nextPage)

		keyboard, err := h.getOperationsKeyboard(ctx, getOperationsKeyboardOptions{
			balanceID: opts.user.GetBalance(balanceName).ID,
			page:      nextPage,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get operations keyboard")
			return "", fmt.Errorf("get operations keyboard: %w", err)
		}

		return model.ChooseOperationToUpdateFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        "Choose operation to update:",
			UpdatedInlineKeyboard: keyboard,
		})
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         messageText,
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

	opts.stateMetaData.Add(model.OperationIDMetadataKey, operation.ID)

	outputMessage := fmt.Sprintf("Choose update operation option:\n%s", operation.GetDetails())
	return model.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        outputMessage,
		UpdatedInlineKeyboard: h.getUpdateOptionKeyboardByOperationType(operation.Type),
	})
}

func (h handlerService) handleChooseUpdateOperationOptionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseUpdateOperationOptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationIDMetadataKey)
	if !ok {
		logger.Error().Msg("operation id not found in metadata")
		return "", fmt.Errorf("operation id not found in metadata")
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         operationID,
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

	switch opts.message.GetText() {
	case model.BotUpdateOperationAmountCommand:
		return model.EnterOperationAmountFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedMessage:          fmt.Sprintf("Enter updated operation amount(Current: `%s`):", operation.Amount),
		})
	case model.BotUpdateOperationDescriptionCommand:
		return model.EnterOperationDescriptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedMessage:          fmt.Sprintf("Enter updated operation description(Current: `%s`):", operation.Description),
		})
	case model.BotUpdateOperationCategoryCommand:
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

		var currentCategory string
		categoriesWithoutAlreadyUsedCategory := slices.DeleteFunc(categories, func(category model.Category) bool {
			currentCategory = category.Title
			return category.ID == operation.CategoryID
		})

		// User does not have enough categories to choose from
		if len(categoriesWithoutAlreadyUsedCategory) == 0 {
			err := h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
				ChatID:          opts.message.GetChatID(),
				MessageID:       opts.message.GetMessageID(),
				InlineMessageID: opts.message.GetInlineMessageID(),
				UpdatedMessage:  ErrNotEnoughCategories.Message,
			})
			if err != nil {
				return "", fmt.Errorf("update message: %w", err)
			}

			return model.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
				ChatID:                  opts.message.GetChatID(),
				FormatMessageInMarkDown: true,
				Message:                 "Please choose other update operation option or finish action by canceling it!",
				InlineKeyboard:          h.getUpdateOptionKeyboardByOperationType(operation.Type),
			})
		}

		return model.ChooseCategoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedMessage:          fmt.Sprintf("Choose updated operation category(Current: `%s`):", currentCategory),
			UpdatedInlineKeyboard:   getInlineKeyboardRows(categoriesWithoutAlreadyUsedCategory, 3),
		})
	case model.BotUpdateOperationDateCommand:
		return model.EnterOperationDateFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedMessage: fmt.Sprintf(
				"Enter updated operation date(Current: `%s`):\nPlease use the following format: DD/MM/YYYY HH:MM. Example: 01/01/2025 12:00",
				operation.CreatedAt.Format(operationTimeFormat),
			),
		})
	default:
		return "", fmt.Errorf("received unknown update operation option: %s", opts.message.GetText())
	}
}

func (h handlerService) handleEnterOperationAmountFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationAmountFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationAmount, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation amount")
		return "", ErrInvalidAmountFormat
	}

	operationID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationIDMetadataKey)
	if !ok {
		logger.Error().Msg("operation id not found in metadata")
		return "", fmt.Errorf("operation id not found in metadata")
	}

	initialOperation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         operationID,
		BalanceIDs: opts.user.GetBalancesIDs(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get operation from store")
		return "", fmt.Errorf("get operation from store: %w", err)
	}
	if initialOperation == nil {
		logger.Info().Msg("operation not found")
		return "", ErrOperationNotFound
	}

	switch initialOperation.Type {
	case model.OperationTypeSpending, model.OperationTypeIncoming:
		balance := opts.user.GetBalance(initialOperation.BalanceID)
		if balance == nil {
			logger.Info().Msg("balance not found")
			return "", ErrBalanceNotFound
		}
		logger.Debug().Any("balance", balance).Msg("got balance")

		err := h.updateOperationAmountForSpendingOrIncomeOperation(ctx, balance, initialOperation, operationAmount)
		if err != nil {
			logger.Error().Err(err).Msgf("update operation amount for %s", initialOperation.Type)
			return "", fmt.Errorf("update operation amount for %s: %w", initialOperation.Type, err)
		}

		return model.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:                  opts.message.GetChatID(),
			FormatMessageInMarkDown: true,
			Message: fmt.Sprintf(
				"Operation amount successfully updated!\nNew amount: `%s`\nPlease choose other update operation option or finish action by canceling it!",
				operationAmount.String(),
			),
			InlineKeyboard: updateOperationOptionsKeyboardForIncomingAndSpendingOperations,
		})

	case model.OperationTypeTransferIn, model.OperationTypeTransferOut:
		pairedOperation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
			ID: initialOperation.ParentOperationID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get paired operation")
			return "", fmt.Errorf("get paired operation: %w", err)
		}
		if pairedOperation == nil {
			logger.Error().Msg("paired operation not found")
			return "", fmt.Errorf("paired operation not found")
		}

		err = h.updateOperationAmountForTransferOperation(ctx, opts.user, initialOperation, pairedOperation, operationAmount)
		if err != nil {
			logger.Error().Err(err).Msg("update operation amount for transfer operation")
			return "", fmt.Errorf("update operation amount for transfer operation: %w", err)
		}

		return model.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:                  opts.message.GetChatID(),
			FormatMessageInMarkDown: true,
			Message: fmt.Sprintf(
				"Operation amount successfully updated!\nNew operation amount: `%s`\nNew paired operation amount: `%s`\nPlease choose other update operation option or finish action by canceling it!",
				initialOperation.Amount, pairedOperation.Amount,
			),
			InlineKeyboard: updateOperationOptionsKeyboardForTransferOperations,
		})
	}

	return "", nil
}

func (h handlerService) updateOperationAmountForSpendingOrIncomeOperation(ctx context.Context, balance *model.Balance, operation *model.Operation, updatedOperationAmount money.Money) error {
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
	case model.OperationTypeIncoming:
		calculateUpdatedIncomingOperation(&balanceAmount, operationAmount, updatedOperationAmount)
	case model.OperationTypeSpending:
		calculateUpdatedSpendingOperation(&balanceAmount, operationAmount, updatedOperationAmount)
	}
	balance.Amount = balanceAmount.StringFixed()

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

func (h handlerService) updateOperationAmountForTransferOperation(ctx context.Context, user *model.User, initialOperation, pairedOperation *model.Operation, updatedOperationAmount money.Money) error {
	logger := h.logger.With().Str("name", "handlerService.updateOperationAmountForTransferOperation").Logger()
	logger.Debug().
		Any("operation", initialOperation).
		Any("pairedOperation", pairedOperation).
		Any("user", user).
		Any("updatedOperationAmount", updatedOperationAmount).
		Msg("got args")

	initialBalance := user.GetBalance(initialOperation.BalanceID)
	if initialBalance == nil {
		logger.Info().Msg("initial balance not found")
		return ErrBalanceNotFound
	}

	pairedBalance := user.GetBalance(pairedOperation.BalanceID)
	if pairedBalance == nil {
		logger.Info().Msg("paired balance not found")
		return ErrBalanceNotFound
	}

	initialBalanceAmount, _ := money.NewFromString(initialBalance.Amount)
	pairedBalanceAmount, _ := money.NewFromString(pairedBalance.Amount)
	initialOperationAmount, _ := money.NewFromString(initialOperation.Amount)
	pairedOperationAmount, _ := money.NewFromString(pairedOperation.Amount)

	calculateOptions := calculateTransferOperationOptions{
		operationType:          initialOperation.Type,
		updatedOperationAmount: updatedOperationAmount,
	}

	if initialOperation.ExchangeRate != "" {
		parsedExchangeRate, _ := money.NewFromString(initialOperation.ExchangeRate)
		calculateOptions.exchangeRate = &parsedExchangeRate
	}

	switch initialOperation.Type {
	case model.OperationTypeTransferIn:
		calculateOptions.transferAmountIn = &initialOperationAmount
		calculateOptions.transferAmountOut = &pairedOperationAmount
		calculateOptions.balanceTo = &initialBalanceAmount
		calculateOptions.balanceFrom = &pairedBalanceAmount
	case model.OperationTypeTransferOut:
		calculateOptions.transferAmountOut = &initialOperationAmount
		calculateOptions.transferAmountIn = &pairedOperationAmount
		calculateOptions.balanceTo = &pairedBalanceAmount
		calculateOptions.balanceFrom = &initialBalanceAmount
	}

	calculateUpdatedTranferOperation(calculateOptions)

	initialOperation.Amount = initialOperationAmount.StringFixed()
	pairedOperation.Amount = pairedOperationAmount.StringFixed()

	for _, operation := range []*model.Operation{initialOperation, pairedOperation} {
		err := h.stores.Operation.Update(ctx, operation.ID, operation)
		if err != nil {
			logger.Error().Err(err).Msg("delete operation from store")
			return fmt.Errorf("delete operation from store: %w", err)
		}
	}

	initialBalance.Amount = initialBalanceAmount.StringFixed()
	pairedBalance.Amount = pairedBalanceAmount.StringFixed()

	for _, balance := range []*model.Balance{initialBalance, pairedBalance} {
		err := h.stores.Balance.Update(ctx, balance)
		if err != nil {
			logger.Error().Err(err).Msg("delete balance from store")
			return fmt.Errorf("delete balance from store: %w", err)
		}
	}

	return nil
}

func (h handlerService) handleEnterOperationDescriptionFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationDescriptionFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationIDMetadataKey)
	if !ok {
		logger.Error().Msg("operation id not found in metadata")
		return "", fmt.Errorf("operation id not found in metadata")
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         operationID,
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

	return model.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		FormatMessageInMarkDown: true,
		Message: fmt.Sprintf(
			"Operation description successfully updated!\nNew operation description: `%s`\nPlease choose other update operation option or finish action by canceling it!",
			operation.Description,
		),
		InlineKeyboard: h.getUpdateOptionKeyboardByOperationType(operation.Type),
	})
}

func (h handlerService) handleChooseCategoryFlowStepForOperationUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
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

	operationID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationIDMetadataKey)
	if !ok {
		logger.Error().Msg("operation id not found in metadata")
		return "", fmt.Errorf("operation id not found in metadata")
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         operationID,
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

	return model.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage: fmt.Sprintf(
			"Operation category successfully updated!\nNew category: `%s`\nPlease choose other update operation option or finish action by canceling it!",
			category.Title,
		),
		UpdatedInlineKeyboard: h.getUpdateOptionKeyboardByOperationType(operation.Type),
	})
}

const operationTimeFormat = "02/01/2006 15:04"

func (h handlerService) handleEnterOperationDateFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationDateFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.OperationIDMetadataKey)
	if !ok {
		logger.Error().Msg("operation id not found in metadata")
		return "", fmt.Errorf("operation id not found in metadata")
	}

	operation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
		ID:         operationID,
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

	parsedOperationDate, err := time.Parse(operationTimeFormat, opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation date")
		return "", ErrInvalidDateFormat
	}

	switch operation.Type {
	case model.OperationTypeSpending, model.OperationTypeIncoming:
		operation.CreatedAt = parsedOperationDate

		err = h.stores.Operation.Update(ctx, operation.ID, operation)
		if err != nil {
			logger.Error().Err(err).Msg("update operation in store")
			return "", fmt.Errorf("update operation in store: %w", err)
		}

	case model.OperationTypeTransferIn, model.OperationTypeTransferOut:
		pairedOperation, err := h.stores.Operation.Get(ctx, GetOperationFilter{
			ID: operation.ParentOperationID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get operation from store")
			return "", fmt.Errorf("get operation from store: %w", err)
		}
		if pairedOperation == nil {
			logger.Info().Msg("paired operation not found")
			return "", ErrOperationNotFound
		}

		for _, operation := range []*model.Operation{operation, pairedOperation} {
			operation.CreatedAt = parsedOperationDate
			err = h.stores.Operation.Update(ctx, operation.ID, operation)
			if err != nil {
				logger.Error().Err(err).Msg("update operation in store")
				return "", fmt.Errorf("update operation in store: %w", err)
			}
		}
	}

	return model.ChooseUpdateOperationOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		FormatMessageInMarkDown: true,
		Message: fmt.Sprintf(
			"Operation category successfully updated!\nNew operation date: `%s`\nPlease choose other update operation option or finish action by canceling it!",
			parsedOperationDate.Format(operationTimeFormat),
		),
		InlineKeyboard: h.getUpdateOptionKeyboardByOperationType(operation.Type),
	})
}

func (h handlerService) getUpdateOptionKeyboardByOperationType(operationType model.OperationType) []InlineKeyboardRow {
	var updateOperationOptionsKeyboard []InlineKeyboardRow
	switch operationType {
	case model.OperationTypeTransferIn, model.OperationTypeTransferOut:
		updateOperationOptionsKeyboard = updateOperationOptionsKeyboardForTransferOperations
	default:
		updateOperationOptionsKeyboard = updateOperationOptionsKeyboardForIncomingAndSpendingOperations
	}

	return updateOperationOptionsKeyboard
}
