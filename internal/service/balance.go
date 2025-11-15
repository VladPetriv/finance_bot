package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

func (h handlerService) handleCreateInitialBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateInitialBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceID := uuid.NewString()
	message := opts.message.GetText()

	err := h.stores.Balance.Create(ctx, &model.Balance{
		ID:     balanceID,
		UserID: opts.user.ID,
		Name:   message,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create balance in store")
		return "", fmt.Errorf("create balance in store: %w", err)
	}

	opts.stateMetaData.Add(model.BalanceIDMetadataKey, balanceID)
	return model.EnterBalanceAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Please enter balance amount:")
}

func (h handlerService) handleCreateBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.BalanceIDMetadataKey, uuid.NewString())
	return model.EnterBalanceNameFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Please enter balance name:",
		Keyboard: rowKeyboardWithCancelButtonOnly,
	})
}

func (h handlerService) handleEnterBalanceNameFlowStepForCreate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceNameFlowStepForCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:   opts.message.GetText(),
		UserID: opts.user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return "", fmt.Errorf("get balance from store: %w", err)
	}
	if balance != nil {
		logger.Info().Msg("balance with entered name already exists")
		return model.EnterBalanceNameFlowStep, ErrBalanceAlreadyExists
	}

	opts.stateMetaData.Add(model.BalanceNameMetadataKey, opts.message.GetText())
	return model.EnterBalanceAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter balance amount:")
}

func (h handlerService) handleEnterBalanceAmountFlowStepForCreate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceAmountFlowStepForCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	parsedAmount, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("convert option amount to money type")
		return "", ErrInvalidAmountFormat
	}

	opts.stateMetaData.Add(model.BalanceAmountMetadataKey, parsedAmount.StringFixed())
	opts.stateMetaData.Add(model.PageMetadataKey, firstPage)

	currenciesKeyboard, err := h.getCurrenciesKeyboard(ctx, firstPage)
	if err != nil {
		logger.Error().Err(err).Msg("get currencies keyboard for balance")
		return "", fmt.Errorf("get currencies keyboard for balance: %w", err)
	}

	return model.EnterBalanceCurrencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Enter balance currency:",
		InlineKeyboard: currenciesKeyboard,
	})
}

func (h handlerService) handleEnterBalanceCurrencyFlowStepForCreate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceCurrencyFlowStepForCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	messageText := opts.message.GetText()
	if isPaginationNeeded(messageText) {
		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData.Add(model.PageMetadataKey, nextPage)

		currenciesKeyboard, err := h.getCurrenciesKeyboard(ctx, nextPage)
		if err != nil {
			logger.Error().Err(err).Msg("get currencies keyboard for balance")
			return "", fmt.Errorf("get currencies keyboard for balance: %w", err)
		}

		return model.EnterBalanceCurrencyFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			UpdatedMessage:        "Enter balance currency:",
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedInlineKeyboard: currenciesKeyboard,
		})
	}

	currency, err := h.stores.Currency.Get(ctx, GetCurrencyFilter{
		ID: messageText,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get currency from store")
		return "", fmt.Errorf("get currency from store: %w", err)
	}
	if currency == nil {
		logger.Error().Msg("currency not found")
		return model.EnterBalanceCurrencyFlowStep, ErrCurrencyNotFound
	}

	balanceID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceIDMetadataKey)
	if !ok {
		logger.Error().Msg("balance id not found in metadata")
		return "", fmt.Errorf("balance id not found")
	}

	balanceName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceNameMetadataKey)
	if !ok {
		logger.Error().Msg("balance name not found in metadata")
		return "", fmt.Errorf("balance name not found")
	}

	balanceAmount, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceAmountMetadataKey)
	if !ok {
		logger.Error().Msg("balance amount not found in metadata")
		return "", fmt.Errorf("balance amount not found")
	}

	balance := model.Balance{
		ID:         balanceID,
		UserID:     opts.user.ID,
		CurrencyID: messageText,
		Name:       balanceName,
		Amount:     balanceAmount,
	}

	err = h.stores.Balance.Create(ctx, &balance)
	if err != nil {
		logger.Error().Err(err).Msg("create balance in store")
		return "", fmt.Errorf("create balance in store: %w", err)
	}

	outputMessage := fmt.Sprintf(
		"Balance Created!\nBalance:\n - Name: %s\n - Amount: %s\n - Currency: %s",
		balance.Name, balance.Amount, currency.GetName(),
	)

	return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		UpdatedKeyboard: balanceKeyboardRows,
		UpdatedMessage:  outputMessage,
	})
}

const maxMonthsPerRow = 3

func (h handlerService) handleGetBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleGetBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	availableMonths := model.Months[:time.Now().Month()]

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		return "", fmt.Errorf("show cancle button: %w", err)
	}

	return model.ChooseMonthBalanceStatisticsFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Please select a month to view your balance statistics:",
		InlineKeyboard: getInlineKeyboardRows(availableMonths, maxMonthsPerRow),
	})
}

func (h handlerService) handleChooseMonthBalanceStatisticsFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseMonthBalanceStatisticsFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.MonthForBalanceStatisticsKey, opts.message.GetText())
	return model.ChooseBalanceFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Select a balance to view information:",
		UpdatedInlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForGetBalance(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForGetBalance").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:            opts.message.GetText(),
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

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: opts.user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list categories from store")
		return "", fmt.Errorf("list categories from store: %w", err)
	}

	monthForBalanceStatistics, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.MonthForBalanceStatisticsKey)
	if !ok {
		logger.Error().Msg("month for balance statistics not found in metadata")
		return "", fmt.Errorf("month for balance statistics not found")
	}

	operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
		BalanceID: balance.ID,
		Month:     model.Month(monthForBalanceStatistics),
	})
	if err != nil {
		logger.Error().Err(err).Msg("list operations from store")
		return "", fmt.Errorf("list operations from store: %w", err)
	}

	outputMessage, err := model.
		NewStatisticsMessageBuilder(balance, operations, categories).
		Build(model.Month(monthForBalanceStatistics))
	if err != nil {
		logger.Error().Err(err).Msg("build statistic message")
		return "", fmt.Errorf("build statistic message: %w", err)
	}

	return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		UpdatedMessage:          outputMessage,
		FormatMessageInMarkDown: true,
		UpdatedKeyboard:         balanceKeyboardRows,
	})
}

func (h handlerService) handleUpdateBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose balance to update:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		logger.Error().Msg("balance not found")
		return "", fmt.Errorf("balance not found")
	}

	opts.stateMetaData.Add(model.BalanceIDMetadataKey, balance.ID)
	opts.stateMetaData.Add(model.CurrentBalanceNameMetadataKey, balance.Name)

	return model.ChooseUpdateBalanceOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose update balance option:",
		UpdatedInlineKeyboard: updateBalanceOptionsKeyboard,
	})
}

func (h handlerService) handleChooseUpdateBalanceOptionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseUpdateBalanceOptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceIDMetadataKey)
	if !ok {
		logger.Error().Msg("balance ID not found in metadata")
		return "", fmt.Errorf("balance ID not found in metadata")
	}

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		BalanceID:       balanceID,
		PreloadCurrency: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return "", fmt.Errorf("get balance from store: %w", err)
	}
	if balance == nil {
		logger.Warn().Msg("balance not found")
		return "", ErrBalanceNotFound
	}
	logger.Debug().Any("balance", balance).Msg("got balance from store")

	switch opts.message.GetText() {
	case model.BotUpdateBalanceNameCommand:
		return model.EnterBalanceNameFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			UpdatedMessage:          fmt.Sprintf("Enter updated balance name(Current: `%s`):", balance.Name),
			FormatMessageInMarkDown: true,
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
		})
	case model.BotUpdateBalanceAmountCommand:
		return model.EnterBalanceAmountFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			UpdatedMessage:          fmt.Sprintf("Enter updated balance amount(Current: `%s`):", balance.Amount),
			FormatMessageInMarkDown: true,
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
		})
	case model.BotUpdateBalanceCurrencyCommand:
		opts.stateMetaData.Add(model.PageMetadataKey, firstPage)
		currenciesKeyboard, err := h.getCurrenciesKeyboard(ctx, firstPage)
		if err != nil {
			logger.Error().Err(err).Msg("get currencies keyboard for balance")
			return "", fmt.Errorf("get currencies keyboard for balance: %w", err)
		}

		return model.EnterBalanceCurrencyFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			UpdatedMessage:          fmt.Sprintf("Enter balance currency(Current: `%s`):", balance.Currency.GetName()),
			FormatMessageInMarkDown: true,
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			UpdatedInlineKeyboard:   currenciesKeyboard,
		})
	default:
		return "", fmt.Errorf("received unknown update balance option: %s", opts.message.GetText())
	}
}

func (h handlerService) handleEnterBalanceNameFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceNameFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	text := opts.message.GetText()
	currentBalanceName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.CurrentBalanceNameMetadataKey)
	if !ok {
		return "", fmt.Errorf("current balance name not found in metadata")
	}

	shouldNotCheckForBalanceAlreadyExistsInStore := currentBalanceName != "" && text == currentBalanceName
	if !shouldNotCheckForBalanceAlreadyExistsInStore {
		balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
			Name:   text,
			UserID: opts.user.ID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get balance from store")
			return "", fmt.Errorf("get balance from store: %w", err)
		}
		if balance != nil {
			logger.Info().Msg("balance with entered name already exists")
			return model.EnterBalanceNameFlowStep, ErrBalanceAlreadyExists
		}
	}

	balanceID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceIDMetadataKey)
	if !ok {
		return "", fmt.Errorf("balance ID not found in metadata")
	}

	updatedBalance, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: balanceID,
		step:      model.EnterBalanceNameFlowStep,
		data:      text,
	})
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return "", err
		}

		logger.Error().Err(err).Msg("update balance in store")
		return "", fmt.Errorf("update balance in store: %w", err)
	}

	return model.ChooseUpdateBalanceOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		FormatMessageInMarkDown: true,
		Message: fmt.Sprintf(
			"Balance name successfully updated!\nNew balance name: `%s`\nPlease choose other update balance option or finish action by canceling it!",
			updatedBalance.Name,
		),
		InlineKeyboard: updateBalanceOptionsKeyboard,
	})
}

func (h handlerService) handleEnterBalanceAmountFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceAmountFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceIDMetadataKey)
	if !ok {
		return "", fmt.Errorf("balance ID not found in metadata")
	}

	updatedBalance, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: balanceID,
		step:      model.EnterBalanceAmountFlowStep,
		data:      opts.message.GetText(),
	})
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return "", err
		}

		logger.Error().Err(err).Msg("update balance in store")
		return "", fmt.Errorf("update balance in store: %w", err)
	}

	if opts.state.Flow == model.StartFlow {
		opts.stateMetaData.Add(model.PageMetadataKey, firstPage)
		currenciesKeyboard, err := h.getCurrenciesKeyboard(ctx, firstPage)
		if err != nil {
			logger.Error().Err(err).Msg("get currencies keyboard for balance")
			return "", fmt.Errorf("get currencies keyboard for balance: %w", err)
		}

		return model.EnterBalanceCurrencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:         opts.message.GetChatID(),
			Message:        "Enter balance currency:",
			InlineKeyboard: currenciesKeyboard,
		})
	}

	return model.ChooseUpdateBalanceOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		FormatMessageInMarkDown: true,
		Message: fmt.Sprintf(
			"Balance amount successfully updated!\nNew balance amount: `%s`\nPlease choose other update balance option or finish action by canceling it!",
			updatedBalance.Amount,
		),
		InlineKeyboard: updateBalanceOptionsKeyboard,
	})
}

func (h handlerService) handleEnterBalanceCurrencyFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceCurrencyFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	messageText := opts.message.GetText()
	if isPaginationNeeded(messageText) {
		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData.Add(model.PageMetadataKey, nextPage)

		currenciesKeyboard, err := h.getCurrenciesKeyboard(ctx, nextPage)
		if err != nil {
			logger.Error().Err(err).Msg("get currencies keyboard for balance")
			return "", fmt.Errorf("get currencies keyboard for balance: %w", err)
		}

		return model.EnterBalanceCurrencyFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			UpdatedMessage:        "Enter balance currency:",
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedInlineKeyboard: currenciesKeyboard,
		})
	}

	currency, err := h.stores.Currency.Get(ctx, GetCurrencyFilter{
		ID: messageText,
	})
	if err != nil {
		logger.Error().Err(err).Msg("check if currency exists in store")
		return "", fmt.Errorf("check if currency exists in store: %w", err)
	}
	if currency == nil {
		logger.Error().Msg("currency not found")
		return model.EnterBalanceCurrencyFlowStep, ErrCurrencyNotFound
	}

	balanceID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceIDMetadataKey)
	if !ok {
		logger.Error().Msg("balance id not found in metadata")
		return "", fmt.Errorf("balance id not found in metadata")
	}

	balance, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: balanceID,
		step:      model.EnterBalanceCurrencyFlowStep,
		data:      messageText,
	})
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return "", err
		}

		logger.Error().Err(err).Msg("update balance in store")
		return "", fmt.Errorf("update balance in store: %w", err)
	}

	if opts.state.Flow == model.StartFlow {
		outputMessage := fmt.Sprintf(
			"Initial balance successfully created!\nBalance:\n\t - Name: `%s`\n\t - Amount: `%s`\n\t - Currency: `%s`",
			balance.Name,
			balance.Amount,
			balance.Currency.GetName(),
		)

		return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			FormatMessageInMarkDown: true,
			InlineMessageID:         opts.message.GetInlineMessageID(),
			UpdatedKeyboard:         defaultKeyboardRows,
			UpdatedMessage:          outputMessage,
		})
	}

	return model.ChooseUpdateBalanceOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage: fmt.Sprintf(
			"Balance currency successfully updated!\nNew balance currency: `%s`\nPlease choose other update balance option or finish action by canceling it!",
			balance.Currency.GetName(),
		),
		UpdatedInlineKeyboard: updateBalanceOptionsKeyboard,
	})
}

type updateBalanceOptions struct {
	balanceID string
	step      model.FlowStep
	data      string
}

func (h handlerService) updateBalance(ctx context.Context, opts updateBalanceOptions) (*model.Balance, error) {
	logger := h.logger.With().Str("name", "handlerService.updateBalance").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		BalanceID: opts.balanceID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return nil, fmt.Errorf("get balance from store: %w", err)
	}
	if balance == nil {
		logger.Error().Msg("balance not found")
		return nil, fmt.Errorf("balance not found")
	}
	logger.Debug().Any("balance", balance).Msg("got balance from store")

	switch opts.step {
	case model.EnterBalanceNameFlowStep:
		balance.Name = opts.data
	case model.EnterBalanceAmountFlowStep:
		price, err := money.NewFromString(opts.data)
		if err != nil {
			logger.Error().Err(err).Msg("convert option amount to money type")
			return nil, ErrInvalidAmountFormat
		}

		balance.Amount = price.StringFixed()
	case model.EnterBalanceCurrencyFlowStep:
		balance.CurrencyID = opts.data
	}

	err = h.stores.Balance.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return nil, fmt.Errorf("update balance in store: %w", err)
	}

	updatedBalance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		BalanceID:       opts.balanceID,
		PreloadCurrency: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return nil, fmt.Errorf("get balance from store: %w", err)
	}

	return updatedBalance, nil
}

func (h handlerService) handleDeleteBalanceFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleDeleteBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if len(opts.user.Balances) == 1 {
		return model.EndFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   opts.message.GetChatID(),
			Message:  "You're not allowed to delete last balance!",
			Keyboard: balanceKeyboardRows,
		})
	}

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ConfirmBalanceDeletionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose balance to delete:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h handlerService) handleConfirmBalanceDeletionFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleConfirmBalanceDeletionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.BalanceNameMetadataKey, opts.message.GetText())
	outputMessage := fmt.Sprintf(
		"Are you sure you want to delete balance %s?\nPlease note that all its operations will be deleted as well.",
		opts.message.GetText(),
	)

	return model.ChooseBalanceFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		UpdatedMessage:        outputMessage,
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedInlineKeyboard: confirmationInlineKeyboardRows,
	})
}

func (h handlerService) handleChooseBalanceFlowStepForDelete(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForDelete").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	confirmBalanceDeletion, err := strconv.ParseBool(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse callback data to bool")
		return "", fmt.Errorf("parse callback data to bool: %w", err)
	}

	if !confirmBalanceDeletion {
		logger.Info().Msg("user did not confirm balance deletion")
		return model.EndFlowStep, h.notifyCancellationAndShowKeyboard(opts.message, balanceKeyboardRows)
	}

	balanceName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceNameMetadataKey)
	if !ok {
		logger.Error().Msg("balance name not found in metadata")
		return "", fmt.Errorf("balance name not found in metadata")
	}

	balance := opts.user.GetBalance(balanceName)
	if balance == nil {
		logger.Error().Msg("balance for deletion not found")
		return "", fmt.Errorf("balance for deletion not found")
	}
	logger.Debug().Any("balance", balance).Msg("got balance for deletion")

	// Run in separate goroutine to not block the main thread and respond to the user as soon as possible.
	go func() {
		balanceOperations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
			BalanceID: balance.ID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("list operations from store")

			return
		}

		for _, operation := range balanceOperations {
			err := h.stores.Operation.Delete(ctx, operation.ID)
			if err != nil {
				logger.Error().Err(err).Str("operationID", operation.ID).Msg("delete operation from store")

				continue
			}
		}

		err = h.stores.Balance.Delete(ctx, balance.ID)
		if err != nil {
			logger.Error().Err(err).Msg("delete balance from store")

			return
		}
	}()

	return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		UpdatedKeyboard: balanceKeyboardRows,
		UpdatedMessage:  "Balance and all its operations have been deleted!",
	})
}
