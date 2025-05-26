package service

import (
	"context"
	"fmt"
	"time"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/VladPetriv/finance_bot/pkg/worker"
	"github.com/google/uuid"
)

type balanceSubscriptionEngine struct {
	logger *logger.Logger
	stores Stores
	apis   APIs

	operationCreationInterval            time.Duration
	extendingScheduledOperationsInterval time.Duration
}

// NewBalanceSubscriptionEngine creates a new instance of balanceSubscriptionEngine.
func NewBalanceSubscriptionEngine(config *config.Config, logger *logger.Logger, stores Stores, apis APIs) *balanceSubscriptionEngine {
	return &balanceSubscriptionEngine{
		logger:                               logger,
		stores:                               stores,
		apis:                                 apis,
		operationCreationInterval:            config.App.OperationCreationInterval,
		extendingScheduledOperationsInterval: config.App.ExtendingScheduledOperationsInterval,
	}
}

func (b *balanceSubscriptionEngine) ScheduleOperationsCreation(ctx context.Context, balanceSubscription models.BalanceSubscription) {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.ScheduleOperationsCreation").Logger()
	logger.Debug().Any("balanceSubscription", balanceSubscription).Msg("got args")

	maxBillingDates := getMaxBillingDatesFromSubscriptionPeriod(balanceSubscription.Period)
	billingDates := models.CalculateScheduledOperationBillingDates(balanceSubscription.Period, balanceSubscription.StartAt, maxBillingDates)

	err := b.createScheduledOperations(ctx, billingDates, balanceSubscription)
	if err != nil {
		logger.Error().Err(err).Msg("create scheduled operation")
	}
}

func (b *balanceSubscriptionEngine) ExtendScheduledOperations(ctx context.Context) {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.ExtendScheduledOperations").Logger()

	ticker := time.NewTicker(b.extendingScheduledOperationsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("finished extending scheduled operations")
			return
		case <-ticker.C:
			balanceSubscriptions, err := b.stores.BalanceSubscription.List(ctx, ListBalanceSubscriptionFilter{
				SubscriptionsWithLastScheduledOperation: true,
			})
			if err != nil {
				logger.Error().Err(err).Msg("list balance subscriptions")
				continue
			}

			scheduledOperations, err := b.stores.BalanceSubscription.ListScheduledOperation(ctx, ListScheduledOperation{
				BalanceSubscriptionIDs: extractIDs(balanceSubscriptions, func(bs models.BalanceSubscription) string {
					return bs.ID
				}),
			})
			if err != nil {
				logger.Error().Err(err).Msg("list scheduled operations")
				continue
			}

			balanceSubscriptionToScheduledOperation := make(map[string]models.ScheduledOperation)
			for _, operation := range scheduledOperations {
				balanceSubscriptionToScheduledOperation[operation.SubscriptionID] = operation
			}

			for _, balanceSubscription := range balanceSubscriptions {
				scheduledOperation, ok := balanceSubscriptionToScheduledOperation[balanceSubscription.ID]
				if !ok {
					logger.Warn().Msg("could not find scheduled opreation by subscription id")
					continue
				}

				maxBillingDates := getMaxBillingDatesFromSubscriptionPeriod(balanceSubscription.Period) + 1
				billingDates := models.CalculateScheduledOperationBillingDates(balanceSubscription.Period, scheduledOperation.CreationDate, maxBillingDates)

				err := b.createScheduledOperations(ctx, billingDates[1:], balanceSubscription)
				if err != nil {
					logger.Error().Err(err).Msg("create scheduled operation")
				}
			}
		}
	}
}

func extractIDs[T any](items []T, getID func(T) string) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = getID(item)
	}
	return ids
}

func (b *balanceSubscriptionEngine) createScheduledOperations(ctx context.Context, billingDates []time.Time, balanceSubscription models.BalanceSubscription) error {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.createScheduledOperations").Logger()
	logger.Debug().Any("billingDates", billingDates).Any("balanceSubscription", balanceSubscription).Msg("got args")

	for _, billingDate := range billingDates {
		err := b.stores.BalanceSubscription.CreateScheduledOperation(ctx, models.ScheduledOperation{
			ID:             uuid.NewString(),
			SubscriptionID: balanceSubscription.ID,
			CreationDate:   billingDate,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create scheduled operation in store")
			return fmt.Errorf("create scheduled operation in store: %w", err)
		}
	}

	return nil
}

const (
	maxBillingDatesForWeeklySubscription = 13 // Represents a quarter in a week.
	maxBillingDatesForMontlySubscription = 3  // Represents a quarter in a months.
	maxBillingDatesForYearlySubscription = 2  // Represents two year.
)

func getMaxBillingDatesFromSubscriptionPeriod(period models.SubscriptionPeriod) int {
	var maxBillingDates int
	switch period {
	case models.SubscriptionPeriodWeekly:
		maxBillingDates = maxBillingDatesForWeeklySubscription
	case models.SubscriptionPeriodMonthly:
		maxBillingDates = maxBillingDatesForMontlySubscription
	case models.SubscriptionPeriodYearly:
		maxBillingDates = maxBillingDatesForYearlySubscription
	}

	return maxBillingDates
}

func (b *balanceSubscriptionEngine) CreateOperations(ctx context.Context) {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.CreateOperations").Logger()

	ticker := time.NewTicker(b.operationCreationInterval)
	defer ticker.Stop()

	pool := worker.NewPool(5, b.createOperation)
	pool.Start(ctx)
	defer pool.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("finished creation operations")
			return
		case <-ticker.C:
			now := time.Now()
			scheduledOperations, err := b.stores.BalanceSubscription.ListScheduledOperation(ctx, ListScheduledOperation{
				BetweenFilter: &BetweenFilter{
					From: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
					To:   time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC),
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("get scheduled operations from store")
				continue
			}
			logger.Debug().Any("scheduledOperations", scheduledOperations).Msg("got scheduled operations")

			for _, scheduledOperation := range scheduledOperations {
				pool.AddJob(scheduledOperation.ID, scheduledOperation)
			}
		}
	}
}

func (b *balanceSubscriptionEngine) createOperation(ctx context.Context, id string, scheduledOperation models.ScheduledOperation) error {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.createOperation").Logger()
	logger.Debug().Any("scheduledOperation", scheduledOperation).Msg("got args")

	balanceSubscription, err := b.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: scheduledOperation.SubscriptionID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance subscription from store")
		return fmt.Errorf("get balance subscription from store: %w", err)
	}
	if balanceSubscription == nil {
		logger.Warn().Msg("balance subscription during operation not found")
		return ErrBalanceSubscriptionNotFound
	}

	balance, err := b.stores.Balance.Get(ctx, GetBalanceFilter{
		BalanceID: balanceSubscription.BalanceID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return fmt.Errorf("get balance from store: %w", err)
	}
	if balance == nil {
		logger.Warn().Msg("balance during operation not found")
		return ErrBalanceNotFound
	}

	category, err := b.stores.Category.Get(ctx, GetCategoryFilter{
		ID: balanceSubscription.CategoryID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return fmt.Errorf("get category from store: %w", err)
	}
	if category == nil {
		logger.Warn().Msg("category during operation not found")
		return ErrCategoryNotFound
	}

	err = b.stores.Operation.Create(ctx, &models.Operation{
		ID:                    uuid.NewString(),
		BalanceID:             balance.ID,
		CategoryID:            category.ID,
		BalanceSubscriptionID: balanceSubscription.ID,
		Type:                  models.OperationTypeSpending,
		Amount:                balanceSubscription.Amount,
		Description:           fmt.Sprintf("Subscprition payment for: %s", balanceSubscription.Name),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("create operation")
		return fmt.Errorf("create operation: %w", err)
	}

	subscriptionAmount, err := money.NewFromString(balanceSubscription.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse subscription amount")
		return fmt.Errorf("parse subscription amount: %w", err)
	}

	balanceAmount, err := money.NewFromString(balance.Amount)
	if err != nil {
		logger.Error().Err(err).Msg("parse balance amount")
		return fmt.Errorf("parse balance amount: %w", err)
	}

	calculatedBalanceAmount := balanceAmount.Sub(subscriptionAmount)
	logger.Debug().Any("calculatedBalanceAmount", calculatedBalanceAmount).Msg("reduced balance amount with subscription amount")
	balance.Amount = calculatedBalanceAmount.StringFixed()

	err = b.stores.Balance.Update(ctx, balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance")
		return fmt.Errorf("update balance: %w", err)
	}

	err = b.stores.BalanceSubscription.DeleteScheduledOperation(ctx, scheduledOperation.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete scheduled operation")
		return fmt.Errorf("delete scheduled operation: %w", err)
	}

	return nil
}

func (b *balanceSubscriptionEngine) NotifyAboutSubscriptionPayment(ctx context.Context) error {
	// TODO: Delivery this logic as separate feature, since it requires a lot of changes.
	return nil
}
