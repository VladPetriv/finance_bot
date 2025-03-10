package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

func (h handlerService) handleCreateInitialBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateInitialBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceID := uuid.NewString()
	message := opts.message.GetText()

	err := h.stores.Balance.Create(ctx, &models.Balance{
		ID:     balanceID,
		UserID: opts.user.ID,
		Name:   message,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create balance in store")
		return "", fmt.Errorf("create balance in store: %w", err)
	}

	opts.stateMetaData[balanceIDMetadataKey] = balanceID
	return models.EnterBalanceAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Please enter balance amount:")
}

func (h handlerService) handleCreateBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceIDMetadataKey] = uuid.NewString()

	return models.EnterBalanceNameFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Please enter balance name:",
		Keyboard: rowKeyboardWithCancelButtonOnly,
	})
}

func (h handlerService) handleEnterBalanceNameFlowStepForCreate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
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
		return models.EnterBalanceNameFlowStep, ErrBalanceAlreadyExists
	}

	opts.stateMetaData[balanceNameMetadataKey] = opts.message.GetText()

	return models.EnterBalanceAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter balance amount:")
}

func (h handlerService) handleEnterBalanceAmountFlowStepForCreate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceAmountFlowStepForCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	parsedAmount, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("convert option amount to money type")
		return "", ErrInvalidAmountFormat
	}

	opts.stateMetaData[balanceAmountMetadataKey] = parsedAmount.StringFixed()

	currenciesKeyboard, err := h.getCurrenciesKeyboardForBalance(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("get currencies keyboard for balance")
		return "", fmt.Errorf("get currencies keyboard for balance: %w", err)
	}

	return models.EnterBalanceCurrencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Enter balance currency:",
		InlineKeyboard: currenciesKeyboard,
	})
}

func (h handlerService) handleEnterBalanceCurrencyFlowStepForCreate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceCurrencyFlowStepForCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	exists, err := h.stores.Currency.Exists(ctx, ExistsCurrencyFilter{
		ID: opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("check if currency exists in store")
		return "", fmt.Errorf("check if currency exists in store: %w", err)
	}
	if !exists {
		logger.Error().Msg("currency not found")
		return models.EnterBalanceCurrencyFlowStep, ErrCurrencyNotFound
	}

	balance := models.Balance{
		ID:         opts.stateMetaData[balanceIDMetadataKey].(string),
		UserID:     opts.user.ID,
		CurrencyID: opts.message.GetText(),
		Name:       opts.stateMetaData[balanceNameMetadataKey].(string),
		Amount:     opts.stateMetaData[balanceAmountMetadataKey].(string),
	}

	err = h.stores.Balance.Create(ctx, &balance)
	if err != nil {
		logger.Error().Err(err).Msg("create balance in store")
		return "", fmt.Errorf("create balance in store: %w", err)
	}

	outputMessage := fmt.Sprintf(
		"Balance Created!\nBalance Info:\n - Name: %s\n - Amount: %s",
		balance.Name, balance.Amount,
	)
	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), outputMessage)
}

const maxMonthsPerRow = 3

func (h handlerService) handleGetBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleGetBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	availableMonths := models.Months[:time.Now().Month()]

	return models.ChooseMonthBalanceStatisticsFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Please select a month to view your balance statistics:",
		Keyboard: getKeyboardRows(availableMonths, maxMonthsPerRow, true),
	})
}

func (h handlerService) handleChooseMonthBalanceStatisticsFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseMonthBalanceStatisticsFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[monthForBalanceStatisticsKey] = opts.message.GetText()

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Select a balance to view information:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForGetBalance(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
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

	monthForBalanceStatistics := models.Month(opts.stateMetaData[monthForBalanceStatisticsKey].(string))

	operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
		BalanceID: balance.ID,
		Month:     monthForBalanceStatistics,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list operations from store")
		return "", fmt.Errorf("list operations from store: %w", err)
	}

	outputMessage, err := models.
		NewStatisticsMessageBuilder(balance, operations, categories).
		Build(monthForBalanceStatistics)
	if err != nil {
		logger.Error().Err(err).Msg("build statistic message")
		return "", fmt.Errorf("build statistic message: %w", err)
	}

	// Unescape markdown symbols.
	outputMessage = strings.ReplaceAll(outputMessage, "(", `\(`)
	outputMessage = strings.ReplaceAll(outputMessage, ")", `\)`)
	outputMessage = strings.ReplaceAll(outputMessage, "!", `\!`)
	outputMessage = strings.ReplaceAll(outputMessage, "-", `\-`)
	outputMessage = strings.ReplaceAll(outputMessage, "+", `\+`)
	outputMessage = strings.ReplaceAll(outputMessage, ".", `\.`)

	return models.EndFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		Message:                 outputMessage,
		FormatMessageInMarkDown: true,
		Keyboard:                defaultKeyboardRows,
	})
}

func (h handlerService) handleUpdateBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose balance to update:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		logger.Error().Msg("balance not found")
		return "", fmt.Errorf("balance not found")
	}

	opts.stateMetaData[balanceIDMetadataKey] = balance.ID
	opts.stateMetaData[currentBalanceNameMetadataKey] = balance.Name

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("send message with keyboard")
		return "", fmt.Errorf("send message with keyboard: %w", err)
	}

	return models.ChooseUpdateBalanceOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose update balance option:",
		InlineKeyboard: updateBalanceOptionsKeyboard,
	})
}

func (h handlerService) handleChooseUpdateBalanceOptionFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseUpdateBalanceOptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	switch opts.message.GetText() {
	case models.BotUpdateBalanceNameCommand:
		return models.EnterBalanceNameFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter updated balance name:")
	case models.BotUpdateBalanceAmountCommand:
		return models.EnterBalanceAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter updated balance amount:")
	case models.BotUpdateBalanceCurrencyCommand:
		currenciesKeyboard, err := h.getCurrenciesKeyboardForBalance(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("get currencies keyboard for balance")
			return "", fmt.Errorf("get currencies keyboard for balance: %w", err)
		}

		return models.EnterBalanceCurrencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:         opts.message.GetChatID(),
			Message:        "Enter balance currency:",
			InlineKeyboard: currenciesKeyboard,
		})
	default:
		return "", fmt.Errorf("received unknown update balance option: %s", opts.message.GetText())
	}
}

func (h handlerService) handleEnterBalanceNameFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceNameFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	text := opts.message.GetText()
	currentBalanceName := opts.stateMetaData[currentBalanceNameMetadataKey].(string)

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
			return models.EnterBalanceNameFlowStep, ErrBalanceAlreadyExists
		}
	}

	_, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: opts.stateMetaData[balanceIDMetadataKey].(string),
		step:      models.EnterBalanceNameFlowStep,
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

	return models.ChooseUpdateBalanceOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Balance name successfully updated!\nPlease choose other update balance option or finish action by canceling it!",
		InlineKeyboard: updateBalanceOptionsKeyboard,
	})
}

func (h handlerService) handleEnterBalanceAmountFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceAmountFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	_, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: opts.stateMetaData[balanceIDMetadataKey].(string),
		step:      models.EnterBalanceAmountFlowStep,
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

	if opts.state.Flow == models.StartFlow {
		currenciesKeyboard, err := h.getCurrenciesKeyboardForBalance(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("get currencies keyboard for balance")
			return "", fmt.Errorf("get currencies keyboard for balance: %w", err)
		}

		return models.EnterBalanceCurrencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:         opts.message.GetChatID(),
			Message:        "Enter balance currency:",
			InlineKeyboard: currenciesKeyboard,
		})
	}

	return models.ChooseUpdateBalanceOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Balance amount successfully updated!\nPlease choose other update balance option or finish action by canceling it!",
		InlineKeyboard: updateBalanceOptionsKeyboard,
	})
}

func (h handlerService) handleEnterBalanceCurrencyFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceCurrencyFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	exists, err := h.stores.Currency.Exists(ctx, ExistsCurrencyFilter{
		ID: opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("check if currency exists in store")
		return "", fmt.Errorf("check if currency exists in store: %w", err)
	}
	if !exists {
		logger.Error().Msg("currency not found")
		return models.EnterBalanceCurrencyFlowStep, ErrCurrencyNotFound
	}

	_, err = h.updateBalance(ctx, updateBalanceOptions{
		balanceID: opts.stateMetaData[balanceIDMetadataKey].(string),
		step:      models.EnterBalanceCurrencyFlowStep,
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

	if opts.state.Flow == models.StartFlow {
		return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), "Initial balance successfully created!")
	}

	return models.ChooseUpdateBalanceOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Balance currency successfully updated!\nPlease choose other update balance option or finish action by canceling it!",
		InlineKeyboard: updateBalanceOptionsKeyboard,
	})
}

type updateBalanceOptions struct {
	balanceID string
	step      models.FlowStep
	data      string
}

func (h handlerService) updateBalance(ctx context.Context, opts updateBalanceOptions) (*models.Balance, error) {
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
	case models.EnterBalanceNameFlowStep:
		balance.Name = opts.data
	case models.EnterBalanceAmountFlowStep:
		price, err := money.NewFromString(opts.data)
		if err != nil {
			logger.Error().Err(err).Msg("convert option amount to money type")
			return nil, ErrInvalidAmountFormat
		}

		balance.Amount = price.StringFixed()
	case models.EnterBalanceCurrencyFlowStep:
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

func (h handlerService) handleDeleteBalanceFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleDeleteBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if len(opts.user.Balances) == 1 {
		return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), "You're not allowed to delete last balance!")
	}

	return models.ConfirmBalanceDeletionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose balance to delete:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h handlerService) handleConfirmBalanceDeletionFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleConfirmBalanceDeletionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceNameMetadataKey] = opts.message.GetText()

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	outputMessage := fmt.Sprintf(
		"Are you sure you want to delete balance %s?\nPlease note that all its operations will be deleted as well.",
		opts.message.GetText(),
	)
	return models.ChooseBalanceFlowStep, h.sendMessageWithConfirmationInlineKeyboard(opts.message.GetChatID(), outputMessage)
}

func (h handlerService) handleChooseBalanceFlowStepForDelete(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForDelete").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	confirmBalanceDeletion, err := strconv.ParseBool(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse callback data to bool")
		return "", fmt.Errorf("parse callback data to bool: %w", err)
	}

	if !confirmBalanceDeletion {
		logger.Info().Msg("user did not confirm balance deletion")
		return models.EndFlowStep, h.notifyCancellationAndShowMenu(opts.message.GetChatID())
	}

	balance := opts.user.GetBalance(opts.stateMetaData[balanceNameMetadataKey].(string))
	if balance == nil {
		logger.Error().Msg("balance for deletion not found")
		return "", fmt.Errorf("balance for deletion not found")
	}
	logger.Debug().Any("balance", balance).Msg("got balance for deletion")

	err = h.stores.Balance.Delete(ctx, balance.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete balance from store")
		return "", fmt.Errorf("delete balance from store: %w", err)
	}

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
	}()

	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), "Balance and all its operations have been deleted!")
}
