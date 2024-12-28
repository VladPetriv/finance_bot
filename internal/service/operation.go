package service

import (
	"context"
	"fmt"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

func (h handlerService) HandleEventOperationCreated(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventOperationCreated").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	logger.Debug().Any("stateMetaData", stateMetaData).Msg("got state metadata")

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
		Username:        msg.GetUsername(),
		PreloadBalances: true,
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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step")

	switch currentStep {
	case models.CreateOperationFlowStep:
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Select a balance to view information:",
			Type:    keyboardTypeRow,
			Rows:    convertSliceToKeyboardRows(user.Balances),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep

	case models.ChooseBalanceFlowStep:
		stateMetaData["balanceName"] = msg.Message.Text

		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose operation type:",
			Type:    keyboardTypeInline,
			Rows: []bot.KeyboardRow{
				{
					Buttons: []string{models.BotCreateIncomingOperationCommand, models.BotCreateSpendingOperationCommand},
				},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("create inline keyboard")
			return fmt.Errorf("create inline keyboard: %w", err)
		}

		nextStep = models.ChooseOprationTypeFlowStep

	case models.ChooseOprationTypeFlowStep:
		stateMetaData["operationType"] = models.OperationCommandToOperationType[msg.CallbackQuery.Data]

		categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
			UserID: user.ID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("list categories from store")
			return fmt.Errorf("list categories from store: %w", err)
		}
		if len(categories) == 0 {
			logger.Info().Msg("no categories found")
			return ErrCategoriesNotFound
		}
		logger.Debug().Any("categories", categories).Msg("got categories from store")

		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose operation category:",
			Type:    keyboardTypeRow,
			Rows:    convertSliceToKeyboardRows(categories),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseCategoryFlowStep

	case models.ChooseCategoryFlowStep:
		stateMetaData["categoryTitle"] = msg.Message.Text

		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: msg.GetChatID(),
			Text:   "Enter operation amount:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterOperationAmountFlowStep

	case models.EnterOperationAmountFlowStep:
		err := h.handleEnterOperationAmountFlowStep(ctx, handleEnterOperationAmountFlowStep{
			user:     user,
			metaData: stateMetaData,
			msg:      msg,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle enter operation amount flow step")
			return fmt.Errorf("handle enter operation amount flow step: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	logger.Info().Msg("handled event operation created")
	return nil
}

type handleEnterOperationAmountFlowStep struct {
	user     *models.User
	metaData map[string]any
	msg      botMessage
}

func (h handlerService) handleEnterOperationAmountFlowStep(ctx context.Context, opts handleEnterOperationAmountFlowStep) error {
	logger := h.logger.With().Str("name", "handlerService.handleEnterOperationAmountFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.metaData["balanceName"].(string))
	if balance == nil {
		logger.Info().Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balance", balance).Msg("got balance")

	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		Title: opts.metaData["categoryTitle"].(string),
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

	balanceAmount, err := money.NewFromString(balance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance amount")
		return fmt.Errorf("parse balance amount: %w", err)
	}
	logger.Debug().Any("balanceAmount", balanceAmount).Msg("parsed balance amount")

	operationAmount, err := money.NewFromString(opts.msg.Message.Text)
	if err != nil {
		logger.Error().Err(err).Msg("parse operation amount")
		return ErrInvalidAmountFormat
	}
	logger.Debug().Any("operationAmount", operationAmount).Msg("parsed operation amount")

	operationType := models.OperationType(opts.metaData["operationType"].(string))
	logger.Debug().Any("operationType", operationType).Msg("parsed operation type")

	operation := &models.Operation{
		ID:         uuid.NewString(),
		Type:       operationType,
		Amount:     operationAmount.String(),
		BalanceID:  balance.ID,
		CategoryID: category.ID,
		CreatedAt:  time.Now(),
	}
	logger.Debug().Any("operation", operation).Msg("built operation for create")

	switch operationType {
	case models.OperationTypeIncoming:
		balanceAmount.Inc(operationAmount)
		logger.Debug().Any("balanceAmount", balanceAmount).Msg("increased balance amount with incoming operation")
		balance.Amount = balanceAmount.String()
	case models.OperationTypeSpending:
		calculatedAmount := balanceAmount.Sub(operationAmount)
		logger.Debug().Any("calculatedAmount", calculatedAmount).Msg("decreased balance amount with spending operation")
		balance.Amount = calculatedAmount.String()
	default:
		logger.Error().Any("operationType", operationType).Msg("invalid operation type")
		return fmt.Errorf("received unknown operation type: %s", operationType)
	}

	err = h.stores.Operation.Create(ctx, operation)
	if err != nil {
		logger.Error().Err(err).Msg("create operation in store")
		return fmt.Errorf("create operation in store: %w", err)
	}

	err = h.stores.Balance.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	err = h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: opts.msg.GetChatID(),
		Text:   "Operation created!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleEventGetOperationsHistory(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventGetOperationsHistory").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	logger.Debug().Any("stateMetaData", stateMetaData).Msg("got state metadata")

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
		Username:        msg.GetUsername(),
		PreloadBalances: true,
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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step")

	switch currentStep {
	case models.GetOperationsHistoryFlowStep:
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Select a balance to view information:",
			Type:    keyboardTypeRow,
			Rows:    convertSliceToKeyboardRows(user.Balances),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep
	case models.ChooseBalanceFlowStep:
		stateMetaData["balanceName"] = msg.Message.Text
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Please select a period for operation history!",
			Type:    keyboardTypeRow,
			Rows: []bot.KeyboardRow{
				{
					Buttons: []string{
						string(models.CreationPeriodDay),
						string(models.CreationPeriodWeek),
						string(models.CreationPeriodMonth),
						string(models.CreationPeriodYear),
					},
				},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		nextStep = models.ChooseTimePeriodForOperationsHistoryFlowStep
	case models.ChooseTimePeriodForOperationsHistoryFlowStep:
		err := h.handlerChooseTimePeriodForOperationsHistoryFlowStep(ctx, handlerChooseTimePeriodForOperationsHistoryFlowStepOptions{
			user:     user,
			metaData: stateMetaData,
			msg:      msg,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle choose time period for operations history flow step")
			return fmt.Errorf("handle choose time period for operations history flow step: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	logger.Info().Msg("handled get operations history event")
	return nil
}

type handlerChooseTimePeriodForOperationsHistoryFlowStepOptions struct {
	user     *models.User
	metaData map[string]any
	msg      botMessage
}

func (h handlerService) handlerChooseTimePeriodForOperationsHistoryFlowStep(ctx context.Context, opts handlerChooseTimePeriodForOperationsHistoryFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handlerChooseTimePeriodForOperationsHistoryFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance := opts.user.GetBalance(opts.metaData["balanceName"].(string))
	if balance == nil {
		logger.Info().Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balance", balance).Msg("got balance")

	creationPeriod := models.GetCreationPeriodFromText(opts.msg.Message.Text)
	operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
		BalanceID:      balance.ID,
		CreationPeriod: creationPeriod,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get all operations from store")
		return fmt.Errorf("get all operations from store: %w", err)
	}
	if operations == nil {
		logger.Info().Msg("operations not found")
		return ErrOperationsNotFound
	}

	resultMessage := fmt.Sprintf("Balance Amount: %v%s\nPeriod: %v\n", balance.Amount, balance.Currency, *creationPeriod)

	for _, o := range operations {
		resultMessage += fmt.Sprintf(
			"\nOperation: %s\nCategory: %s\nAmount: %v%s\nCreation date: %v\n- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -",
			o.Type, o.CategoryID, o.Amount, balance.Currency, o.CreatedAt.Format(time.ANSIC),
		)
	}

	err = h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: opts.msg.GetChatID(),
		Text:   resultMessage,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
