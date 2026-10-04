package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/lib/pq"
)

type automaticReportStore struct {
	*database.PostgreSQL
}

// NewAutomaticReport creates a new instance of automatic report store.
func NewAutomaticReport(db *database.PostgreSQL) *automaticReportStore {
	return &automaticReportStore{
		db,
	}
}

func (a *automaticReportStore) Create(ctx context.Context, report *model.AutomaticReport) error {
	tx, err := a.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO automatic_reports (id, name, period) VALUES ($1, $2, $3)",
		report.ID, report.Name, report.Period,
	)
	if err != nil {
		return err
	}

	for _, balanceID := range report.BalanceIDs {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO automatic_report_balances (automatic_report_id, balance_id) VALUES ($1, $2)",
			report.ID, balanceID,
		)
		if err != nil {
			return err
		}
	}

	for _, categoryID := range report.CategoryIDs {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO automatic_report_categories (automatic_report_id, category_id) VALUES ($1, $2)",
			report.ID, categoryID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (a *automaticReportStore) CreateScheduledReportExecution(ctx context.Context, execution *model.ScheduledReportExecution) error {
	_, err := a.DB.ExecContext(
		ctx,
		"INSERT INTO scheduled_report_executions (id, automatic_report_id, execution_date) VALUES ($1, $2, $3)",
		execution.ID, execution.AutomaticReportID, execution.ExecutionDate,
	)
	return err
}

func (a *automaticReportStore) Get(ctx context.Context, filter service.GetAutomaticReportFilter) (*model.AutomaticReport, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select(
			"automatic_reports.id",
			"automatic_reports.name",
			"automatic_reports.period",
			"COALESCE(array_agg(DISTINCT automatic_report_balances.balance_id), '{}') AS balance_ids",
			"COALESCE(array_agg(DISTINCT automatic_report_categories.category_id), '{}') AS category_ids",
		).
		From("automatic_reports").
		LeftJoin("automatic_report_balances ON automatic_reports.id = automatic_report_balances.automatic_report_id").
		LeftJoin("automatic_report_categories ON automatic_reports.id = automatic_report_categories.automatic_report_id").
		GroupBy("automatic_reports.id")

	if filter.ID != "" {
		stmt = stmt.Where(sq.Eq{"automatic_reports.id": filter.ID})
	}
	if filter.Name != "" {
		stmt = stmt.Where(sq.Eq{"automatic_reports.name": filter.Name})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get automatic report query: %w", err)
	}

	var automaticReport model.AutomaticReport
	err = a.DB.GetContext(ctx, &automaticReport, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &automaticReport, nil
}

func (a *automaticReportStore) Count(ctx context.Context, filter service.ListAutomaticReportFilter) (int, error) {
	stmt := applyAutomaticReportFilterForCountQuery(filter)

	query, args, err := stmt.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count automatic reports query: %w", err)
	}

	var count int64
	err = a.DB.GetContext(ctx, &count, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, err
	}

	return int(count), nil
}

func applyAutomaticReportFilterForCountQuery(filter service.ListAutomaticReportFilter) *sq.SelectBuilder {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("COUNT(DISTINCT automatic_reports.id)").
		From("automatic_reports")

	if len(filter.BalanceIDs) > 0 {
		stmt = stmt.Where(
			`EXISTS (
				SELECT
					1
				FROM
					automatic_report_balances
				WHERE
					automatic_report_balances.automatic_report_id = automatic_reports.id
					AND automatic_report_balances.balance_id = ANY(?)
			)`,
			pq.Array(filter.BalanceIDs),
		)
	}

	if len(filter.CategoryIDs) > 0 {
		stmt = stmt.Where(
			`EXISTS (
				SELECT
					1
				FROM
					automatic_report_categories
				WHERE
					automatic_report_categories.automatic_report_id = automatic_reports.id
					AND automatic_report_categories.category_id = ANY(?)
			)`,
			pq.Array(filter.CategoryIDs),
		)
	}

	return &stmt
}

func (a *automaticReportStore) List(ctx context.Context, filter service.ListAutomaticReportFilter) ([]model.AutomaticReport, error) {
	stmt := applyAutomaticReportFilterForListQuery(filter)

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list automatic reports query: %w", err)
	}

	var automaticReports []model.AutomaticReport
	err = a.DB.SelectContext(ctx, &automaticReports, query, args...)
	if err != nil {
		return nil, err
	}

	return automaticReports, nil
}

func applyAutomaticReportFilterForListQuery(filter service.ListAutomaticReportFilter) *sq.SelectBuilder {
	expectedColumns := []string{
		"automatic_reports.id",
		"automatic_reports.name",
		"automatic_reports.period",
		"COALESCE(array_agg(DISTINCT automatic_report_balances.balance_id), '{}') AS balance_ids",
		"COALESCE(array_agg(DISTINCT automatic_report_categories.category_id), '{}') AS category_ids",
	}

	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select(expectedColumns...).
		From("automatic_reports").
		Join("automatic_report_balances ON automatic_reports.id = automatic_report_balances.automatic_report_id").
		Join("automatic_report_categories ON automatic_reports.id = automatic_report_categories.automatic_report_id").
		GroupBy("automatic_reports.id")

	if len(filter.BalanceIDs) > 0 {
		stmt = stmt.Having("array_agg(DISTINCT automatic_report_balances.balance_id) && ?", pq.Array(filter.BalanceIDs))
	}

	if len(filter.CategoryIDs) > 0 {
		stmt = stmt.Having("array_agg(DISTINCT automatic_report_categories.category_id) && ?", pq.Array(filter.CategoryIDs))
	}

	if filter.Pagination != nil {
		stmt = applyLimitAndOffsetForStatement(stmt, filter.Pagination)
	}

	if filter.OrderByCreatedAtDesc {
		stmt = stmt.GroupBy(expectedColumns[:len(expectedColumns)-2]...).
			OrderBy("automatic_reports.created_at DESC")
	}

	return &stmt
}

func (a *automaticReportStore) ListScheduledReportExecutions(ctx context.Context, filter service.ListScheduledReportExecutionFilter) ([]model.ScheduledReportExecution, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select(
			"scheduled_report_executions.id",
			"scheduled_report_executions.automatic_report_id",
			"scheduled_report_executions.execution_date",
		).
		From("scheduled_report_executions")

	if filter.AutomaticReportID != "" {
		stmt = stmt.Where(sq.Eq{"scheduled_report_executions.automatic_report_id": filter.AutomaticReportID})
	}

	if filter.BetweenFilter != nil {
		stmt = stmt.Where(sq.And{
			sq.GtOrEq{"scheduled_report_executions.execution_date": filter.BetweenFilter.From},
			sq.Lt{"scheduled_report_executions.execution_date": filter.BetweenFilter.To},
		})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list scheduled report executions query: %w", err)
	}

	var scheduledReportExecutions []model.ScheduledReportExecution
	err = a.DB.SelectContext(ctx, &scheduledReportExecutions, query, args...)
	if err != nil {
		return nil, err
	}

	return scheduledReportExecutions, nil
}

func (a *automaticReportStore) Update(ctx context.Context, report *model.AutomaticReport) error {
	tx, err := a.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(
		ctx,
		"UPDATE automatic_reports SET name = $1, period = $2, updated_at = NOW() WHERE id = $3",
		report.Name, report.Period, report.ID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM automatic_report_balances WHERE automatic_report_id = $1",
		report.ID,
	)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM automatic_report_categories WHERE automatic_report_id = $1",
		report.ID,
	)
	if err != nil {
		return err
	}

	for _, balanceID := range report.BalanceIDs {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO automatic_report_balances (automatic_report_id, balance_id) VALUES ($1, $2)",
			report.ID, balanceID,
		)
		if err != nil {
			return err
		}
	}
	for _, categoryID := range report.CategoryIDs {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO automatic_report_categories (automatic_report_id, category_id) VALUES ($1, $2)",
			report.ID, categoryID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (a *automaticReportStore) Delete(ctx context.Context, id string) error {
	_, err := a.DB.ExecContext(
		ctx,
		"DELETE FROM automatic_reports WHERE id = $1",
		id,
	)

	return err
}

func (a *automaticReportStore) DeleteScheduledReportExecution(ctx context.Context, id string) error {
	_, err := a.DB.ExecContext(
		ctx,
		"DELETE FROM scheduled_report_executions WHERE id = $1",
		id,
	)

	return err
}
