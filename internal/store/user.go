package store

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
)

type userStore struct {
	*database.PostgreSQL
}

// NewUser returns new instance of user store.
func NewUser(db *database.PostgreSQL) *userStore {
	return &userStore{
		db,
	}
}

func (u *userStore) Create(ctx context.Context, user *models.User) error {
	_, err := u.DB.ExecContext(
		ctx,
		"INSERT INTO users (id, username) VALUES ($1, $2);",
		user.ID, user.Username,
	)

	return err
}

func (u *userStore) CreateSettings(ctx context.Context, settings *models.UserSettings) error {
	_, err := u.DB.ExecContext(
		ctx,
		"INSERT INTO user_settings (id, user_id, ai_parser_enabled) VALUES ($1, $2, $3);",
		settings.ID, settings.UserID, settings.AIParserEnabled,
	)

	return err
}

func (u userStore) Get(ctx context.Context, filter service.GetUserFilter) (*models.User, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "username").
		From("users")

	if filter.Username != "" {
		stmt = stmt.Where(sq.Eq{"username": filter.Username})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, err
	}

	var user models.User
	err = u.DB.GetContext(ctx, &user, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	if filter.PreloadBalances {
		stmt := sq.
			StatementBuilder.
			PlaceholderFormat(sq.Dollar).
			Select("id", "user_id", "currency_id", "name", "amount").
			From("balances").
			Where(sq.Eq{"user_id": user.ID})

		query, args, err := stmt.ToSql()
		if err != nil {
			return nil, err
		}

		var balances []models.Balance
		err = u.DB.SelectContext(ctx, &balances, query, args...)
		if err != nil {
			return nil, err
		}

		user.Balances = balances
	}
	if filter.PreloadSettings {
		stmt := sq.
			StatementBuilder.
			PlaceholderFormat(sq.Dollar).
			Select("id", "user_id", "ai_parser_enabled", "created_at", "updated_at").
			From("user_settings").
			Where(sq.Eq{"user_id": user.ID})

		query, args, err := stmt.ToSql()
		if err != nil {
			return nil, err
		}

		var userSettings models.UserSettings
		err = u.DB.GetContext(ctx, &userSettings, query, args...)
		if err != nil {
			return nil, err
		}

		user.Settings = &userSettings
	}

	return &user, nil
}
