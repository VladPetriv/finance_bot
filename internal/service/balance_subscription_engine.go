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

	err := b.createScheduledOperations(ctx, balanceSubscription.StartAt, balanceSubscription)
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

			now := time.Now()
			for _, balanceSubscription := range balanceSubscriptions {
				startAt := time.Date(
					now.Year(),
					now.Month(),
					balanceSubscription.StartAt.Day(),
					balanceSubscription.StartAt.Hour(),
					balanceSubscription.StartAt.Minute(),
					balanceSubscription.StartAt.Second(),
					balanceSubscription.StartAt.Nanosecond(),
					balanceSubscription.StartAt.Location(),
				)

				err := b.createScheduledOperations(ctx, startAt, balanceSubscription)
				if err != nil {
					logger.Error().Err(err).Msg("create scheduled operation")
				}
			}
		}
	}
}

func (b *balanceSubscriptionEngine) createScheduledOperations(ctx context.Context, startAt time.Time, balanceSubscription models.BalanceSubscription) error {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.createScheduledOperations").Logger()
	logger.Debug().Time("startAt", startAt).Any("balanceSubscription", balanceSubscription)

	maxBillingDates := getMaxBillingDatesFromSubscriptionPeriod(balanceSubscription.Period)
	billingDates := models.CalculateScheduledOperationBillingDates(balanceSubscription.Period, startAt, maxBillingDates)

	for _, billingDate := range billingDates {
		err := b.stores.BalanceSubscription.CreateScheduledOperationCreation(ctx, models.ScheduledOperationCreation{
			ID:             uuid.NewString(),
			SubscriptionID: balanceSubscription.ID,
			CreationDate:   billingDate,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create scheduled operation creation in store")
			return fmt.Errorf("create scheduled operation creation in store: %w", err)
		}
	}

	return nil
}

const (
	maxBillingDatesForWeeklySubscription = 13 // Represents a quarter in a week.
	maxBillingDatesForMontlySubscription = 3  // Represents a quarter in a months.
	maxBillingDatesForYearlySubscription = 1  // Represents one year.
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
			scheduledOperations, err := b.stores.BalanceSubscription.ListScheduledOperationCreation(ctx, ListScheduledOperationCreation{
				BetweenFilter: &BetweenFilter{
					From: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
					To:   time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()),
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

func (b *balanceSubscriptionEngine) createOperation(ctx context.Context, id string, scheduledOperationCreation models.ScheduledOperationCreation) error {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.createOperation").Logger()
	logger.Debug().Any("scheduledOperationCreation", scheduledOperationCreation).Msg("got args")

	balanceSubscription, err := b.stores.BalanceSubscription.Get(ctx, GetBalanceSubscriptionFilter{
		ID: scheduledOperationCreation.SubscriptionID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance subscription from store")
		return fmt.Errorf("get balance subscription from store: %w", err)
	}
	if balanceSubscription == nil {
		logger.Warn().Msg("balance subscription during operation creation not found")
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
		logger.Warn().Msg("balance during operation creation not found")
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
		logger.Warn().Msg("category during operation creation not found")
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

	err = b.stores.BalanceSubscription.DeleteScheduledOperationCreation(ctx, scheduledOperationCreation.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete scheduled operation creation")
		return fmt.Errorf("delete scheduled operation creation: %w", err)
	}

	return nil
}

func (b *balanceSubscriptionEngine) NotifyAboutSubscriptionPayment(ctx context.Context) error {
	// TODO: Delivery this logic as separate feature, since it requires a lot of changes.
	return nil
}
