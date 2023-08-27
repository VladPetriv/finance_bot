package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/VladPetriv/finance_bot/pkg/money"
)

type operationService struct {
	logger         *logger.Logger
	OperationStore OperationStore
	balanceStore   BalanceStore
	categoryStore  CategoryStore
}

// NewOperation returns new instance of operation service.
func NewOperation(logger *logger.Logger, operationStore OperationStore, balanceStore BalanceStore, categoryStore CategoryStore) *operationService {
	return &operationService{
		logger:         logger,
		balanceStore:   balanceStore,
		OperationStore: operationStore,
		categoryStore:  categoryStore,
	}
}

func (o operationService) CreateOperation(ctx context.Context, opts CreateOperationOptions) error {
	logger := o.logger
	logger.Debug().Interface("opts", opts).Msg("got args")

	balance, err := o.balanceStore.Get(ctx, opts.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return fmt.Errorf("get balance from store: %w", err)
	}
	if balance == nil {
		logger.Info().Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Interface("balance", balance).Msg("got balance")

	if opts.Operation.Type == models.OperationTypeIncoming {
		err := o.handleIncomingOperationType(ctx, balance, opts.Operation)
		if err != nil {
			return err
		}
		logger.Info().Msg("handled incoming operation type")
	}

	if opts.Operation.Type == models.OperationTypeSpending {
		err := o.handleSpendingOperationType(ctx, balance, opts.Operation)
		if err != nil {
			return err
		}
		logger.Info().Msg("handled spending operation type")
	}

	logger.Info().Msg("operation created")
	return nil
}

func (o operationService) handleIncomingOperationType(ctx context.Context, balance *models.Balance, operation *models.Operation) error {
	logger := o.logger

	logger.Debug().
		Interface("balance", balance).
		Interface("operation", operation).
		Msg("got args")

	operationAmount, err := money.NewFromString(operation.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("convert operation amount to money type")
		return ErrInvalidAmountFormat
	}
	logger.Debug().Interface("operationAmount", operationAmount).Msg("converted operation amount to money type")

	balanceAmount, err := money.NewFromString(balance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("convert balance amount to money type")
		return fmt.Errorf("convert balance amount to money type: %w", err)
	}
	logger.Debug().Interface("balanceAmount", balanceAmount).Msg("converted balance amount to money type")

	balanceAmount.Inc(operationAmount)
	logger.Debug().Interface("balanceAmount", balanceAmount).Msg("increased balance amount with incoming operation")

	balance.Amount = balanceAmount.String()
	err = o.balanceStore.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	err = o.OperationStore.Update(ctx, operation.ID, operation)
	if err != nil {
		logger.Error().Err(err).Msg("create operation in store")
		return fmt.Errorf("create operation in store: %w", err)
	}

	logger.Info().Msg("handled case with incoming operation")
	return nil
}

func (o operationService) handleSpendingOperationType(ctx context.Context, balance *models.Balance, operation *models.Operation) error {
	logger := o.logger
	logger.Debug().
		Interface("balance", balance).
		Interface("operation", operation).
		Msg("got args")

	operationAmount, err := money.NewFromString(operation.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("convert operation amount to money type")
		return ErrInvalidAmountFormat
	}
	logger.Debug().Interface("operationAmount", operationAmount).Msg("converted operation amount to money type")

	balanceAmount, err := money.NewFromString(balance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("convert balance amount to money type")
		return fmt.Errorf("convert balance amount to money type: %w", err)
	}
	logger.Debug().Interface("balanceAmount", balanceAmount).Msg("converted balance amount to money type")

	calculatedAmount := balanceAmount.Sub(operationAmount)
	logger.Debug().Interface("calculatedAmount", calculatedAmount).Msg("decrease balance amount with spending operation")

	balance.Amount = calculatedAmount.String()
	err = o.balanceStore.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	err = o.OperationStore.Update(ctx, operation.ID, operation)
	if err != nil {
		logger.Error().Err(err).Msg("create operation in store")
		return fmt.Errorf("create operation in store: %w", err)
	}

	logger.Info().Msg("handled case with spending operation")
	return nil
}
