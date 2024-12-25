package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

func (h handlerService) HandleEventBalanceCreated(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventBalanceCreated").Logger()
	logger.Debug().Interface("msg", msg).Msg("got args")

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	logger.Debug().Any("stateMetaData", stateMetaData).Msg("got state metadata")

	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		state.Steps = append(state.Steps, nextStep)
		state.Metedata = stateMetaData
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Interface("updatedState", updatedState).Msg("updated state in store")
	}()

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on create balance flow")

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

	switch currentStep {
	case models.CreateInitialBalanceFlowStep:
		err := h.handleCreateBalanceFlowStep(ctx, handleCreateBalanceFlowStepOptions{
			msg:       msg,
			user:      user,
			isInitial: true,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle enter balance name flow step")
			return fmt.Errorf("handle enter balance name flow step: %w", err)
		}

		nextStep = models.EndFlowStep

	case models.CreateBalanceFlowStep:
		err := h.handleCreateBalanceFlowStep(ctx, handleCreateBalanceFlowStepOptions{
			msg:      msg,
			user:     user,
			metadata: stateMetaData,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle enter balance name flow step")
			return fmt.Errorf("handle enter balance name flow step: %w", err)
		}

		nextStep = models.EnterBalanceNameFlowStep

	case models.EnterBalanceNameFlowStep, models.EnterBalanceAmountFlowStep, models.EnterBalanceCurrencyFlowStep:
		step, err := h.processBalanceUpdate(ctx, processBalanceUpdateOptions{
			metadata:    stateMetaData,
			currentStep: currentStep,
			msg:         msg,
			finalMsg:    "Balance successfully created!",
		})
		if err != nil {
			logger.Error().Err(err).Msg("process balance update")
			return fmt.Errorf("process balance update: %w", err)
		}

		nextStep = step
	}

	logger.Info().Msg("handled event balance created")
	return nil
}

type handleCreateBalanceFlowStepOptions struct {
	user      *models.User
	msg       botMessage
	isInitial bool
	metadata  map[string]any
}

func (h *handlerService) handleCreateBalanceFlowStep(ctx context.Context, opts handleCreateBalanceFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handleCreateBalanceFlowStep").Logger()
	logger.Debug().Interface("opts", opts).Msg("got args")

	balanceID := uuid.NewString()

	if !opts.isInitial {
		opts.msg.Message.Text = ""
	}

	err := h.stores.Balance.Create(ctx, &models.Balance{
		ID:     balanceID,
		UserID: opts.user.ID,
		Name:   opts.msg.Message.Text,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create balance in store")
		return fmt.Errorf("create balance in store: %w", err)
	}

	var (
		outputText          string
		sendDefaultKeyboard bool
	)

	switch opts.isInitial {
	case true:
		outputText = "Initial balance successfully created!"
		sendDefaultKeyboard = true
	case false:
		outputText = "Please enter balance name!"
		opts.metadata["balanceID"] = balanceID
	}

	if sendDefaultKeyboard {
		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  opts.msg.GetChatID(),
			Message: outputText,
			Type:    keyboardTypeRow,
			Rows:    defaultKeyboardRows,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard with welcome message")
			return fmt.Errorf("create keyboard with welcome message: %w", err)
		}

		return nil
	}

	err = h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: opts.msg.GetChatID(),
		Text:   outputText,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleEventBalanceUpdated(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventBalanceUpdated").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

	var nextStep models.FlowStep
	stateMetaData := ctx.Value(contextFieldNameState).(*models.State).Metedata
	logger.Debug().Any("stateMetaData", stateMetaData).Msg("got state metadata")

	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		state.Steps = append(state.Steps, nextStep)
		state.Metedata = stateMetaData
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Interface("updatedState", updatedState).Msg("updated state in store")
	}()

	currentStep := ctx.Value(contextFieldNameState).(*models.State).GetCurrentStep()
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on create balance flow")

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

	switch currentStep {
	case models.UpdateBalanceFlowStep:
		err := h.handleUpdateBalanceFlowStep(handleUpdateBalanceFlowStepOptions{
			msg:      msg,
			user:     user,
			metadata: stateMetaData,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle update balance flow step")
			return fmt.Errorf("handle update balance flow step: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep

	case models.ChooseBalanceFlowStep:
		balance := user.GetBalance(msg.Message.Text)
		if balance == nil {
			logger.Info().Msg("balance not found")
			return ErrBalanceNotFound
		}

		err := h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: msg.GetChatID(),
			Text:   fmt.Sprintf("Enter new name for balance %s:", balance.Name),
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		stateMetaData["balanceID"] = balance.ID
		nextStep = models.EnterBalanceNameFlowStep

	case models.EnterBalanceNameFlowStep, models.EnterBalanceAmountFlowStep, models.EnterBalanceCurrencyFlowStep:
		step, err := h.processBalanceUpdate(ctx, processBalanceUpdateOptions{
			metadata:    stateMetaData,
			currentStep: currentStep,
			msg:         msg,
			finalMsg:    "Balance successfully updated!",
		})
		if err != nil {
			logger.Error().Err(err).Msg("process balance update")
			return fmt.Errorf("process balance update: %w", err)
		}

		nextStep = step
	}

	logger.Info().Msg("handled event balance updated")
	return nil
}

type handleUpdateBalanceFlowStepOptions struct {
	user     *models.User
	msg      botMessage
	metadata map[string]any
}

const maxBalancesPerRow = 3

func (h handlerService) handleUpdateBalanceFlowStep(opts handleUpdateBalanceFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateBalanceFlowStep").Logger()
	logger.Debug().Interface("opts", opts).Msg("got args")

	keyboardRows := make([]bot.KeyboardRow, 0)

	var currentRow bot.KeyboardRow
	for i, balance := range opts.user.Balances {
		currentRow.Buttons = append(currentRow.Buttons, balance.Name)

		// When row is full or we're at the last balance item, append row
		if len(currentRow.Buttons) == maxBalancesPerRow || i == len(opts.user.Balances)-1 {
			keyboardRows = append(keyboardRows, currentRow)
			currentRow = bot.KeyboardRow{} // Reset current row
		}
	}

	err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  opts.msg.GetChatID(),
		Message: "Choose balance to update:",
		Type:    keyboardTypeRow,
		Rows:    keyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create keyboard with welcome message")
		return fmt.Errorf("create keyboard with welcome message: %w", err)
	}

	logger.Info().Msg("handled update balance flow step")
	return nil
}

type processBalanceUpdateOptions struct {
	metadata    map[string]any
	currentStep models.FlowStep
	msg         botMessage
	finalMsg    string
}

func (h handlerService) processBalanceUpdate(ctx context.Context, opts processBalanceUpdateOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.processBalanceUpdate").Logger()
	logger.Debug().Interface("opts", opts).Msg("got args")

	balanceID := opts.metadata["balanceID"].(string)
	logger.Debug().Any("balanceID", balanceID).Msg("got balance id")

	balance, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: balanceID,
		step:      opts.currentStep,
		data:      opts.msg.Message.Text,
	})
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return "", fmt.Errorf("update balance in store: %w", err)
	}

	switch opts.currentStep {
	case models.EnterBalanceNameFlowStep:
		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: opts.msg.GetChatID(),
			Text:   "Enter balance amount:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return "", fmt.Errorf("send message: %w", err)
		}

		return models.EnterBalanceAmountFlowStep, nil

	case models.EnterBalanceAmountFlowStep:
		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: opts.msg.GetChatID(),
			Text:   "Enter balance currency:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return "", fmt.Errorf("send message: %w", err)
		}

		return models.EnterBalanceCurrencyFlowStep, nil
	case models.EnterBalanceCurrencyFlowStep:
		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: opts.msg.GetChatID(),
			Text: fmt.Sprintf(
				"%s\nBalance Info:\n - Name: %s\n - Amount: %v\n - Currency: %s",
				opts.finalMsg, balance.Name, balance.Amount, balance.Currency,
			),
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return "", fmt.Errorf("send message: %w", err)
		}

		return models.EndFlowStep, nil
	}

	return "", nil
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
		logger.Info().Msg("balance not found")
		return nil, ErrBalanceNotFound
	}
	logger.Debug().Any("balance", balance).Msg("got balance from store")

	switch opts.step {
	case models.EnterBalanceNameFlowStep:
		balance.Name = opts.data

	case models.EnterBalanceAmountFlowStep:
		price, err := money.NewFromString(opts.data)
		if err != nil {
			logger.Error().Err(err).Msg("convert option amount to money type")
			return nil, fmt.Errorf("convert option amount to money type: %w", err)
		}

		balance.Amount = price.String()
	case models.EnterBalanceCurrencyFlowStep:
		balance.Currency = opts.data
	}

	err = h.stores.Balance.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return nil, fmt.Errorf("update balance in store: %w", err)
	}

	return balance, nil
}
