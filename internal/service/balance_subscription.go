package service

import (
	"context"
	"fmt"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/money"
)

func (h *handlerService) handleCreateBalanceSubscriptionFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateInitialBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Select source balance for subscription:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForCreateBalanceSubscription(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForCreateBalanceSubscription").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceNameMetadataKey] = opts.message.GetText()

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: opts.user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list user categories from store")
		return "", fmt.Errorf("list user categories from store: %w", err)
	}
	if len(categories) == 0 {
		return models.EndFlowStep, ErrCategoriesNotFound
	}

	return models.ChooseCategoryFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose category for your subscription:",
		Keyboard: getKeyboardRows(categories, 3, true),
	})
}

func (h *handlerService) handleChooseCategoryFlowStepForCreateBalanceSubscription(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseCategoryFlowStepForCreateBalanceSubscription").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[categoryTitleMetadataKey] = opts.message.GetText()
	return models.EnterBalanceSubscriptionNameFlowStep, h.showCancelButton(opts.message.GetChatID(), "Enter balance subscription name:")
}

func (h *handlerService) handleEnterBalanceSubscriptionNameFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceSubscriptionNameFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceNameMetadataKey] = opts.message.GetText()

	return models.EnterBalanceSubscriptionAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter balance subscription amount:")
}

func (h *handlerService) handleEnterBalanceSubscriptionAmountFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceSubscriptionAmountToFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	parsedAmount, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		return "", ErrInvalidAmountFormat
	}

	opts.stateMetaData[balanceSubscriptionAmountMetadataKey] = parsedAmount.String()

	return models.EndFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:  opts.message.GetChatID(),
		Message: "Enter subscription start date and time:\nUse format: DD/MM/YYYY HH:MM\nExample: 01/01/2025 12:00",
		Keyboard: []KeyboardRow{
			{Buttons: []string{string(models.SubscriptionPeriodWeekly), string(models.SubscriptionPeriodMonthly), string(models.SubscriptionPeriodYearly)}},
			{Buttons: []string{models.BotCancelCommand}},
		},
	})
}

func (h *handlerService) handleEnterStartAtDateForBalanceSubscriptionFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterStartAtDateForBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	parsedStartAtTime, err := time.Parse("02/01/2006 15:04", opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation date")
		return "", ErrInvalidDateFormat
	}

	fmt.Printf("parsedStartAtTime: %v\n", parsedStartAtTime)

	// TODO: create subscription in store

	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), "Balance subscription successfully created!")
}
