package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

func (h handlerService) HandleEventBalanceCreated(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventBalanceCreated").Logger()
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
	case models.CreateInitialBalanceFlowStep:
		err := h.handleCreateBalanceFlowStep(ctx, handleCreateBalanceFlowStepOptions{
			msg:       msg,
			user:      user,
			isInitial: true,
		})
		if err != nil {
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}
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
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}
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
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

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
	logger.Debug().Any("opts", opts).Msg("got args")

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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on create balance flow")

	switch currentStep {
	case models.UpdateBalanceFlowStep:
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose balance to update:",
			Type:    keyboardTypeRow,
			Rows:    getKeyboardRows(user.Balances, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard with welcome message")
			return fmt.Errorf("create keyboard with welcome message: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep

	case models.ChooseBalanceFlowStep:
		balance := user.GetBalance(msg.Message.Text)
		if balance == nil {
			logger.Info().Msg("balance not found")
			return fmt.Errorf("balance not found")
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
			if errs.IsExpected(err) {
				logger.Info().Err(err).Msg(err.Error())
				return err
			}

			logger.Error().Err(err).Msg("process balance update")
			return fmt.Errorf("process balance update: %w", err)
		}

		nextStep = step
	}

	logger.Info().Msg("handled event balance updated")
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
	logger.Debug().Any("opts", opts).Msg("got args")

	if opts.msg.Message.Text != "" {
		balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
			Name: opts.msg.Message.Text,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get balance from store")
			return "", fmt.Errorf("get balance from store: %w", err)
		}
		if balance != nil {
			logger.Info().Any("balance", balance).Msg("balance already exists")
			return "", ErrBalanceAlreadyExists
		}
	}

	balanceID := opts.metadata["balanceID"].(string)
	logger.Debug().Any("balanceID", balanceID).Msg("got balance id")

	balance, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: balanceID,
		step:      opts.currentStep,
		data:      opts.msg.Message.Text,
	})
	if err != nil {
		if errs.IsExpected(err) {
			logger.Info().Err(err).Msg(err.Error())
			return "", err
		}

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
		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID: opts.msg.GetChatID(),
			Type:   keyboardTypeRow,
			Rows:   defaultKeyboardRows,
			Message: fmt.Sprintf(
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
			logger.Info().Err(err).Msg("convert option amount to money type")
			return nil, ErrInvalidAmountFormat
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

func (h handlerService) HandleEventGetBalance(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.HandleEventGetBalance").Logger()
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
	case models.GetBalanceFlowStep:
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Select a balance to view information:",
			Type:    keyboardTypeRow,
			Rows:    getKeyboardRows(user.Balances, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard with welcome message")
			return fmt.Errorf("create keyboard with welcome message: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep
	case models.ChooseBalanceFlowStep:
		err := h.processGetBalanceInfo(ctx, msg)
		if err != nil {
			logger.Error().Err(err).Msg("process get balance info")
			return fmt.Errorf("process get balance info: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	logger.Info().Msg("handled event get balance")
	return nil
}

func (h handlerService) processGetBalanceInfo(ctx context.Context, msg botMessage) error {
	logger := h.logger.With().Str("name", "handlerService.getBalanceInfo").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name: msg.Message.Text,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return fmt.Errorf("get balance from store: %w", err)
	}
	if balance == nil {
		logger.Error().Msg("balance not found")
		return fmt.Errorf("balance not found")
	}
	logger.Debug().Any("balance", balance).Msg("got balance from store")

	// TODO: In the feature it would be great to add some statistics about operations on this balance.
	err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID: msg.GetChatID(),
		Type:   keyboardTypeRow,
		Rows:   defaultKeyboardRows,
		Message: fmt.Sprintf(
			"Balance info(%s):\n - Amount: %v\n - Currency: %s",
			balance.Name, balance.Amount, balance.Currency,
		),
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
