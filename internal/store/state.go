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

type stateStore struct {
	*database.PostgreSQL
}

// NewState returns new instance of state store.
func NewState(db *database.PostgreSQL) *stateStore {
	return &stateStore{
		db,
	}
}

func (s *stateStore) Create(ctx context.Context, state *models.State) error {
	_, err := s.DB.ExecContext(
		ctx,
		"INSERT INTO states (id, user_username, flow, steps, metadata) VALUES ($1, $2, $3, $4, $5);",
		state.ID, state.UserID, state.Flow, state.Steps, state.Metedata,
	)

	return err
}

func (s *stateStore) Get(ctx context.Context, filter service.GetStateFilter) (*models.State, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "user_username", "flow", "steps", "metadata", "created_at", "updated_at").
		From("states")

	if filter.UserID != "" {
		stmt = stmt.Where(sq.Eq{"user_username": filter.UserID})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get state query: %w", err)
	}

	var state models.State
	err = s.DB.GetContext(ctx, &state, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &state, nil
}

func (s *stateStore) Update(ctx context.Context, state *models.State) (*models.State, error) {
	var updatedState models.State
	err := s.DB.QueryRowContext(
		ctx,
		`UPDATE states
		SET
			user_username = $1,
		 	flow = $2,
			steps = $3,
			metadata = $4,
		 	updated_at = NOW()
		WHERE id = $5
		RETURNING *;`,
		state.UserID, state.Flow, state.Steps, state.Metedata, state.ID,
	).Scan(
		&updatedState.ID,
		&updatedState.UserID,
		&updatedState.Flow,
		&updatedState.Steps,
		&updatedState.Metedata,
		&updatedState.CreatedAt,
		&updatedState.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updatedState, nil
}

func (s *stateStore) Delete(ctx context.Context, stateID string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM states WHERE id = $1;", stateID)
	return err
}
