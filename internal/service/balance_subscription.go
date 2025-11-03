package service

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

// Create Balance Subscriptions
func (h *handlerService) handleCreateBalanceSubscriptionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateInitialBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Select source balance for subscription:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForCreateBalanceSubscription(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForCreateBalanceSubscription").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		return model.EndFlowStep, ErrBalanceNotFound
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
		return model.EndFlowStep, ErrCategoriesNotFound
	}

	return model.ChooseCategoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose category for your subscription:",
		UpdatedInlineKeyboard: getInlineKeyboardRows(categories, 3),
	})
}

func (h *handlerService) handleChooseCategoryFlowStepForCreateBalanceSubscription(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
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
		return model.EndFlowStep, ErrCategoryNotFound
	}

	opts.stateMetaData[categoryIDMetadataKey] = category.ID

	return model.EnterBalanceSubscriptionNameFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		InlineMessageID: opts.message.GetInlineMessageID(),
		UpdatedMessage:  "Enter balance subscription name:",
	})
}

func (h *handlerService) handleEnterBalanceSubscriptionNameFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceSubscriptionNameFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData[balanceSubscriptionNameMetadataKey] = opts.message.GetText()

	return model.EnterBalanceSubscriptionAmountFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter balance subscription amount:")
}

func (h *handlerService) handleEnterBalanceSubscriptionAmountFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterBalanceSubscriptionAmountToFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	parsedAmount, err := money.NewFromString(opts.message.GetText())
	if err != nil {
		return "", ErrInvalidAmountFormat
	}

	opts.stateMetaData[balanceSubscriptionAmountMetadataKey] = parsedAmount.String()

	return model.ChooseBalanceSubscriptionFrequencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose subscription frequency:",
		InlineKeyboard: balanceSubscriptionFrequencyKeyboard,
	})
}

func (h *handlerService) handleChooseBalanceSubscriptionFrequencyFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceSubscriptionFrequencyFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	period, err := model.ParseSubscriptionPeriod(opts.message.GetText())
	if err != nil {
		return model.ChooseBalanceSubscriptionFrequencyFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:         opts.message.GetChatID(),
			Message:        "Invalid subscription frequency. Please choose from the options below:",
			InlineKeyboard: balanceSubscriptionFrequencyKeyboard,
		})
	}

	opts.stateMetaData[balanceSubscriptionPeriodMetadataKey] = period

	return model.EnterStartAtDateForBalanceSubscriptionFlowStep, h.apis.Messenger.UpdateMessage(
		UpdateMessageOptions{
			ChatID:          opts.message.GetChatID(),
			MessageID:       opts.message.GetMessageID(),
			InlineMessageID: opts.message.GetInlineMessageID(),
			UpdatedMessage:  "Enter subscription start date and time:\nUse format: DD/MM/YYYY\nExample: 01/01/2025:",
		},
	)
}

const balanceSubscriptionTimeFormat = "02/01/2006"

func (h *handlerService) handleEnterStartAtDateForBalanceSubscriptionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterStartAtDateForBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	parsedStartAtTime, err := time.Parse(balanceSubscriptionTimeFormat, opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse operation date")
		return "", ErrInvalidDateFormat
	}

	period, err := model.ParseSubscriptionPeriod(opts.stateMetaData[balanceSubscriptionPeriodMetadataKey].(string))
	if err != nil {
		return "", fmt.Errorf("parse subscription period: %w", err)
	}

	balanceSubscription := model.BalanceSubscription{
		ID:         uuid.NewString(),
		BalanceID:  opts.stateMetaData[balanceIDMetadataKey].(string),
		CategoryID: opts.stateMetaData[categoryIDMetadataKey].(string),
		Name:       opts.stateMetaData[balanceSubscriptionNameMetadataKey].(string),
		Amount:     opts.stateMetaData[balanceSubscriptionAmountMetadataKey].(string),
		Period:     period,
		StartAt:    parsedStartAtTime,
	}

	err = h.stores.BalanceSubscription.Create(ctx, balanceSubscription)
	if err != nil {
		logger.Error().Err(err).Msg("create balance subscription in store")
		return "", fmt.Errorf("create balance subscription in store: %w", err)
	}

	go h.services.BalanceSubscriptionEngine.ScheduleOperationsCreation(ctx, balanceSubscription)

	return model.EndFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  fmt.Sprintf("Balance subscription successfully created!\n\n%s", balanceSubscription.GetDetails()),
		Keyboard: balanceSubscriptionKeyboardRows,
	})
}

// List Balance Subscriptions
func (h *handlerService) handleListBalanceSubscriptionFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleListBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	opts.stateMetaData[pageMetadataKey] = firstPage

	return model.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Select balance:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForListBalanceSubscriptions(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForListBalanceSubscriptions").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	messageText := opts.message.GetText()

	balanceName, ok := opts.stateMetaData[balanceNameMetadataKey].(string)
	if !ok {
		opts.stateMetaData[balanceNameMetadataKey] = messageText
		opts.stateMetaData[pageMetadataKey] = firstPage

		balanceName = messageText
	}

	balance := opts.user.GetBalance(balanceName)
	if balance == nil {
		return model.EndFlowStep, ErrBalanceNotFound
	}

	if isPaginationNeeded(messageText) {
		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData[pageMetadataKey] = nextPage

		message, keyboard, err := h.getListBalanceSubscriptionsKeyboard(ctx, getBalanceSubscriptionsKeyboardOptions{
			userID:    opts.user.ID,
			balanceID: balance.ID,
			page:      nextPage,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get list balance subscriptions keyboard")
			return "", fmt.Errorf("get list balance subscriptions keyboard: %w", err)
		}

		return model.ChooseBalanceFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedInlineKeyboard:   keyboard,
			UpdatedMessage:          message,
		})
	}

	message, keyboard, err := h.getListBalanceSubscriptionsKeyboard(ctx, getBalanceSubscriptionsKeyboardOptions{
		balanceID: balance.ID,
		page:      firstPage,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get list balance subscriptions keyboard")
		return "", fmt.Errorf("get list balance subscriptions keyboard: %w", err)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedInlineKeyboard:   keyboard,
		UpdatedMessage:          message,
	})
}

// Update Balance Subscriptions
func (h *handlerService) handleUpdateBalanceSubscriptionFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Select balance:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForUpdateBalanceSubscription(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForUpdateBalanceSubscription").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		return model.EndFlowStep, ErrBalanceNotFound
	}

	opts.stateMetaData[balanceIDMetadataKey] = balance.ID
	opts.stateMetaData[pageMetadataKey] = firstPage

	keyboard, err := h.getBalanceSubscriptionsKeyboard(ctx, getBalanceSubscriptionsKeyboardOptions{
		balanceID: balance.ID,
		page:      firstPage,
	})
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get balance subscriptions keyboard")
		return "", fmt.Errorf("get balance subscriptions keyboard: %w", err)
	}

	return model.ChooseBalanceSubscriptionToUpdateFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose balance subscription to update:",
		UpdatedInlineKeyboard: keyboard,
	})
}

func (h *handlerService) handleChooseBalanceSubscriptionToUpdateFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceSubscriptionToUpdateFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	messageText := opts.message.GetText()
	if isPaginationNeeded(messageText) {
		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData[pageMetadataKey] = nextPage

		keyboard, err := h.getBalanceSubscriptionsKeyboard(ctx, getBalanceSubscriptionsKeyboardOptions{
			balanceID: opts.stateMetaData[balanceIDMetadataKey].(string),
			page:      nextPage,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get balance subscriptions keyboard")
			return "", fmt.Errorf("get balance subscriptions keyboard: %w", err)
		}

		return model.ChooseBalanceSubscriptionToUpdateFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        "Choose balance subscription to update:",
			UpdatedInlineKeyboard: keyboard,
		})
	}

	balanceSubscription, err := h.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: messageText,
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

	return model.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose update balance subscription option:",
		UpdatedInlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseUpdateBalanceSubscriptionOptionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseUpdateBalanceSubscriptionOptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got message")

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

	switch opts.message.GetText() {
	case model.BotUpdateBalanceSubscriptionNameCommand:
		return model.EnterBalanceSubscriptionNameFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedMessage:          fmt.Sprintf("Enter updated balance subscription name(Current: `%s`):", balanceSubscription.Name),
		})
	case model.BotUpdateBalanceSubscriptionAmountCommand:
		return model.EnterBalanceSubscriptionAmountFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedMessage:          fmt.Sprintf("Enter updated balance subscription amount(Current: `%s`):", balanceSubscription.Amount),
		})
	case model.BotUpdateBalanceSubscriptionCategoryCommand:
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

		var currentCategory string
		categoriesWithoutAlreadyUsedCategory := slices.DeleteFunc(categories, func(category model.Category) bool {
			currentCategory = category.Title
			return category.ID == balanceSubscription.CategoryID
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
				logger.Error().Err(err).Msg("update message")
				return "", fmt.Errorf("update message: %w", err)
			}

			return model.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
				ChatID:         opts.message.GetChatID(),
				Message:        "Please choose other update balance subscription option or finish action by canceling it!",
				InlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
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
	case model.BotUpdateBalanceSubscriptionPeriodCommand:
		return model.ChooseBalanceSubscriptionFrequencyFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedMessage:          fmt.Sprintf("Select updated balance subscription frequency(Current: `%s`):", balanceSubscription.Period),
			UpdatedInlineKeyboard:   balanceSubscriptionFrequencyKeyboard,
		})

	default:
		return "", fmt.Errorf("received unknown update balance subscription option: %s", opts.message.GetText())
	}
}

func (h *handlerService) handleEnterBalanceSubscriptionNameFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
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

	return model.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		FormatMessageInMarkDown: true,
		Message: fmt.Sprintf(
			"Balance subscription name successfully updated!\nNew name: `%s`\nPlease choose other update operation option or finish action by canceling it!",
			opts.message.GetText(),
		),
		InlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

func (h *handlerService) handleEnterBalanceSubscriptionAmountFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
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

	return model.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		FormatMessageInMarkDown: true,
		Message: fmt.Sprintf(
			"Balance subscription amount successfully updated!\nNew amount: `%s`\nPlease choose other update operation option or finish action by canceling it!",
			parsedAmount.StringFixed(),
		),
		InlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseCategoryFlowStepForBalanceSubscriptionUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
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

	return model.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage: fmt.Sprintf(
			"Balance subscription category successfully updated!\nNew category: `%s`\nPlease choose other update operation option or finish action by canceling it!",
			category.Title,
		),
		UpdatedInlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseBalanceSubscriptionFrequencyFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceSubscriptionFrequencyFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	period, err := model.ParseSubscriptionPeriod(opts.message.GetText())
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

	return model.ChooseUpdateBalanceSubscriptionOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage: fmt.Sprintf(
			"Balance subscription period successfully updated!\nNew period: `%s`\nPlease choose other update operation option or finish action by canceling it!",
			balanceSubscription.Period,
		),
		UpdatedInlineKeyboard: updateBalanceSubscriptionOptionsKeyboard,
	})
}

// Delete Balance Subscriptions
func (h *handlerService) handleDeleteBalanceSubscriptionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleDeleteBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Select balance:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForBalanceSubscriptionDelete(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForBalanceSubscriptionDelete").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.message.GetText())
	if balance == nil {
		return model.EndFlowStep, ErrBalanceNotFound
	}

	opts.stateMetaData[balanceIDMetadataKey] = balance.ID
	opts.stateMetaData[pageMetadataKey] = firstPage

	keyboard, err := h.getBalanceSubscriptionsKeyboard(ctx, getBalanceSubscriptionsKeyboardOptions{
		balanceID: balance.ID,
		page:      firstPage,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance subscriptions keyboard")
		return "", fmt.Errorf("get balance subscriptions keyboard: %w", err)
	}

	return model.ChooseBalanceSubscriptionToDeleteFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose balance subscription to delete:",
		UpdatedInlineKeyboard: keyboard,
	})
}

func (h *handlerService) handleChooseBalanceSubscriptionToDeleteFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceSubscriptionToDeleteFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	messageText := opts.message.GetText()
	if isPaginationNeeded(messageText) {
		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData[pageMetadataKey] = nextPage

		keyboard, err := h.getBalanceSubscriptionsKeyboard(ctx, getBalanceSubscriptionsKeyboardOptions{
			balanceID: opts.stateMetaData[balanceIDMetadataKey].(string),
			page:      nextPage,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get balance subscriptions keyboard")
			return "", fmt.Errorf("get balance subscriptions keyboard: %w", err)
		}

		return model.ChooseBalanceSubscriptionToDeleteFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        "Choose balance subscription to delete:",
			UpdatedInlineKeyboard: keyboard,
		})
	}

	balanceSubscription, err := h.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: messageText,
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

	return model.ConfirmDeleteBalanceSubscriptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedInlineKeyboard: confirmationInlineKeyboardRows,
		UpdatedMessage:        balanceSubscription.GetDeletionMessage(),
	})
}

func (h *handlerService) handleConfirmDeleteBalanceSubscriptionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleConfirmDeleteBalanceSubscriptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	confirmDeletion, err := strconv.ParseBool(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse callback data to bool")
		return "", fmt.Errorf("parse callback data to bool: %w", err)
	}

	if !confirmDeletion {
		logger.Info().Msg("user did not confirm balance subscription deletion")
		return model.EndFlowStep, h.notifyCancellationAndShowKeyboard(opts.message, balanceSubscriptionKeyboardRows)
	}

	err = h.stores.BalanceSubscription.Delete(ctx, opts.stateMetaData[balanceSubscriptionIDMetadataKey].(string))
	if err != nil {
		logger.Error().Err(err).Msg("delete balance subscription")
		return "", fmt.Errorf("delete balance subscription: %w", err)
	}

	return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		UpdatedKeyboard: balanceSubscriptionKeyboardRows,
		UpdatedMessage:  "Balance subscription successfully deleted!",
	})
}
