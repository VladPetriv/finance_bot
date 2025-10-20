package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/VladPetriv/finance_bot/pkg/worker"
	"github.com/google/uuid"
)

type balanceSubscriptionEngine struct {
	logger *logger.Logger
	stores Stores
	apis   APIs

	operationCreationInterval               time.Duration
	extendingScheduledOperationsInterval    time.Duration
	notifyAboutSubscriptionPaymentsInterval time.Duration
}

// NewBalanceSubscriptionEngine creates a new instance of balanceSubscriptionEngine.
func NewBalanceSubscriptionEngine(config *config.Config, logger *logger.Logger, stores Stores, apis APIs) *balanceSubscriptionEngine {
	return &balanceSubscriptionEngine{
		logger:                                  logger,
		stores:                                  stores,
		apis:                                    apis,
		operationCreationInterval:               config.App.OperationCreationInterval,
		extendingScheduledOperationsInterval:    config.App.ExtendingScheduledOperationsInterval,
		notifyAboutSubscriptionPaymentsInterval: config.App.NotifyAboutSubscriptionPaymentsInterval,
	}
}

func (b *balanceSubscriptionEngine) ScheduleOperationsCreation(ctx context.Context, balanceSubscription model.BalanceSubscription) {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.ScheduleOperationsCreation").Logger()
	logger.Debug().Any("balanceSubscription", balanceSubscription).Msg("got args")

	maxBillingDates := getMaxBillingDatesFromSubscriptionPeriod(balanceSubscription.Period)
	billingDates := model.CalculateScheduledOperationBillingDates(balanceSubscription.Period, balanceSubscription.StartAt, maxBillingDates)

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
				BalanceSubscriptionIDs: extractIDs(balanceSubscriptions, func(bs model.BalanceSubscription) string {
					return bs.ID
				}),
			})
			if err != nil {
				logger.Error().Err(err).Msg("list scheduled operations")
				continue
			}

			balanceSubscriptionToScheduledOperation := make(map[string]model.ScheduledOperation)
			for _, operation := range scheduledOperations {
				balanceSubscriptionToScheduledOperation[operation.SubscriptionID] = operation
			}

			for _, balanceSubscription := range balanceSubscriptions {
				scheduledOperation, ok := balanceSubscriptionToScheduledOperation[balanceSubscription.ID]
				if !ok {
					logger.Warn().Msg("could not find scheduled operation by subscription id")
					continue
				}

				maxBillingDates := getMaxBillingDatesFromSubscriptionPeriod(balanceSubscription.Period) + 1
				billingDates := model.CalculateScheduledOperationBillingDates(balanceSubscription.Period, scheduledOperation.CreationDate, maxBillingDates)

				err := b.createScheduledOperations(ctx, billingDates[1:], balanceSubscription)
				if err != nil {
					logger.Error().Err(err).Msg("create scheduled operation")
				}
			}
		}
	}
}

func (b *balanceSubscriptionEngine) createScheduledOperations(ctx context.Context, billingDates []time.Time, balanceSubscription model.BalanceSubscription) error {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.createScheduledOperations").Logger()
	logger.Debug().Any("billingDates", billingDates).Any("balanceSubscription", balanceSubscription).Msg("got args")

	for _, billingDate := range billingDates {
		err := b.stores.BalanceSubscription.CreateScheduledOperation(ctx, model.ScheduledOperation{
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
	maxBillingDatesForWeeklySubscription  = 13 // Represents a quarter in a week.
	maxBillingDatesForMonthlySubscription = 3  // Represents a quarter in a months.
	maxBillingDatesForYearlySubscription  = 2  // Represents two year.
)

func getMaxBillingDatesFromSubscriptionPeriod(period model.SubscriptionPeriod) int {
	var maxBillingDates int
	switch period {
	case model.SubscriptionPeriodWeekly:
		maxBillingDates = maxBillingDatesForWeeklySubscription
	case model.SubscriptionPeriodMonthly:
		maxBillingDates = maxBillingDatesForMonthlySubscription
	case model.SubscriptionPeriodYearly:
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

func (b *balanceSubscriptionEngine) createOperation(ctx context.Context, id string, scheduledOperation model.ScheduledOperation) error {
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

	err = b.stores.Operation.Create(ctx, &model.Operation{
		ID:                    uuid.NewString(),
		BalanceID:             balance.ID,
		CategoryID:            category.ID,
		BalanceSubscriptionID: balanceSubscription.ID,
		Type:                  model.OperationTypeSpending,
		Amount:                balanceSubscription.Amount,
		Description:           fmt.Sprintf("Subscprition payment for: %s", balanceSubscription.Name),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("create operation")
		return fmt.Errorf("create operation: %w", err)
	}

	subscriptionAmount, _ := money.NewFromString(balanceSubscription.Amount)
	balanceAmount, _ := money.NewFromString(balance.Amount)

	calculateSpendingOperation(&balanceAmount, subscriptionAmount)

	balance.Amount = balanceAmount.StringFixed()
	logger.Debug().Any("calculatedBalanceAmount", balance.Amount).Msg("reduced balance amount with subscription amount")

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

func (b *balanceSubscriptionEngine) NotifyAboutSubscriptionPayment(ctx context.Context) {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.NotifyAboutSubscriptionPayment").Logger()

	ticker := time.NewTicker(b.notifyAboutSubscriptionPaymentsInterval)
	defer ticker.Stop()

	pool := worker.NewPool(5, b.notifyUserAboutSubscriptionPayment)
	pool.Start(ctx)
	defer pool.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("finished notifying users about subscription payments")
			return
		case <-ticker.C:
			balanceSubscriptions, err := b.stores.BalanceSubscription.List(ctx, ListBalanceSubscriptionFilter{
				SubscriptionsForUserWhoHasEnabledSubscriptionNotifications: true,
			})
			if err != nil {
				logger.Error().Err(err).Msg("list balance subscriptions")
				continue
			}
			if len(balanceSubscriptions) == 0 {
				logger.Debug().Msg("no balance subscriptions for users who have enabled subscription notifications found")
				continue
			}
			logger.Debug().Any("balanceSubscriptions", balanceSubscriptions).Msg("got balance subscriptions")

			now := time.Now()
			nextDay := now.Day() + 1
			scheduledOperations, err := b.stores.BalanceSubscription.ListScheduledOperation(ctx, ListScheduledOperation{
				BetweenFilter: &BetweenFilter{
					From: time.Date(now.Year(), now.Month(), nextDay, 0, 0, 0, 0, time.UTC),
					To:   time.Date(now.Year(), now.Month(), nextDay, 23, 59, 59, 999999999, time.UTC),
				},
				BalanceSubscriptionIDs: extractIDs(balanceSubscriptions, func(bs model.BalanceSubscription) string {
					return bs.ID
				}),
				NotNotified: true,
			})
			if err != nil {
				logger.Error().Err(err).Msg("get scheduled operations from store")
				continue
			}
			logger.Debug().Any("scheduledOperations", scheduledOperations).Msg("got scheduled operations")

			for _, scheduledOperation := range scheduledOperations {
				balanceSubscriptionIndex := slices.IndexFunc(balanceSubscriptions, func(bs model.BalanceSubscription) bool {
					return bs.ID == scheduledOperation.SubscriptionID
				})
				if balanceSubscriptionIndex == -1 {
					logger.Error().Str("subscriptionID", scheduledOperation.SubscriptionID).Msg("subscription not found")
					continue
				}

				pool.AddJob(scheduledOperation.ID, notifyUserAboutSubscriptionPaymentOptions{
					scheduledOperation:  scheduledOperation,
					balanceSubscription: balanceSubscriptions[balanceSubscriptionIndex],
				})
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

type notifyUserAboutSubscriptionPaymentOptions struct {
	scheduledOperation  model.ScheduledOperation
	balanceSubscription model.BalanceSubscription
}

func (b *balanceSubscriptionEngine) notifyUserAboutSubscriptionPayment(ctx context.Context, id string, opts notifyUserAboutSubscriptionPaymentOptions) error {
	logger := b.logger.With().Str("name", "balanceSubscriptionEngine.notifyUserAboutSubscriptionPayment").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	user, err := b.stores.User.Get(ctx, GetUserFilter{
		BalanceID: opts.balanceSubscription.BalanceID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Warn().Msg("user during operation not found")
		return ErrUserNotFound
	}
	logger.Debug().Any("user", user).Msg("got user")

	balance, err := b.stores.Balance.Get(ctx, GetBalanceFilter{
		BalanceID:       opts.balanceSubscription.BalanceID,
		PreloadCurrency: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return fmt.Errorf("get balance from store: %w", err)
	}
	if balance == nil {
		logger.Warn().Msg("balance during operation not found")
		return ErrBalanceNotFound
	}
	logger.Debug().Any("balance", balance).Msg("got balance")

	err = b.apis.Messenger.SendMessage(user.ChatID, buildSubscriptionNotificationMessage(opts.balanceSubscription, balance))
	if err != nil {
		logger.Error().Err(err).Msg("send message to user")
		return fmt.Errorf("send message to user: %w", err)
	}

	err = b.stores.BalanceSubscription.MarkScheduledOperationAsNotified(ctx, opts.scheduledOperation.ID)
	if err != nil {
		logger.Error().Err(err).Msg("mark scheduled operation as notified")
		return fmt.Errorf("mark scheduled operation as notified: %w", err)
	}

	return nil
}

func buildSubscriptionNotificationMessage(subscription model.BalanceSubscription, balance *model.Balance) string {
	subscriptionAmount, _ := money.NewFromString(subscription.Amount)
	currentBalanceAmount, _ := money.NewFromString(balance.Amount)

	remaningBalanceAmount, _ := money.NewFromString(balance.Amount)
	remaningBalanceAmount.Sub(subscriptionAmount)

	symbol := balance.Currency.Symbol

	var balanceStatus string
	switch {
	case remaningBalanceAmount.Equal(money.Zero):
		balanceStatus = "⚠️ Balance will be exactly zero after payment"
	case remaningBalanceAmount.GreaterThan(money.Zero):
		balanceStatus = fmt.Sprintf("✅ Balance after payment: %s%s", symbol, remaningBalanceAmount.StringFixed())
	case remaningBalanceAmount.LessThan(money.Zero):
		deficitAmount, _ := money.NewFromString(subscription.Amount)
		deficitAmount.Sub(currentBalanceAmount)
		balanceStatus = fmt.Sprintf("❌ Insufficient funds! Need %s%s more", symbol, deficitAmount.StringFixed())
	}

	return fmt.Sprintf(
		"🔔 Your subscription payment \"%s\" charges tomorrow\n\n💰 Amount: %s%s\n📅 Period: %s\n💳 Current balance: %s%s\n%s",
		subscription.Name,
		symbol, subscriptionAmount.StringFixed(),
		subscription.Period,
		symbol, currentBalanceAmount.StringFixed(),
		balanceStatus,
	)
}
