package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
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

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		return models.EndFlowStep, ErrBalanceNotFound
	}

	opts.stateMetaData[balanceIDMetadataKey] = balance.ID

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

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		Title: opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return "", fmt.Errorf("get category from store: %w", err)
	}
	if category == nil {
		return models.EndFlowStep, ErrCategoryNotFound
	}

	opts.stateMetaData[categoryIDMetadataKey] = category.ID
	return models.EnterBalanceSubscriptionNameFlowStep, h.showCancelButton(opts.message.GetChatID(), "Enter balance subscription name:")
}

func (h *handlerService) handleEnterBalanceSubscriptionNameFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceSubscriptionNameFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceSubscriptionNameMetadataKey] = opts.message.GetText()

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

	return models.ChooseBalanceSubscriptionFrequencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Choose subscription frequency:",
		Keyboard: balanceSubscriptionFrequencyKeyboard,
	})
}

func (h *handlerService) handleChooseBalanceSubscriptionFrequencyFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceSubscriptionFrequencyFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	period, err := models.ParseSubscriptionPeriod(opts.message.GetText())
	if err != nil {
		return models.ChooseBalanceSubscriptionFrequencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   opts.message.GetChatID(),
			Message:  "Invalid subscription frequency. Please choose from the options below:",
			Keyboard: balanceSubscriptionFrequencyKeyboard,
		})
	}

	opts.stateMetaData[balanceSubscriptionPeriodMetadataKey] = period

	return models.EnterStartAtDateForBalanceSubscriptionFlowStep, h.showCancelButton(
		opts.message.GetChatID(),
		"Enter subscription start date and time:\nUse format: DD/MM/YYYY HH:MM\nExample: 01/01/2025 12:00:",
	)
}

func (h *handlerService) handleEnterStartAtDateForBalanceSubscriptionFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterStartAtDateForBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	parsedStartAtTime, err := time.Parse("02/01/2006 15:04", opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation date")
		return "", ErrInvalidDateFormat
	}

	period, err := models.ParseSubscriptionPeriod(opts.stateMetaData[balanceSubscriptionPeriodMetadataKey].(string))
	if err != nil {
		return "", fmt.Errorf("parse subscription period: %w", err)
	}

	err = h.stores.BalanceSubscription.Create(ctx, models.BalanceSubscription{
		ID:         uuid.NewString(),
		BalanceID:  opts.stateMetaData[balanceIDMetadataKey].(string),
		CategoryID: opts.stateMetaData[categoryIDMetadataKey].(string),
		Name:       opts.stateMetaData[balanceSubscriptionNameMetadataKey].(string),
		Amount:     opts.stateMetaData[balanceSubscriptionAmountMetadataKey].(string),
		Period:     period,
		StartAt:    parsedStartAtTime,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create balance subscription in store")
		return "", fmt.Errorf("create balance subscription in store: %w", err)
	}

	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), "Balance subscription successfully created!")
}

func (h *handlerService) handleListBalanceSubscriptionsFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleListBalanceSubscriptionsFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Select balance:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForListBalanceSubscriptions(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForListBalanceSubscriptions").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		return models.EndFlowStep, ErrBalanceNotFound
	}

	balanceSubscriptions, err := h.stores.BalanceSubscription.List(ctx, ListBalanceSubscriptionFilter{
		BalanceID: balance.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list balance subscriptions")
		return "", fmt.Errorf("list balance subscriptions: %w", err)
	}
	if len(balanceSubscriptions) == 0 {
		return models.EndFlowStep, ErrNoBalanceSubscriptionsFound
	}

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: opts.user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list categories")
		return "", fmt.Errorf("list categories: %w", err)
	}
	if len(categories) == 0 {
		return models.EndFlowStep, ErrCategoriesNotFound
	}

	var outputMessage string
	for _, subscription := range balanceSubscriptions {
		var categoryTitle string
		categoryIndex := slices.IndexFunc(categories, func(category models.Category) bool {
			return category.ID == subscription.CategoryID
		})
		if categoryIndex != -1 {
			categoryTitle = categories[categoryIndex].Title
		}

		outputMessage += fmt.Sprintf(
			"Title: %s\nAmount: %s\nFrequency: %s\nCategory Title: %s\n--------\n",
			subscription.Name, subscription.Amount, subscription.Period, categoryTitle,
		)
	}

	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), outputMessage)
}

func (h *handlerService) handleUpdateBalanceSubscriptionFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Select balance:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForUpdateBalanceSubscription(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForUpdateBalanceSubscription").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		return models.EndFlowStep, ErrBalanceNotFound
	}

	
}
