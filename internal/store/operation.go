package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/models"
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

func (o *operationStore) Create(ctx context.Context, operation *models.Operation) error {
	var createdAt time.Time
	switch operation.CreatedAt.IsZero() {
	case true:
		createdAt = time.Now()
	case false:
		createdAt = operation.CreatedAt
	}

	_, err := o.DB.ExecContext(
		ctx,
		`INSERT INTO
			operations (id, category_id, balance_id, type, amount, description, created_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7);
		`,

		operation.ID, operation.CategoryID, operation.BalanceID, operation.Type, operation.Amount, operation.Description, createdAt,
	)
	return err
}

func (o *operationStore) Get(ctx context.Context, filter service.GetOperationFilter) (*models.Operation, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "category_id", "balance_id", "type", "amount", "description", "created_at", "updated_at").
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

	var operation models.Operation
	err = o.DB.GetContext(ctx, &operation, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &operation, nil
}

func (o *operationStore) List(ctx context.Context, filter service.ListOperationsFilter) ([]models.Operation, error) {
	stmt := applyListOperationsFilter(applyListOperationsOptions{listQuery: true}, filter)

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list operation query: %w", err)
	}

	var operations []models.Operation
	err = o.DB.SelectContext(ctx, &operations, query, args...)
	if err != nil {
		return nil, err
	}

	return operations, nil
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
		expectedColumns = []string{"id", "category_id", "balance_id", "type", "amount", "description", "created_at", "updated_at"}
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
		startDate, endDate := filter.CreationPeriod.CalculateTimeRange()
		stmt = stmt.Where(sq.GtOrEq{"created_at": startDate}).Where(sq.LtOrEq{"created_at": endDate})
	}

	if filter.Month != "" {
		startDate, endDate := filter.Month.GetTimeRange(time.Now())
		stmt = stmt.Where(sq.GtOrEq{"created_at": startDate}).Where(sq.LtOrEq{"created_at": endDate})
	}

	if filter.Pagination != nil {
		var offset uint64
		if filter.Pagination.Page > 1 {
			offset = uint64(filter.Pagination.Page*filter.Pagination.Limit) - uint64(filter.Pagination.Limit)
		}

		stmt = stmt.
			Limit(uint64(filter.Pagination.Limit)).
			Offset(offset)
	}

	if filter.OrderByCreatedAtDesc {
		stmt = stmt.GroupBy("id", "category_id", "balance_id", "type", "amount", "description", "created_at", "updated_at").
			OrderBy("created_at DESC")
	}

	return &stmt
}

func (o *operationStore) Update(ctx context.Context, operationID string, operation *models.Operation) error {
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
