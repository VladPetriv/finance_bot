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

// NewBalance returns new instance of balance service.
func NewBalance(logger *logger.Logger, balanceStore BalanceStore) *balanceService {
	return &balanceService{
		logger:       logger,
		balanceStore: balanceStore,
	}
}

func (b balanceService) GetBalanceInfo(ctx context.Context, userID string) (*models.Balance, error) {
	logger := b.logger

	balance, err := b.balanceStore.Get(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Msg("get balance by user id")
		return nil, fmt.Errorf("get balance by user id: %w", err)
	}
	if balance == nil {
		logger.Info().Str("userID", userID).Msg("balance not found")
		return nil, ErrBalanceNotFound
	}
	logger.Debug().Interface("balance", balance).Msg("got balance")

	logger.Info().Msg("successfully got balance info")
	return balance, nil
}
