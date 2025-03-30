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

// GetID returns the ID of the subscription.
func (b BalanceSubscription) GetID() string {
	return b.ID
}

// GetName returns the name of the subscription.
func (b BalanceSubscription) GetName() string {
	return b.Name
}

// GetDetails returns the balance subscription details in string format.
func (b BalanceSubscription) GetDetails() string {
	return fmt.Sprintf(
		"Subscription Details:\nName: %s\nAmount: %s\nPeriod: %s\nStart At: %s",
		b.Name, b.Amount, b.Period, b.StartAt.Format("2006-01-02 15:04"),
	)
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
