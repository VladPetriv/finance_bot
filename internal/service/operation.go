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

	balance, err := o.balanceStore.Get(ctx, opts.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("get balance by user id")
		return fmt.Errorf("get balance by user id: %w", err)
	}
	if balance == nil {
		logger.Info().Str("userID", opts.UserID).Msg("balance not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Interface("balance", balance).Msg("got balance")

	if opts.Operation.Type == models.OperationTypeIncoming {
		err := o.handleIncomingOperationType(ctx, balance, opts.Operation)
		if err != nil {
			return err
		}
	}

	if opts.Operation.Type == models.OperationTypeSpending {
		err := o.handleSpendingOperationType(ctx, balance, opts.Operation)
		if err != nil {
			return err
		}
	}

	logger.Info().Msg("operation successfully created")
	return nil
}

func (o operationService) handleIncomingOperationType(ctx context.Context, balance *models.Balance, operation *models.Operation) error {
	logger := o.logger

	operationAmount, err := money.NewFromString(operation.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("convert operation string amount to money type")
		// TODO: return custom error
		return fmt.Errorf("should be customer error here")
	}

	balanceAmount, err := money.NewFromString(balance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("convert balance string amount to money type")
		return fmt.Errorf("convert balance string amount to money type: %w", err)
	}

	balanceAmount.Inc(operationAmount)

	balance.Amount = balanceAmount.String()
	err = o.balanceStore.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in storage")
		return fmt.Errorf("update balance in storage: %w", err)
	}

	err = o.OperationStore.Update(ctx, operation.ID, operation)
	if err != nil {
		logger.Error().Err(err).Msg("create operation in storage")
		return fmt.Errorf("create operation in storage: %w", err)
	}

	logger.Info().Msg("successfully handle case with incoming operation")
	return nil
}

func (o operationService) handleSpendingOperationType(ctx context.Context, balance *models.Balance, operation *models.Operation) error {
	logger := o.logger

	operationAmount, err := money.NewFromString(operation.Amount)
	if err != nil {
		logger.Info().Err(err).Msg("convert operation string amount to money type")
		return ErrInvalidAmountFormat
	}

	balanceAmount, err := money.NewFromString(balance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("convert balance string amount to money type")
		return fmt.Errorf("convert balance string amount to money type: %w", err)
	}

	calculatedAmount := balanceAmount.Sub(operationAmount)

	balance.Amount = calculatedAmount.String()
	err = o.balanceStore.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in storage")
		return fmt.Errorf("update balance in storage: %w", err)
	}

	err = o.OperationStore.Update(ctx, operation.ID, operation)
	if err != nil {
		logger.Error().Err(err).Msg("create operation in storage")
		return fmt.Errorf("create operation in storage: %w", err)
	}

	logger.Info().Msg("successfully handle case with incoming operation")
	return nil
}
