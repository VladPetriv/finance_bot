package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"

	sq "github.com/Masterminds/squirrel"
)

type balanceStore struct {
	*database.PostgreSQL
}

// NewBalance returns new instance of balance store.
func NewBalance(db *database.PostgreSQL) *balanceStore {
	return &balanceStore{
		db,
	}
}
func (b *balanceStore) Create(ctx context.Context, balance *models.Balance) error {
	_, err := b.DB.ExecContext(
		ctx,
		"INSERT INTO balances (id, user_id, currency_id, name, amount) VALUES ($1, $2, $3, $4, $5);",
		balance.ID, balance.UserID, balance.CurrencyID, balance.Name, balance.Amount,
	)

	return err
}

func (b *balanceStore) Get(ctx context.Context, filter service.GetBalanceFilter) (*models.Balance, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "user_id", "currency_id", "name", "amount").
		From("balances")

	if filter.BalanceID != "" {
		stmt = stmt.Where(sq.Eq{"id": filter.BalanceID})
	}
	if filter.Name != "" {
		stmt = stmt.Where(sq.Eq{"name": filter.Name})
	}
	if filter.UserID != "" {
		stmt = stmt.Where(sq.Eq{"user_id": filter.UserID})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get balance query: %w", err)
	}

	var balance models.Balance
	err = b.DB.GetContext(ctx, &balance, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if filter.PreloadCurrency {
		var currency models.Currency
		err = b.DB.GetContext(ctx, &currency, "SELECT id, name, code, symbol FROM currencies WHERE id = $1;", balance.CurrencyID)
		if err != nil {
			return nil, err
		}

		balance.Currency = currency
	}

	return &balance, nil
}

func (b *balanceStore) Update(ctx context.Context, balance *models.Balance) error {
	_, err := b.DB.ExecContext(
		ctx,
		"UPDATE balances SET user_id = $2, currency_id = $3, name = $4, amount = $5 WHERE id = $1;",
		balance.ID, balance.UserID, balance.CurrencyID, balance.Name, balance.Amount,
	)

	if err != nil {
		return err
	}

	return nil
}

func (b *balanceStore) Delete(ctx context.Context, balanceID string) error {
	_, err := b.DB.ExecContext(ctx, "DELETE FROM balances WHERE id = $1;", balanceID)
	return err
}
