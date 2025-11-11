package store

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/model"
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

func (u *userStore) Create(ctx context.Context, user *model.User) error {
	_, err := u.DB.ExecContext(
		ctx,
		"INSERT INTO users (id, chat_id, username) VALUES ($1, $2, $3);",
		user.ID, user.ChatID, user.Username,
	)

	return err
}

func (u *userStore) CreateSettings(ctx context.Context, settings *model.UserSettings) error {
	_, err := u.DB.ExecContext(
		ctx,
		"INSERT INTO user_settings (id, user_id, ai_parser_enabled, notify_about_subscription_payments) VALUES ($1, $2, $3, $4);",
		settings.ID, settings.UserID, settings.AIParserEnabled, settings.NotifyAboutSubscriptionPayments,
	)

	return err
}

func (u *userStore) UpdateSettings(ctx context.Context, settings *model.UserSettings) error {
	_, err := u.DB.ExecContext(
		ctx,
		"UPDATE user_settings SET ai_parser_enabled = $1, notify_about_subscription_payments = $2 WHERE id = $3;",
		settings.AIParserEnabled, settings.NotifyAboutSubscriptionPayments, settings.ID,
	)

	return err
}

func (u *userStore) Get(ctx context.Context, filter service.GetUserFilter) (*model.User, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("users.id", "users.chat_id", "users.username").
		From("users")

	if filter.Username != "" {
		stmt = stmt.Where(sq.Eq{"users.username": filter.Username})
	}
	if filter.BalanceID != "" {
		stmt = stmt.
			InnerJoin("balances ON balances.user_id = users.id").
			Where(sq.Eq{"balances.id": filter.BalanceID})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, err
	}

	var user model.User
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
			Select("id", "user_id", "currency_id", "name", "amount", "created_at", "updated_at").
			From("balances").
			OrderBy("created_at").
			Where(sq.Eq{"user_id": user.ID})

		query, args, err := stmt.ToSql()
		if err != nil {
			return nil, err
		}

		var balances []model.Balance
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
			Select("id", "user_id", "ai_parser_enabled", "notify_about_subscription_payments", "created_at", "updated_at").
			From("user_settings").
			Where(sq.Eq{"user_id": user.ID})

		query, args, err := stmt.ToSql()
		if err != nil {
			return nil, err
		}

		var userSettings model.UserSettings
		err = u.DB.GetContext(ctx, &userSettings, query, args...)
		if err != nil {
			return nil, err
		}

		user.Settings = &userSettings
	}

	return &user, nil
}
