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

// GetDeletionMessage returns the deletion message for the balance subscription.
func (b BalanceSubscription) GetDeletionMessage() string {
	return fmt.Sprintf(
		"Are you sure you want to delete the subscription '%s' (%s, %s)?\nAll previously created operations will remain untouched, but no new operations will be generated in the future.",
		b.Name,
		b.Amount,
		b.Period,
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

// CalculateScheduledOperationBillingDates generates future creation dates for an operation.
// It creates dates based on subscription period (weekly/monthly/yearly), starting
// from the start date and continuing until either the max occurrences or
// end window date is reached, whichever comes first.
// Returns: A sorted slice of dates when operations should be created.
func CalculateScheduledOperationBillingDates(period SubscriptionPeriod, startDate time.Time, maxDates int) []time.Time {
	billingDates := make([]time.Time, 0, maxDates)

	for value := range maxDates {
		switch period {
		case SubscriptionPeriodWeekly:
			billingDates = append(billingDates, startDate.AddDate(0, 0, value*7))
		case SubscriptionPeriodMonthly:
			billingDates = append(billingDates, startDate.AddDate(0, value, 0))
		case SubscriptionPeriodYearly:
			billingDates = append(billingDates, startDate.AddDate(value, 0, 0))
		}
	}

	return billingDates
}

// ScheduledOperationCreation represents a scheduled time for operation that will be created based on the subscription.
type ScheduledOperationCreation struct {
	ID             string    `db:"id"`
	SubscriptionID string    `db:"subscription_id"`
	CreationDate   time.Time `db:"creation_date"`
}
