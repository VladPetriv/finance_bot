package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/google/uuid"
)

func (h handlerService) HandleEventCategoryCreated(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventCategoryCreated").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on create balance flow")

	switch currentStep {
	case models.CreateCategoryFlowStep:
		err := h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: msg.Message.Chat.ID,
			Text:   "Enter category name:",
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

	logger.Info().Msg("handled create category event")
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
		Title: opts.msg.Message.Text,
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

func (h handlerService) HandleEventListCategories(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventListCategories").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on create balance flow")

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

	logger.Info().Msg("handled list categories event")
	return nil
}

type handleListCategoriesFlowStepOptions struct {
	userID string
	msg    botMessage
}

func (h handlerService) handleListCategoriesFlowStep(ctx context.Context, opts handleListCategoriesFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handleListCategoriesFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: opts.userID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list categories from store")
		return fmt.Errorf("list categories from store: %w", err)
	}
	if len(categories) == 0 {
		logger.Info().Msg("categories not found")

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
