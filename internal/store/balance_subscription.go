package store

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
)

type balanceSubscriptionStore struct {
	db *database.PostgreSQL
}

// NewBalanceSubscriptionStore creates a new instance of balance subscription store.
func NewBalanceSubscriptionStore(db *database.PostgreSQL) *balanceSubscriptionStore {
	return &balanceSubscriptionStore{
		db: db,
	}
}

func (b *balanceSubscriptionStore) Create(ctx context.Context, subscription models.BalanceSubscription) error {
	_, err := b.db.DB.ExecContext(
		ctx,
		`INSERT INTO
			balance_subscriptions (id, balance_id, category_id, name, period, start_at)
        VALUES
        	($1, $2, $3, $4, $5, $6);`,
		subscription.ID, subscription.BalanceID, subscription.CategoryID, subscription.Name, subscription.Period, subscription.StartAt,
	)
	return err
}

func (b *balanceSubscriptionStore) List(ctx context.Context, filter service.ListBalanceSubscriptionFilter) ([]models.BalanceSubscription, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "balance_id", "category_id", "name", "period", "start_at", "created_at", "updated_at").
		From("balance_subscriptions")

	if filter.UserID != "" {
		stmt = stmt.Where(sq.Eq{"user_id": filter.UserID})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list balance subscriptions query: %w", err)
	}

	var subscriptions []models.BalanceSubscription
	err = b.db.DB.SelectContext(ctx, &subscriptions, query, args...)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (b *balanceSubscriptionStore) Update(ctx context.Context, subscription *models.BalanceSubscription) error {
	_, err := b.db.DB.ExecContext(
		ctx,
		`
		UPDATE balance_subscriptions
		SET
			category_id = $1,
			name = $2,
			period = $3,
			start_at = $4,
			updated_at = NOW()
		WHERE
			id = $5;`,
		subscription.CategoryID, subscription.Name, subscription.Period, subscription.StartAt, subscription.ID,
	)
	return err
}

func (b *balanceSubscriptionStore) Delete(ctx context.Context, subscriptionID string) error {
	_, err := b.db.DB.ExecContext(ctx, "DELETE FROM balance_subscriptions WHERE id = $1", subscriptionID)
	return err
}
