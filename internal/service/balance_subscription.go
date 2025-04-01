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

	parsedStartAtTime, err := time.Parse(defaultTimeFormat, opts.message.GetText())
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

func (h *handlerService) handleListBalanceSubscriptionFlowStep(_ context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleListBalanceSubscriptionFlowStep").Logger()
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

	opts.stateMetaData[balanceIDMetadataKey] = balance.ID

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return models.ChooseBalanceSubscriptionToUpdateFlowStep, h.sendBalanceSubscriptions(ctx, sendBalanceSubscriptionsOptions{
		balanceID:                      balance.ID,
		chatID:                         opts.message.GetChatID(),
		includeLastShowedOperationDate: true,
		stateMetadata:                  opts.stateMetaData,
	})
}

func (h *handlerService) handleChooseBalanceSubscriptionToUpdateFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceSubscriptionToUpdateFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if opts.message.GetText() == models.BotShowMoreCommand {
		return models.ChooseBalanceSubscriptionToUpdateFlowStep, h.sendBalanceSubscriptions(ctx, sendBalanceSubscriptionsOptions{
			balanceID:                      opts.stateMetaData[balanceIDMetadataKey].(string),
			chatID:                         opts.message.GetChatID(),
			includeLastShowedOperationDate: true,
			stateMetadata:                  opts.stateMetaData,
		})
	}

	balanceSubscription, err := h.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance subscription from store")
		return "", fmt.Errorf("get balance subscription from store: %w", err)
	}
	if balanceSubscription == nil {
		logger.Info().Msg("balance subscription not found")
		return "", ErrBalanceSubscriptionNotFound
	}

	opts.stateMetaData[balanceSubscriptionIDMetadataKey] = balanceSubscription.ID

	err = h.showCancelButton(opts.message.GetChatID(), balanceSubscription.GetDetails())
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return models.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose update balance subscription option:",
		InlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseUpdateBalanceSubscriptionOptionFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseUpdateBalanceSubscriptionOptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got message")

	switch opts.message.GetText() {
	case models.BotUpdateBalanceSubscriptionNameCommand:
		return models.EnterBalanceSubscriptionNameFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter updated balance subscription name:")
	case models.BotUpdateBalanceSubscriptionAmountCommand:
		return models.EnterBalanceSubscriptionAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter updated balance subscription amount:")
	case models.BotUpdateBalanceSubscriptionCategoryCommand:
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

		balanceSubscription, err := h.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
			ID: opts.stateMetaData[balanceSubscriptionIDMetadataKey].(string),
		})
		if err != nil {
			logger.Error().Err(err).Msg("get balance subscription from store")
			return "", fmt.Errorf("get balance subscription from store: %w", err)
		}
		if balanceSubscription == nil {
			logger.Info().Msg("balance subscription not found")
			return "", ErrBalanceSubscriptionNotFound
		}

		categoriesWithoutAlreadyUsedCategory := slices.DeleteFunc(categories, func(category models.Category) bool {
			return category.ID == balanceSubscription.CategoryID
		})
		if len(categoriesWithoutAlreadyUsedCategory) == 0 {
			return "", ErrNotEnoughCategories
		}

		return models.ChooseCategoryFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   opts.message.GetChatID(),
			Message:  "Choose updated operation category:",
			Keyboard: getKeyboardRows(categoriesWithoutAlreadyUsedCategory, 3, true),
		})
	case models.BotUpdateBalanceSubscriptionPeriodCommand:
		return models.ChooseBalanceSubscriptionFrequencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   opts.message.GetChatID(),
			Message:  "Select updated balance subscription frequency:",
			Keyboard: balanceSubscriptionFrequencyKeyboard,
		})

	}

	return "", nil
}

func (h *handlerService) handleEnterBalanceSubscriptionNameFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceSubscriptionNameFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceSubscription, err := h.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: opts.stateMetaData[balanceSubscriptionIDMetadataKey].(string),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance subscription from store")
		return "", fmt.Errorf("get balance subscription from store: %w", err)
	}
	if balanceSubscription == nil {
		logger.Info().Msg("balance subscription not found")
		return "", ErrBalanceSubscriptionNotFound
	}

	balanceSubscription.Name = opts.message.GetText()

	err = h.stores.BalanceSubscription.Update(ctx, balanceSubscription)
	if err != nil {
		logger.Error().Err(err).Msg("update balance subscription")
		return "", fmt.Errorf("update balance subscription: %w", err)
	}

	err = h.apis.Messenger.SendMessage(opts.message.GetChatID(), balanceSubscription.GetDetails())
	if err != nil {
		logger.Error().Err(err).Msg("send message with balance subscription details")
		return "", fmt.Errorf("send message with balance subscription details: %w", err)
	}

	return models.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Balance subscription name successfully updated!\nPlease choose other update operation option or finish action by canceling it!",
		InlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

func (h *handlerService) handleEnterBalanceSubscriptionAmountFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceSubscriptionAmountFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	parsedAmount, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse input amount")
		return "", ErrInvalidAmountFormat
	}

	balanceSubscription, err := h.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: opts.stateMetaData[balanceSubscriptionIDMetadataKey].(string),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance subscription from store")
		return "", fmt.Errorf("get balance subscription from store: %w", err)
	}
	if balanceSubscription == nil {
		logger.Info().Msg("balance subscription not found")
		return "", ErrBalanceSubscriptionNotFound
	}

	balanceSubscription.Amount = parsedAmount.StringFixed()

	err = h.stores.BalanceSubscription.Update(ctx, balanceSubscription)
	if err != nil {
		logger.Error().Err(err).Msg("update balance subscription")
		return "", fmt.Errorf("update balance subscription: %w", err)
	}

	err = h.apis.Messenger.SendMessage(opts.message.GetChatID(), balanceSubscription.GetDetails())
	if err != nil {
		logger.Error().Err(err).Msg("send message with balance subscription details")
		return "", fmt.Errorf("send message with balance subscription details: %w", err)
	}

	return models.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Balance subscription amount successfully updated!\nPlease choose other update operation option or finish action by canceling it!",
		InlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseCategoryFlowStepForBalanceSubscriptionUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseCategoryFlowStepForBalanceSubscriptionUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		Title: opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return "", fmt.Errorf("get category from store: %w", err)
	}
	if category == nil {
		logger.Info().Msg("category not found")
		return "", ErrCategoryNotFound
	}

	balanceSubscription, err := h.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: opts.stateMetaData[balanceSubscriptionIDMetadataKey].(string),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance subscription from store")
		return "", fmt.Errorf("get balance subscription from store: %w", err)
	}
	if balanceSubscription == nil {
		logger.Info().Msg("balance subscription not found")
		return "", ErrBalanceSubscriptionNotFound
	}

	balanceSubscription.CategoryID = category.ID

	err = h.stores.BalanceSubscription.Update(ctx, balanceSubscription)
	if err != nil {
		logger.Error().Err(err).Msg("update balance subscription")
		return "", fmt.Errorf("update balance subscription: %w", err)
	}

	err = h.showCancelButton(opts.message.GetChatID(), balanceSubscription.GetDetails())
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return models.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Balance subscription category successfully updated!\nPlease choose other update operation option or finish action by canceling it!",
		InlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseBalanceSubscriptionFrequencyFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceSubscriptionFrequencyFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	period, err := models.ParseSubscriptionPeriod(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse subscriptions period from input")
		return "", fmt.Errorf("parse subscription period: %w", err)
	}

	balanceSubscription, err := h.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: opts.stateMetaData[balanceSubscriptionIDMetadataKey].(string),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance subscription from store")
		return "", fmt.Errorf("get balance subscription from store: %w", err)
	}
	if balanceSubscription == nil {
		logger.Info().Msg("balance subscription not found")
		return "", ErrBalanceSubscriptionNotFound
	}

	balanceSubscription.Period = period

	err = h.stores.BalanceSubscription.Update(ctx, balanceSubscription)
	if err != nil {
		logger.Error().Err(err).Msg("update balance subscription")
		return "", fmt.Errorf("update balance subscription: %w", err)
	}

	err = h.showCancelButton(opts.message.GetChatID(), balanceSubscription.GetDetails())
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return models.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Balance subscription period successfully updated!\nPlease choose other update operation option or finish action by canceling it!",
		InlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

func (h *handlerService) handleDeleteBalanceSubscriptionFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleDeleteBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	return models.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Select balance:",
		Keyboard: getKeyboardRows(opts.user.Balances, 3, true),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForBalanceSubscriptionDelete(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForBalanceSubscriptionDelete").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		return models.EndFlowStep, ErrBalanceNotFound
	}

	opts.stateMetaData[balanceIDMetadataKey] = balance.ID

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return models.ChooseBalanceSubscriptionToDeleteFlowStep, h.sendBalanceSubscriptions(ctx, sendBalanceSubscriptionsOptions{
		balanceID:                      balance.ID,
		chatID:                         opts.message.GetChatID(),
		includeLastShowedOperationDate: true,
		stateMetadata:                  opts.stateMetaData,
	})
}

func (h *handlerService) handleChooseBalanceSubscriptionToDeleteFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceSubscriptionToDeleteFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if opts.message.GetText() == models.BotShowMoreCommand {
		return models.ChooseBalanceSubscriptionToDeleteFlowStep, h.sendBalanceSubscriptions(ctx, sendBalanceSubscriptionsOptions{
			balanceID:                      opts.stateMetaData[balanceIDMetadataKey].(string),
			chatID:                         opts.message.GetChatID(),
			includeLastShowedOperationDate: true,
			stateMetadata:                  opts.stateMetaData,
		})
	}

	balanceSubscription, err := h.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance subscription from store")
		return "", fmt.Errorf("get balance subscription from store: %w", err)
	}
	if balanceSubscription == nil {
		logger.Info().Msg("balance subscription not found")
		return "", ErrBalanceSubscriptionNotFound
	}

	opts.stateMetaData[balanceSubscriptionIDMetadataKey] = balanceSubscription.ID

	return models.ConfirmDeleteBalanceSubscriptionFlowStep, h.sendMessageWithConfirmationInlineKeyboard(
		opts.message.GetChatID(),
		balanceSubscription.GetDeletionMessage(),
	)
}

func (h *handlerService) handleConfirmDeleteBalanceSubscriptionFlowStep(ctx context.Context, opts flowProcessingOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleConfirmDeleteBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	confirmDeletion, err := strconv.ParseBool(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse callback data to bool")
		return "", fmt.Errorf("parse callback data to bool: %w", err)
	}

	if !confirmDeletion {
		logger.Info().Msg("user did not confirm balance subscription deletion")
		return models.EndFlowStep, h.notifyCancellationAndShowMenu(opts.message.GetChatID())
	}

	err = h.stores.BalanceSubscription.Delete(ctx, opts.stateMetaData[balanceSubscriptionIDMetadataKey].(string))
	if err != nil {
		logger.Error().Err(err).Msg("delete balance subscription")
		return "", fmt.Errorf("delete balance subscription: %w", err)
	}

	return models.EndFlowStep, h.sendMessageWithDefaultKeyboard(opts.message.GetChatID(), "Balance subscription successfully deleted!")
}

type sendBalanceSubscriptionsOptions struct {
	balanceID                      string
	chatID                         int
	includeLastShowedOperationDate bool
	stateMetadata                  map[string]any
}

const balanceSubscriptionsPerMessage = 10

func (h *handlerService) sendBalanceSubscriptions(ctx context.Context, opts sendBalanceSubscriptionsOptions) error {
	logger := h.logger.With().Str("name", "handlerService.sendBalanceSubscriptions").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	filter := ListBalanceSubscriptionFilter{
		BalanceID: opts.balanceID,
	}

	if opts.includeLastShowedOperationDate {
		lastOperationTime, ok := opts.stateMetadata[lastBalanceSubscriptionDateMetadataKey].(string)
		if ok {
			parsedTime, err := time.Parse("2006-01-02 15:04:05", lastOperationTime)
			if err != nil {
				logger.Error().Err(err).Msg("parse last operation date")
				return fmt.Errorf("parse last operation date: %w", err)
			}
			filter.CreatedAtLessThan = parsedTime
		}
	}

	balanceSubscriptionsCount, err := h.stores.BalanceSubscription.Count(ctx, filter)
	if err != nil {
		logger.Error().Err(err).Msg("count balance subscriptions in store")
		return fmt.Errorf("count balance subscriptions in store: %w", err)
	}
	if balanceSubscriptionsCount == 0 {
		logger.Info().Any("balanceID", opts.balanceID).Msg("balance subscriptions not found")
		return ErrNoBalanceSubscriptionsFound
	}

	filter.OrderByCreatedAtDesc = true
	filter.Limit = balanceSubscriptionsPerMessage
	balanceSubscriptions, err := h.stores.BalanceSubscription.List(ctx, filter)
	if err != nil {
		logger.Error().Err(err).Msg("list balance subscriptions from store")
		return fmt.Errorf("list balance subscriptions from store: %w", err)
	}
	if len(balanceSubscriptions) == 0 {
		logger.Info().Any("balanceID", opts.balanceID).Msg("balance subscriptions not found")
		return ErrNoBalanceSubscriptionsFound
	}

	// Store the timestamp of the most recent balance subscription in metadata.
	// This timestamp serves as a pagination cursor, enabling the retrieval
	// of subsequent operations in chronological order.
	opts.stateMetadata[lastBalanceSubscriptionDateMetadataKey] = balanceSubscriptions[len(balanceSubscriptions)-1].CreatedAt

	err = h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.chatID,
		Message:        "Select balance subscription to update:",
		InlineKeyboard: convertModelToInlineKeyboardRowsWithPagination(balanceSubscriptionsCount, balanceSubscriptions, operationsPerMessage),
	})
	if err != nil {
		logger.Error().Err(err).Msg("create inline keyboard")
		return fmt.Errorf("create inline keyboard: %w", err)
	}

	return nil
}
