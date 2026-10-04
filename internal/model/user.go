package model

import (
	"fmt"
	"time"
)

// User represents an user model.
type User struct {
	ID       string `db:"id"`
	ChatID   int    `db:"chat_id"`
	Username string `db:"username"`

	Balances []Balance
	Settings *UserSettings
}

// GetLocation returns the user's timezone location, falling back to UTC when settings are not loaded.
func (u *User) GetLocation() *time.Location {
	if u.Settings == nil {
		return time.UTC
	}

	return u.Settings.GetLocation()
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

	AIParserEnabled                 bool   `db:"ai_parser_enabled"`
	NotifyAboutSubscriptionPayments bool   `db:"notify_about_subscription_payments"`
	Timezone                        string `db:"timezone"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// GetLocation returns the user's timezone location, falling back to UTC when it is empty or unknown.
func (u *UserSettings) GetLocation() *time.Location {
	location, err := time.LoadLocation(u.Timezone)
	if err != nil {
		return time.UTC
	}

	return location
}

// GetDetails returns brief and formatted information about user settings
func (u *UserSettings) GetDetails() string {
	aiParserIcon := "❌"
	aiParserStatus := "Disabled"
	if u.AIParserEnabled {
		aiParserIcon = "✅"
		aiParserStatus = "Enabled"
	}

	notifyIcon := "❌"
	notifyStatus := "Disabled"
	if u.NotifyAboutSubscriptionPayments {
		notifyIcon = "✅"
		notifyStatus = "Enabled"
	}

	return fmt.Sprintf(`⚙️ *User Settings*

🤖 AI Parser: %s %s
🔔 Subscription Notifications: %s %s
🌍 Timezone: %s`,
		aiParserIcon, aiParserStatus,
		notifyIcon, notifyStatus,
		u.GetLocation().String(),
	)
}
