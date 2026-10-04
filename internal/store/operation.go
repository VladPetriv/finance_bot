package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
)

type operationStore struct {
	*database.PostgreSQL
}

// NewOperation returns new instance of operation store.
func NewOperation(db *database.PostgreSQL) *operationStore {
	return &operationStore{
		db,
	}
}

func (o *operationStore) Create(ctx context.Context, operation *model.Operation) error {
	var createdAt time.Time
	switch operation.CreatedAt.IsZero() {
	case true:
		createdAt = time.Now().UTC()
	case false:
		createdAt = operation.CreatedAt
	}

	_, err := o.DB.ExecContext(
		ctx,
		`INSERT INTO
			operations (id, category_id, balance_id, parent_operation_id, type, amount, exchange_rate, description, created_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9);
		`,

		operation.ID, operation.CategoryID, operation.BalanceID, operation.ParentOperationID, operation.Type, operation.Amount, operation.ExchangeRate, operation.Description, createdAt,
	)
	return err
}

func (o *operationStore) Get(ctx context.Context, filter service.GetOperationFilter) (*model.Operation, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "category_id", "balance_id", "parent_operation_id", "type", "amount", "exchange_rate", "description", "created_at", "updated_at").
		From("operations")

	if filter.ID != "" {
		stmt = stmt.Where(sq.Eq{"id": filter.ID})
	}
	if filter.Type != "" {
		stmt = stmt.Where(sq.Eq{"type": filter.Type})
	}
	if filter.Amount != "" {
		stmt = stmt.Where(sq.Eq{"amount": filter.Amount})
	}
	if len(filter.BalanceIDs) != 0 {
		stmt = stmt.Where(sq.Eq{"balance_id": filter.BalanceIDs})
	}
	if !filter.CreateAtFrom.IsZero() {
		stmt = stmt.Where(sq.Gt{"created_at": filter.CreateAtFrom})
	}
	if !filter.CreateAtTo.IsZero() {
		stmt = stmt.Where(sq.Lt{"created_at": filter.CreateAtTo})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get operation query: %w", err)
	}

	var operation model.Operation
	err = o.DB.GetContext(ctx, &operation, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &operation, nil
}

func (o *operationStore) List(ctx context.Context, filter service.ListOperationsFilter) ([]model.Operation, error) {
	stmt := applyListOperationsFilter(applyListOperationsOptions{listQuery: true}, filter)

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list operation query: %w", err)
	}

	var operations []model.Operation
	err = o.DB.SelectContext(ctx, &operations, query, args...)
	if err != nil {
		return nil, err
	}

	return operations, nil
}

func (o *operationStore) ListOperationYears(ctx context.Context, filter service.ListOperationYearsFilter) ([]model.Year, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select().
		Column(sq.Expr("EXTRACT(YEAR FROM created_at AT TIME ZONE ?)::int AS year", locationOrUTC(filter.Location).String())).
		Distinct().
		From("operations").
		GroupBy("year").
		OrderBy("year")

	if filter.BalanceID != "" {
		stmt = stmt.Where(sq.Eq{"operations.balance_id": filter.BalanceID})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list operation year query: %w", err)
	}

	var years []model.Year
	err = o.DB.SelectContext(ctx, &years, query, args...)
	if err != nil {
		return nil, err
	}

	return years, nil
}

func (o *operationStore) ListOperationMonths(ctx context.Context, filter service.ListOperationMonthsFilter) ([]model.Month, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select().
		Column(sq.Expr("EXTRACT(MONTH FROM created_at AT TIME ZONE ?)::int AS month", locationOrUTC(filter.Location).String())).
		Distinct().
		From("operations").
		GroupBy("month").
		OrderBy("month")

	if filter.BalanceID != "" {
		stmt = stmt.Where(sq.Eq{"operations.balance_id": filter.BalanceID})
	}
	if filter.Year != 0 {
		location := locationOrUTC(filter.Location)
		start := time.Date(filter.Year, 1, 1, 0, 0, 0, 0, location)
		end := time.Date(filter.Year+1, 1, 1, 0, 0, 0, 0, location)

		stmt = stmt.Where(
			sq.And{
				sq.GtOrEq{"operations.created_at": start},
				sq.Lt{"operations.created_at": end},
			},
		)
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list operation months query: %w", err)
	}

	var months []int
	err = o.DB.SelectContext(ctx, &months, query, args...)
	if err != nil {
		return nil, err
	}
	if len(months) == 0 {
		return nil, nil
	}

	result := make([]model.Month, len(months))
	for i, month := range months {
		result[i] = model.Months[month-1]
	}

	return result, nil
}

func (o *operationStore) Count(ctx context.Context, filter service.ListOperationsFilter) (int, error) {
	stmt := applyListOperationsFilter(applyListOperationsOptions{countQuery: true}, filter)

	query, args, err := stmt.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count operation query: %w", err)
	}

	var count int64
	err = o.DB.GetContext(ctx, &count, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return int(count), nil
}

type applyListOperationsOptions struct {
	listQuery  bool
	countQuery bool
}

func applyListOperationsFilter(options applyListOperationsOptions, filter service.ListOperationsFilter) *sq.SelectBuilder {
	var expectedColumns []string
	if options.countQuery {
		expectedColumns = []string{"COUNT(id)"}
	}

	if options.listQuery {
		expectedColumns = []string{"id", "category_id", "balance_id", "parent_operation_id", "type", "amount", "exchange_rate", "description", "created_at", "updated_at"}
	}

	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select(expectedColumns...).
		From("operations")

	if filter.BalanceID != "" {
		stmt = stmt.Where(sq.Eq{"balance_id": filter.BalanceID})
	}

	if filter.CreationPeriod != "" {
		startDate, endDate := filter.CreationPeriod.CalculateTimeRange(locationOrUTC(filter.Location))
		stmt = stmt.Where(sq.GtOrEq{"created_at": startDate}).Where(sq.LtOrEq{"created_at": endDate})
	}

	if filter.Month != "" {
		year, _ := strconv.Atoi(filter.Year.GetName())
		startDate, endDate := filter.Month.GetTimeRange(time.Now(), year, locationOrUTC(filter.Location))
		stmt = stmt.Where(sq.GtOrEq{"created_at": startDate}).Where(sq.LtOrEq{"created_at": endDate})
	}

	if filter.BetweenFilter != nil {
		stmt = stmt.Where(sq.GtOrEq{"created_at": filter.BetweenFilter.From}).Where(sq.Lt{"created_at": filter.BetweenFilter.To})
	}

	if filter.Pagination != nil {
		stmt = applyLimitAndOffsetForStatement(stmt, filter.Pagination)
	}

	if filter.OrderByCreatedAtDesc {
		stmt = stmt.GroupBy("id", "category_id", "balance_id", "parent_operation_id", "type", "amount", "exchange_rate", "description", "created_at", "updated_at").
			OrderBy("created_at DESC")
	}

	return &stmt
}

func (o *operationStore) Update(ctx context.Context, operationID string, operation *model.Operation) error {
	_, err := o.DB.ExecContext(
		ctx,
		`UPDATE operations
		SET
			category_id = $1,
			balance_id = $2,
			type = $3,
			amount = $4,
			description = $5,
			created_at = $6,
			updated_at = NOW()
		WHERE
			id = $7;`,
		operation.CategoryID, operation.BalanceID, operation.Type, operation.Amount, operation.Description, operation.CreatedAt, operationID,
	)

	return err
}

func (o *operationStore) Delete(ctx context.Context, operationID string) error {
	_, err := o.DB.ExecContext(ctx, "DELETE FROM operations WHERE id = $1;", operationID)
	return err
}

func locationOrUTC(location *time.Location) *time.Location {
	if location == nil {
		return time.UTC
	}

	return location
}
