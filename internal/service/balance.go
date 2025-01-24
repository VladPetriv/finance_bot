package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

func (h handlerService) HandleBalanceCreate(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleBalanceCreate").Logger()

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
		Username:        msg.GetSenderName(),
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
		balance := user.GetBalance(msg.GetText())
		if balance != nil {
			logger.Info().Any("balance", balance).Msg("balance already exists")
			return ErrBalanceAlreadyExists
		}

		step, err := h.processBalanceUpdate(ctx, processBalanceUpdateOptions{
			metadata:    stateMetaData,
			currentStep: currentStep,
			msg:         msg.GetText(),
			chatID:      msg.GetChatID(),
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

	return nil
}

type handleCreateBalanceFlowStepOptions struct {
	user      *models.User
	msg       Message
	isInitial bool
	metadata  map[string]any
}

func (h *handlerService) handleCreateBalanceFlowStep(ctx context.Context, opts handleCreateBalanceFlowStepOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handleCreateBalanceFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceID := uuid.NewString()
	message := opts.msg.GetText()

	if !opts.isInitial {
		message = ""
	}

	err := h.stores.Balance.Create(ctx, &models.Balance{
		ID:     balanceID,
		UserID: opts.user.ID,
		Name:   message,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create balance in store")
		return fmt.Errorf("create balance in store: %w", err)
	}

	options := SendWithKeyboardOptions{
		ChatID: opts.msg.GetChatID(),
	}

	switch opts.isInitial {
	case true:
		options.Message = "Initial balance successfully created!"
		options.Keyboard = defaultKeyboardRows
	case false:
		opts.metadata[balanceIDMetadataKey] = balanceID
		options.Message = "Please enter balance name!"
		options.Keyboard = rowKeyboardWithCancelButtonOnly
	}

	err = h.apis.Messenger.SendWithKeyboard(options)
	if err != nil {
		logger.Error().Err(err).Msg("create keyboard")
		return fmt.Errorf("create keyboard: %w", err)
	}

	return nil
}

func (h handlerService) HandleBalanceUpdate(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleBalanceUpdate").Logger()

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
		Username:        msg.GetSenderName(),
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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on update balance flow")

	switch currentStep {
	case models.UpdateBalanceFlowStep:
		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Choose balance to update:",
			Keyboard: getKeyboardRows(user.Balances, 3, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard with welcome message")
			return fmt.Errorf("create keyboard with welcome message: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep
	case models.ChooseBalanceFlowStep:
		balance := user.GetBalance(msg.GetText())
		if balance == nil {
			logger.Error().Msg("balance not found")
			return fmt.Errorf("balance not found")
		}
		stateMetaData[balanceIDMetadataKey] = balance.ID
		stateMetaData[currentBalanceNameMetadataKey] = balance.Name
		stateMetaData[currentBalanceCurrencyMetadataKey] = balance.Currency
		stateMetaData[currentBalanceAmountMetadataKey] = balance.Amount

		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID: msg.GetChatID(),
			Message: `Send '-' if you want to keep the current balance value. Otherwise, send your new value.
Please note: this symbol can be used for any balance value you don't want to change.`,
			Keyboard: rowKeyboardWithCancelButtonOnly,
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message with keyboard")
			return fmt.Errorf("send message with keyboard: %w", err)
		}

		err = h.apis.Messenger.SendMessage(msg.GetChatID(), fmt.Sprintf("Enter new name for balance %s:", balance.Name))
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		nextStep = models.EnterBalanceNameFlowStep
	case models.EnterBalanceNameFlowStep, models.EnterBalanceAmountFlowStep, models.EnterBalanceCurrencyFlowStep:
		if msg.GetText() != "-" {
			balance := user.GetBalance(msg.GetText())
			if balance != nil {
				logger.Info().Any("balance", balance).Msg("balance already exists")
				return ErrBalanceAlreadyExists
			}
		}

		text := msg.GetText()

		if msg.GetText() == "-" {
			switch currentStep {
			case models.EnterBalanceNameFlowStep:
				text = stateMetaData[currentBalanceNameMetadataKey].(string)
			case models.EnterBalanceAmountFlowStep:
				text = stateMetaData[currentBalanceAmountMetadataKey].(string)
			case models.EnterBalanceCurrencyFlowStep:
				text = stateMetaData[currentBalanceCurrencyMetadataKey].(string)
			}
		}

		step, err := h.processBalanceUpdate(ctx, processBalanceUpdateOptions{
			metadata:    stateMetaData,
			currentStep: currentStep,
			chatID:      msg.GetChatID(),
			msg:         text,
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

	return nil
}

type processBalanceUpdateOptions struct {
	metadata    map[string]any
	currentStep models.FlowStep
	chatID      int
	msg         string
	finalMsg    string
}

func (h handlerService) processBalanceUpdate(ctx context.Context, opts processBalanceUpdateOptions) (models.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.processBalanceUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balance, err := h.updateBalance(ctx, updateBalanceOptions{
		balanceID: opts.metadata[balanceIDMetadataKey].(string),
		step:      opts.currentStep,
		data:      opts.msg,
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
		err := h.apis.Messenger.SendMessage(opts.chatID, "Enter balance amount:")
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return "", fmt.Errorf("send message: %w", err)
		}

		return models.EnterBalanceAmountFlowStep, nil
	case models.EnterBalanceAmountFlowStep:
		err := h.apis.Messenger.SendMessage(opts.chatID, "Enter balance currency:")
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return "", fmt.Errorf("send message: %w", err)
		}

		return models.EnterBalanceCurrencyFlowStep, nil
	case models.EnterBalanceCurrencyFlowStep:
		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID: opts.chatID,
			Message: fmt.Sprintf(
				"%s\nBalance Info:\n - Name: %s\n - Amount: %v\n - Currency: %s",
				opts.finalMsg, balance.Name, balance.Amount, balance.Currency,
			),
			Keyboard: defaultKeyboardRows,
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
			logger.Error().Err(err).Msg("convert option amount to money type")
			return nil, ErrInvalidAmountFormat
		}

		balance.Amount = price.StringFixed()
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

func (h handlerService) HandleBalanceGet(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleBalanceGet").Logger()

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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on get balance flow")

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username:        msg.GetSenderName(),
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
		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Select a balance to view information:",
			Keyboard: getKeyboardRows(user.Balances, 3, true),
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

	return nil
}

func (h handlerService) processGetBalanceInfo(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.getBalanceInfo").Logger()
	logger.Debug().Any("msg", msg).Msg("got args")

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name: msg.GetText(),
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
	err = h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   msg.GetChatID(),
		Message:  fmt.Sprintf("Balance info(%s):\n - Amount: %v\n - Currency: %s", balance.Name, balance.Amount, balance.Currency),
		Keyboard: defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleBalanceDelete(ctx context.Context, msg Message) error {
	logger := h.logger.With().Str("name", "handlerService.HandleBalanceDelete").Logger()

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
		Username:        msg.GetSenderName(),
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
	logger.Debug().Any("currentStep", currentStep).Msg("got current step on delete balance flow")

	switch currentStep {
	case models.DeleteBalanceFlowStep:
		if len(user.Balances) == 1 {
			err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
				ChatID:   msg.GetChatID(),
				Message:  "You're not allowed to delete last balance!",
				Keyboard: defaultKeyboardRows,
			})
			if err != nil {
				logger.Error().Err(err).Msg("create keyboard")
				return fmt.Errorf("create keyboard: %w", err)
			}

			nextStep = models.EndFlowStep
			return nil
		}

		err := h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:   msg.GetChatID(),
			Message:  "Choose balance to delete:",
			Keyboard: getKeyboardRows(user.Balances, 3, true),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard with welcome message")
			return fmt.Errorf("create keyboard with welcome message: %w", err)
		}

		nextStep = models.ConfirmBalanceDeletionFlowStep
	case models.ConfirmBalanceDeletionFlowStep:
		stateMetaData[balanceNameMetadataKey] = msg.GetText()

		err := h.showCancelButton(msg.GetChatID())
		if err != nil {
			logger.Error().Err(err).Msg("show cancel button")
			return fmt.Errorf("show cancel button: %w", err)
		}

		err = h.sendMessageWithConfirmationInlineKeyboard(
			msg.GetChatID(),
			fmt.Sprintf(
				"Are you sure you want to delete balance %s?\nPlease note that all its operations will be deleted as well.",
				msg.GetText(),
			))
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard with welcome message")
			return fmt.Errorf("create keyboard with welcome message: %w", err)
		}

		nextStep = models.ChooseBalanceFlowStep
	case models.ChooseBalanceFlowStep:
		err := h.handleChooseBalanceFlowStepForDeletionFlow(ctx, handleChooseBalanceFlowStepForDeletionFlowOptions{
			user:     user,
			metaData: stateMetaData,
			msg:      msg,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle choose balance flow step for deletion flow")
			return fmt.Errorf("handle choose balance flow step for deletion flow: %w", err)
		}

		nextStep = models.EndFlowStep
	}

	return nil
}

type handleChooseBalanceFlowStepForDeletionFlowOptions struct {
	user     *models.User
	metaData map[string]any
	msg      Message
}

func (h handlerService) handleChooseBalanceFlowStepForDeletionFlow(ctx context.Context, opts handleChooseBalanceFlowStepForDeletionFlowOptions) error {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForDeletionFlow").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	confirmBalanceDeletion, err := strconv.ParseBool(opts.msg.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse callback data to bool")
		return fmt.Errorf("parse callback data to bool: %w", err)
	}

	if !confirmBalanceDeletion {
		logger.Info().Msg("user did not confirm balance deletion")
		return h.notifyCancellationAndShowMenu(opts.msg.GetChatID())
	}

	balance := opts.user.GetBalance(opts.metaData[balanceNameMetadataKey].(string))
	if balance == nil {
		logger.Error().Msg("balance for deletion not found")
		return fmt.Errorf("balance for deletion not found")
	}
	logger.Debug().Any("balance", balance).Msg("got balance for deletion")

	err = h.stores.Balance.Delete(ctx, balance.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete balance from store")
		return fmt.Errorf("delete balance from store: %w", err)
	}

	// Run in separate goroutine to not block the main thread and respond to the user as soon as possible.
	go func() {
		balanceOperations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
			BalanceID: balance.ID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("list operations from store")

			return
		}

		for _, operation := range balanceOperations {
			err := h.stores.Operation.Delete(ctx, operation.ID)
			if err != nil {
				logger.Error().Err(err).Str("operationID", operation.ID).Msg("delete operation from store")

				continue
			}
		}
	}()

	err = h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:   opts.msg.GetChatID(),
		Message:  "Balance and all its operations have been deleted!",
		Keyboard: defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
