package model

import "time"

// User represents an user model.
type User struct {
	ID       string `db:"id"`
	ChatID   int    `db:"chat_id"`
	Username string `db:"username"`

	Balances []Balance
	Settings *UserSettings
}

// GetBalancesIDs returns the balances IDs.
func (u *User) GetBalancesIDs() []string {
	ids := make([]string, 0, len(u.Balances))
	for _, balance := range u.Balances {
		ids = append(ids, balance.ID)
	}

	return ids
}

// GetBalance returns the balance by matching the input with a name or ID.
func (u *User) GetBalance(value string) *Balance {
	for _, balance := range u.Balances {
		if balance.Name == value || balance.ID == value {
			return &balance
		}
	}

	return nil
}

// UserSettings represents a user settings model.
type UserSettings struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`

	AIParserEnabled                 bool `db:"ai_parser_enabled"`
	NotifyAboutSubscriptionPayments bool `db:"notify_about_subscription_payments"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
