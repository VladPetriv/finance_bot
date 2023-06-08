package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type balanceService struct {
	logger       *logger.Logger
	balanceStore BalanceStore
}

var _ BalanceService = (*balanceService)(nil)

// NewBalance creates a new instance of Balance service.
func NewBalance(logger *logger.Logger, balanceStore BalanceStore) *balanceService {
	return &balanceService{
		logger:       logger,
		balanceStore: balanceStore,
	}
}

// TODO: Do I need a service method with calling only 1 method?
func (b balanceService) CreateBalance(ctx context.Context, balance *models.Balance) error {
	logger := b.logger

	err := b.balanceStore.Create(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("create balance")
		return fmt.Errorf("create balance: %w", err)
	}

	logger.Info().Msg("balance successfully created")
	return nil
}
