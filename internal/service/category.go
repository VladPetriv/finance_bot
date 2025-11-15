package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/google/uuid"
)

func (h handlerService) handleCreateCategoryFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateCategoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "Enter category name:")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.EnterCategoryNameFlowStep, nil
}

func (h handlerService) handleEnterCategoryNameFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterCategoryNameFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		UserID: opts.user.ID,
		Title:  opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return "", fmt.Errorf("get category from store: %w", err)
	}
	if category != nil {
		logger.Info().Msg("category already exists")
		return "", ErrCategoryAlreadyExists
	}

	err = h.stores.Category.Create(ctx, &model.Category{
		ID:     uuid.NewString(),
		UserID: opts.user.ID,
		Title:  opts.message.GetText(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("create category in store")
		return "", fmt.Errorf("create category in store: %w", err)
	}

	return model.EndFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  "Category created!",
		Keyboard: categoryKeyboardRows,
	})
}

func (h handlerService) handleListCategoriesFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleListCategoriesFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	categories, err := h.listCategories(ctx, opts.user.ID)
	if err != nil {
		if errs.IsExpected(err) {
			return model.EndFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "You don't have any created categories yet!")
		}

		logger.Error().Err(err).Msg("handle list categories flow step")
		return "", fmt.Errorf("handle list categories flow step: %w", err)
	}

	outputMessage := "Categories: \n"

	for i, c := range categories {
		i++
		outputMessage += fmt.Sprintf("%d. %s\n", i, c.Title)
	}
	logger.Debug().Any("outputMessage", outputMessage).Msg("built output message")

	return model.EndFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.message.GetChatID(),
		Message:  outputMessage,
		Keyboard: categoryKeyboardRows,
	})
}

func (h handlerService) handleUpdateCategoryFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateCategoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	categories, err := h.listCategories(ctx, opts.user.ID)
	if err != nil {
		if errs.IsExpected(err) {
			return model.EndFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "You don't have any created categories yet!")
		}

		logger.Error().Err(err).Msg("handle list categories flow step")
		return "", fmt.Errorf("handle list categories flow step: %w", err)
	}

	err = h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseCategoryFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose category to update:",
		InlineKeyboard: getInlineKeyboardRows(categories, 3),
	})
}

func (h handlerService) handleChooseCategoryFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseCategoryFlowStep").Logger()
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

	opts.stateMetaData.Add(model.PreviousCategoryIDMetadataKey, category.ID)
	return model.EnterUpdatedCategoryNameFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage:          fmt.Sprintf("Enter updated category name(Current: `%s`)", category.Title),
	})
}

func (h handlerService) handleEnterUpdatedCategoryNameFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterUpdatedCategoryNameFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	previousCategoryID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.PreviousCategoryIDMetadataKey)
	if !ok {
		logger.Error().Msg("previous category id not found in metadata")
		return "", fmt.Errorf("previous category id not found in metadata")
	}

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		ID: previousCategoryID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return "", fmt.Errorf("get category from store: %w", err)
	}
	if category == nil {
		logger.Info().Msg("category not found")
		return "", ErrCategoryNotFound
	}
	logger.Debug().Any("category", category).Msg("got category from store")

	category.Title = opts.message.GetText()

	err = h.stores.Category.Update(ctx, category)
	if err != nil {
		logger.Error().Err(err).Msg("update category in store")
		return "", fmt.Errorf("update category in store: %w", err)
	}

	return model.EndFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		FormatMessageInMarkDown: true,
		Message:                 fmt.Sprintf("Category updated!\nNew category: `%s`", category.Title),
		Keyboard:                categoryKeyboardRows,
	})
}

func (h handlerService) handleDeleteCategoryFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleDeleteCategoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	categories, err := h.listCategories(ctx, opts.user.ID)
	if err != nil {
		if errs.IsExpected(err) {
			return model.EndFlowStep, err
		}

		logger.Error().Err(err).Msg("list categories")
		return "", fmt.Errorf("list categories: %w", err)
	}

	err = h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseCategoryFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose category to delete:",
		InlineKeyboard: getInlineKeyboardRows(categories, 3),
	})
}

func (h handlerService) handleChooseCategoryFlowStepForDelete(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseCategoryFlowStepForDelete").Logger()
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
		return model.EndFlowStep, ErrCategoryNotFound
	}
	logger.Debug().Any("category", category).Msg("got category from store")

	err = h.stores.Category.Delete(ctx, category.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete category in store")
		return "", fmt.Errorf("delete category in store: %w", err)
	}

	return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		UpdatedKeyboard: categoryKeyboardRows,
		UpdatedMessage:  "Category deleted successfully!",
	})
}

func (h handlerService) listCategories(ctx context.Context, userID string) ([]model.Category, error) {
	logger := h.logger.With().Str("name", "handlerService.listCategories").Logger()
	logger.Debug().Any("userID", userID).Msg("got args")

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: userID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list categories from store")
		return nil, fmt.Errorf("list categories from store: %w", err)
	}
	if len(categories) == 0 {
		logger.Info().Msg("categories not found")
		return nil, ErrCategoriesNotFound
	}

	logger.Info().Any("categories", categories).Msg("got categories from store")
	return categories, nil
}
