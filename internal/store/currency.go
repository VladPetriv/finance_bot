package store

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
)

type currencyStore struct {
	*database.PostgreSQL
}

// NewCurrency creates a new currency store.
func NewCurrency(db *database.PostgreSQL) *currencyStore {
	return &currencyStore{
		db,
	}
}

func (c *currencyStore) CreateIfNotExists(ctx context.Context, currency *models.Currency) error {
	_, err := c.DB.ExecContext(
		ctx,
		"INSERT INTO currencies (id, name, code, symbol) VALUES ($1, $2, $3, $4) ON CONFLICT (code) DO NOTHING;",
		currency.ID, currency.Name, currency.Code, currency.Symbol,
	)

	return err
}

func (c *currencyStore) Count(ctx context.Context) (int, error) {
	var count int
	err := c.DB.GetContext(ctx, &count, "SELECT COUNT(*) FROM currencies;")
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (c *currencyStore) List(ctx context.Context, filter service.ListCurrenciesFilter) ([]models.Currency, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "name", "code", "symbol").
		From("currencies")

	if filter.Pagination != nil {
		stmt = applyLimitAndOffsetForStatement(stmt, filter.Pagination)
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, err
	}

	var currencies []models.Currency
	err = c.DB.SelectContext(ctx, &currencies, query, args...)
	if err != nil {
		return nil, err
	}

	return currencies, nil
}

func (c *currencyStore) Exists(ctx context.Context, filter service.ExistsCurrencyFilter) (bool, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("1").
		From("currencies")

	if filter.ID != "" {
		stmt = stmt.Where(sq.Eq{"id": filter.ID})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return false, err
	}

	var exists bool
	err = c.DB.GetContext(ctx, &exists, fmt.Sprintf("SELECT EXISTS (%s);", query), args...)
	if err != nil {
		return false, err
	}

	return exists, nil
}
