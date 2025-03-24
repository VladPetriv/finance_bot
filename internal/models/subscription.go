package models

import (
	"fmt"
	"time"
)

// BalanceSubscription represents a subscription for an internal user balance.
type BalanceSubscription struct {
	ID         string `db:"id"`
	BalanceID  string `db:"balance_id"`
	CategoryID string `db:"category_id"`

	Name   string             `db:"name"`
	Amount string             `db:"amount"`
	Period SubscriptionPeriod `db:"period"`

	StartAt   time.Time `db:"start_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// SubscriptionPeriod represents the period of a subscription.
type SubscriptionPeriod string

const (
	// SubscriptionPeriodWeekly represents a weekly subscription period.
	SubscriptionPeriodWeekly SubscriptionPeriod = "weekly"
	// SubscriptionPeriodMonthly represents a monthly subscription period.
	SubscriptionPeriodMonthly SubscriptionPeriod = "monthly"
	// SubscriptionPeriodYearly represents a yearly subscription period.
	SubscriptionPeriodYearly SubscriptionPeriod = "yearly"
)

// ParseSubscriptionPeriod parses a string into a SubscriptionPeriod.
func ParseSubscriptionPeriod(period string) (SubscriptionPeriod, error) {
	switch period {
	case "weekly":
		return SubscriptionPeriodWeekly, nil
	case "monthly":
		return SubscriptionPeriodMonthly, nil
	case "yearly":
		return SubscriptionPeriodYearly, nil
	default:
		return "", fmt.Errorf("invalid subscription period: %s", period)
	}
}
