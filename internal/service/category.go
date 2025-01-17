package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/google/uuid"
)

func (h handlerService) HandleCategoryCreate(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleCategoryCreate").Logger()

	var nextStep models.FlowStep
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		if nextStep != "" {
			state.Steps = append(state.Steps, nextStep)
		}
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Any("updatedState", updatedState).Msg("updated state in store")
	}()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username: msg.GetUsername(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Info().Msg("user not found")
		return ErrUserNotFound
	}
	logger.Debug().Any("user", user).Msg("got user from store")

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on create category flow")

	switch currentStep {
	case models.CreateCategoryFlowStep:
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.Message.Chat.ID,
			Message: "Enter category name:",
			Type:    keyboardTypeRow,
			Rows:    rowKeyboardWithCancelButtonOnly,
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterCategoryNameFlowStep
	case models.EnterCategoryNameFlowStep:
		err := h.handleEnterCategoryNameFlowStep(ctx, handleEnterCategoryNameFlowStepOptions{
			userID: user.ID,
			msg:    msg,
		})
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle enter category name flow step")
			return fmt.Errorf("handle enter category name flow step: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	return nil
}

type handleEnterCategoryNameFlowStepOptions struct {
	userID string
	msg    botMessage
}

func (h handlerService) handleEnterCategoryNameFlowStep(ctx context.Context, opts handleEnterCategoryNameFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handleEnterCategoryNameFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		UserID: opts.userID,
		Title:  opts.msg.Message.Text,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return fmt.Errorf("get category from store: %w", err)
	}
	if category != nil {
		logger.Info().Msg("category already exists")
		return ErrCategoryAlreadyExists
	}

	err = h.stores.Category.Create(ctx, &models.Category{
		ID:     uuid.NewString(),
		UserID: opts.userID,
		Title:  opts.msg.Message.Text,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create category in store")
		return fmt.Errorf("create category in store: %w", err)
	}

	err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  opts.msg.GetChatID(),
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
		Message: "Category created!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleCategoryList(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleCategoryList").Logger()

	var nextStep models.FlowStep
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		if nextStep != "" {
			state.Steps = append(state.Steps, nextStep)
		}
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Any("updatedState", updatedState).Msg("updated state in store")
	}()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username: msg.GetUsername(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Info().Msg("user not found")
		return ErrUserNotFound
	}
	logger.Debug().Any("user", user).Msg("got user from store")

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on list categories flow")

	switch currentStep {
	case models.ListCategoriesFlowStep:
		err := h.handleListCategoriesFlowStep(ctx, handleListCategoriesFlowStepOptions{
			userID: user.ID,
			msg:    msg,
		})
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle list categories flow step")
			return fmt.Errorf("handle list categories flow step: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	return nil
}

type handleListCategoriesFlowStepOptions struct {
	userID string
	msg    botMessage
}

func (h handlerService) handleListCategoriesFlowStep(ctx context.Context, opts handleListCategoriesFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handleListCategoriesFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	categories, err := h.listCategories(ctx, opts.userID)
	if err != nil {
		if errs.IsExpected(err) {
			err = h.services.Message.SendMessage(&SendMessageOptions{
				ChatID: opts.msg.GetChatID(),
				Text:   "You don't have any create categories yet!",
			})
			if err != nil {
				logger.Error().Err(err).Msg("send message")
				return fmt.Errorf("send message: %w", err)
			}

			return nil
		}

		logger.Error().Err(err).Msg("handle list categories flow step")
		return fmt.Errorf("handle list categories flow step: %w", err)
	}

	outputMessage := "Categories: \n"

	for i, c := range categories {
		i++
		outputMessage += fmt.Sprintf("%d. %s\n", i, c.Title)
	}
	logger.Debug().Any("outputMessage", outputMessage).Msg("built output message")

	err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  opts.msg.GetChatID(),
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
		Message: outputMessage,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleCategoryUpdate(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleCategoryUpdate").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		if nextStep != "" {
			state.Steps = append(state.Steps, nextStep)
		}
		state.Metedata = stateMetaData
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Any("updatedState", updatedState).Msg("updated state in store")
	}()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username: msg.GetUsername(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Info().Msg("user not found")
		return ErrUserNotFound
	}
	logger.Debug().Any("user", user).Msg("got user from store")

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on update category flow")

	switch currentStep {
	case models.UpdateCategoryFlowStep:
		categories, err := h.listCategories(ctx, user.ID)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("handle list categories flow step")
			return fmt.Errorf("handle list categories flow step: %w", err)
		}

		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.Message.Chat.ID,
			Type:    keyboardTypeRow,
			Rows:    getKeyboardRows(categories, true),
			Message: "Choose category to update:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.ChooseCategoryFlowStep
	case models.ChooseCategoryFlowStep:
		stateMetaData[previousCategoryTitleMetadataKey] = msg.Message.Text

		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Enter updated category name:",
			Type:    keyboardTypeRow,
			Rows: []bot.KeyboardRow{
				{
					Buttons: []string{models.BotCancelCommand},
				},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterUpdatedCategoryNameFlowStep
	case models.EnterUpdatedCategoryNameFlowStep:
		err := h.handleEnterUpdatedCategoryNameFlowStep(ctx, handleEnterUpdatedCategoryNameFlowStepOptions{
			userID:   user.ID,
			metaData: stateMetaData,
			msg:      msg,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle enter uopdated category name flow step")
			return fmt.Errorf("handle enter uopdated category name flow step: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	return nil
}

type handleEnterUpdatedCategoryNameFlowStepOptions struct {
	userID   string
	metaData map[string]any
	msg      botMessage
}

func (h handlerService) handleEnterUpdatedCategoryNameFlowStep(ctx context.Context, opts handleEnterUpdatedCategoryNameFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handleEnterUpdatedCategoryNameFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		UserID: opts.userID,
		Title:  opts.metaData[previousCategoryTitleMetadataKey].(string),
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

	category.Title = opts.msg.Message.Text

	err = h.stores.Category.Update(ctx, category)
	if err != nil {
		logger.Error().Err(err).Msg("update category in store")
		return fmt.Errorf("update category in store: %w", err)
	}

	err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  opts.msg.GetChatID(),
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
		Message: "Category updated!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleCategoryDelete(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleCategoryDelete").Logger()

	var nextStep models.FlowStep
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		if nextStep != "" {
			state.Steps = append(state.Steps, nextStep)
		}
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Any("updatedState", updatedState).Msg("updated state in store")
	}()

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username: msg.GetUsername(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Info().Msg("user not found")
		return ErrUserNotFound
	}
	logger.Debug().Any("user", user).Msg("got user from store")

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on delete category flow")

	switch currentStep {
	case models.DeleteCategoryFlowStep:
		categories, err := h.listCategories(ctx, user.ID)
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				nextStep = models.EndFlowStep
				return err
			}

			logger.Error().Err(err).Msg("handle list categories flow step")
			return fmt.Errorf("handle list categories flow step: %w", err)
		}

		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.Message.Chat.ID,
			Type:    keyboardTypeRow,
			Rows:    getKeyboardRows(categories, true),
			Message: "Choose category to delete:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.ChooseCategoryFlowStep
	case models.ChooseCategoryFlowStep:
		category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
			UserID: user.ID,
			Title:  msg.Message.Text,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get category from store")
			return fmt.Errorf("get category from store: %w", err)
		}
		if category == nil {
			logger.Info().Msg("category not found")
			nextStep = models.EndFlowStep
			return ErrCategoryNotFound
		}
		logger.Debug().Any("category", category).Msg("got category from store")

		err = h.stores.Category.Delete(ctx, category.ID)
		if err != nil {
			logger.Error().Err(err).Msg("delete category in store")
			return fmt.Errorf("delete category in store: %w", err)
		}

		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Type:    keyboardTypeRow,
			Rows:    defaultKeyboardRows,
			Message: "Category deleted!",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	return nil
}

func (h handlerService) listCategories(ctx context.Context, userID string) ([]models.Category, error) {
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
