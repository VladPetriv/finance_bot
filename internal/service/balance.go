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

func (h *handlerService) handleCreateBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceID := uuid.NewString()
	message := opts.message.GetText()

	if opts.message.GetText() == models.BotCreateBalanceCommand {
		message = ""
	}

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

	switch opts.message.GetText() == models.BotCreateBalanceCommand {
	case true:
		return models.EnterBalanceNameFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   opts.message.GetChatID(),
			Message:  "Please enter balance name:",
			Keyboard: rowKeyboardWithCancelButtonOnly,
		})
	case false:
		return models.EnterBalanceAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Please enter balance amount:")
	default:
		return "", nil
	}
}

func (h handlerService) handleEnterBalanceNameFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceNameFlowStep").Logger()
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

	_, err = h.updateBalance(ctx, updateBalanceOptions{
		balanceID: opts.stateMetaData[balanceIDMetadataKey].(string),
		step:      models.EnterBalanceNameFlowStep,
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

	return models.EnterBalanceAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter balance amount:")
}

func (h handlerService) handleGetBalanceFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleGetBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

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
	if len(categories) == 0 {
		logger.Error().Msg("no categories found")
		return "", ErrCategoriesNotFound
	}

	operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
		BalanceID:      balance.ID,
		CreationPeriod: &models.CreationPeriodCurrentMonth,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list operations from store")
		return "", fmt.Errorf("list operations from store: %w", err)
	}

	outputMessage, err := models.
		NewStatisticsMessageBuilder(balance, operations, categories).
		Build()
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
	opts.stateMetaData[currentBalanceCurrencyMetadataKey] = balance.CurrencyID
	opts.stateMetaData[currentBalanceAmountMetadataKey] = balance.Amount

	err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID: opts.message.GetChatID(),
		Message: `Send '-' if you want to keep the current balance value. Otherwise, send your new value.
Please note: this symbol can be used for any balance value you don't want to change.`,
		Keyboard: rowKeyboardWithCancelButtonOnly,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message with keyboard")
		return "", fmt.Errorf("send message with keyboard: %w", err)
	}

	outputMessage := fmt.Sprintf("Enter new name for balance %s:", balance.Name)
	return models.EnterBalanceNameFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), outputMessage)
}

const usePreviousModelValueFlag = "-"

func (h handlerService) handleEnterBalanceNameFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceNameFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	text := opts.message.GetText()
	currentBalanceName, currentBalanceNameExistsInMetadata := opts.stateMetaData[currentBalanceNameMetadataKey].(string)

	switch text == usePreviousModelValueFlag {
	case true:
		if currentBalanceNameExistsInMetadata {
			text = currentBalanceName
			break
		}
		logger.Warn().Msg("current balance name does not exists in state metadata")

	case false:
		if currentBalanceName != "" && text == currentBalanceName {
			break
		}

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

	return models.EnterBalanceAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter balance amount:")
}

const currenciesPerKeyboardRow = 3

func (h handlerService) handleEnterBalanceAmountFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceAmountFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	text := opts.message.GetText()
	if text == usePreviousModelValueFlag {
		currentBalanceAmount, ok := opts.stateMetaData[currentBalanceAmountMetadataKey].(string)
		if ok {
			text = currentBalanceAmount
		}
	}

	_, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: opts.stateMetaData[balanceIDMetadataKey].(string),
		step:      models.EnterBalanceAmountFlowStep,
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

	currencies, err := h.stores.Currency.List(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("list currencies from store")
		return "", fmt.Errorf("list currencies from store: %w", err)
	}
	if len(currencies) == 0 {
		logger.Info().Msg("currencies not found")
		return "", fmt.Errorf("no currencies found")
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

	return models.EnterBalanceCurrencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Enter balance currency:",
		InlineKeyboard: currenciesKeyboard,
	})
}

func (h handlerService) handleEnterBalanceCurrencyFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceCurrencyFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	text := opts.message.GetText()

	switch text {
	case usePreviousModelValueFlag:
		currentBalanceCurrency, ok := opts.stateMetaData[currentBalanceCurrencyMetadataKey].(string)
		if ok {
			text = currentBalanceCurrency
		}
	default:
		exists, err := h.stores.Currency.Exists(ctx, ExistsCurrencyFilter{
			ID: text,
		})
		if err != nil {
			logger.Error().Err(err).Msg("check if currency exists in store")
			return "", fmt.Errorf("check if currency exists in store: %w", err)
		}
		if !exists {
			logger.Error().Msg("currency not found")
			return models.EnterBalanceCurrencyFlowStep, ErrCurrencyNotFound
		}
	}

	balance, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: opts.stateMetaData[balanceIDMetadataKey].(string),
		step:      models.EnterBalanceCurrencyFlowStep,
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

	outputMessage := fmt.Sprintf(
		"Balance Info:\n - Name: %s\n - Amount: %s\n - Currency: %s",
		balance.Name, balance.Amount, balance.GetCurrency().Name,
	)
	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), outputMessage)
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
