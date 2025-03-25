package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
)

type balanceSubscriptionStore struct {
	db *database.PostgreSQL
}

// NewBalanceSubscription creates a new instance of balance subscription store.
func NewBalanceSubscription(db *database.PostgreSQL) *balanceSubscriptionStore {
	return &balanceSubscriptionStore{
		db: db,
	}
}

func (b *balanceSubscriptionStore) Create(ctx context.Context, subscription models.BalanceSubscription) error {
	_, err := b.db.DB.ExecContext(
		ctx,
		`INSERT INTO
			balance_subscriptions (id, balance_id, category_id, name, amount, period, start_at)
    	VALUES
     		($1, $2, $3, $4, $5, $6, $7);`,
		subscription.ID, subscription.BalanceID, subscription.CategoryID, subscription.Name, subscription.Amount, subscription.Period, subscription.StartAt,
	)
	return err
}

func (b *balanceSubscriptionStore) Get(ctx context.Context, filter service.GetBalanceSubscriptionFilter) (*models.BalanceSubscription, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "balance_id", "category_id", "name", "amount", "period", "start_at", "created_at", "updated_at").
		From("balance_subscriptions")

	if filter.ID != "" {
		stmt = stmt.Where(sq.Eq{"id": filter.ID})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get balance subscription query: %w", err)
	}

	var subscription models.BalanceSubscription
	err = b.db.DB.GetContext(ctx, &subscription, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &subscription, nil
}

func (b *balanceSubscriptionStore) Count(ctx context.Context, filter service.ListBalanceSubscriptionFilter) (int, error) {
	stmt := applyListBalanceSubscriptionFilter(applyListBalanceSubscriptionOptions{countQuery: true}, filter)
	query, args, err := stmt.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count balance subscriptions query: %w", err)
	}

	var count int64
	err = b.db.DB.GetContext(ctx, &count, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return int(count), nil
}

func (b *balanceSubscriptionStore) List(ctx context.Context, filter service.ListBalanceSubscriptionFilter) ([]models.BalanceSubscription, error) {
	stmt := applyListBalanceSubscriptionFilter(applyListBalanceSubscriptionOptions{listQuery: true}, filter)

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

type applyListBalanceSubscriptionOptions struct {
	listQuery  bool
	countQuery bool
}

func applyListBalanceSubscriptionFilter(options applyListBalanceSubscriptionOptions, filter service.ListBalanceSubscriptionFilter) *sq.SelectBuilder {
	var expectedColumns []string
	if options.countQuery {
		expectedColumns = []string{"COUNT(id)"}
	}

	if options.listQuery {
		expectedColumns = []string{"id", "balance_id", "category_id", "name", "amount", "period", "start_at", "created_at", "updated_at"}
	}

	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select(expectedColumns...).
		From("balance_subscriptions")

	if filter.BalanceID != "" {
		stmt = stmt.Where(sq.Eq{"balance_id": filter.BalanceID})
	}

	if !filter.CreatedAtLessThan.IsZero() {
		stmt = stmt.Where(sq.Lt{"created_at": filter.CreatedAtLessThan})
	}

	if filter.Limit != 0 {
		stmt = stmt.Limit(uint64(filter.Limit))
	}

	if filter.OrderByCreatedAtDesc {
		stmt = stmt.GroupBy("id", "balance_id", "category_id", "name", "amount", "period", "start_at", "created_at", "updated_at").
			OrderBy("created_at DESC")
	}

	return &stmt
}

func (b *balanceSubscriptionStore) Update(ctx context.Context, subscription *models.BalanceSubscription) error {
	_, err := b.db.DB.ExecContext(
		ctx,
		`
		UPDATE balance_subscriptions
		SET
			category_id = $1,
			name = $2,
			amount = $3,
			period = $4,
			start_at = $5,
			updated_at = NOW()
		WHERE
			id = $6;`,
		subscription.CategoryID, subscription.Name, subscription.Amount, subscription.Period, subscription.StartAt, subscription.ID,
	)
	return err
}

func (b *balanceSubscriptionStore) Delete(ctx context.Context, subscriptionID string) error {
	_, err := b.db.DB.ExecContext(ctx, "DELETE FROM balance_subscriptions WHERE id = $1;", subscriptionID)
	return err
}
