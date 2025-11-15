package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Metadata represents metadata associated with a flow
type Metadata map[MetadataKey]any

// Add adds a key-value pair to the metadata.
func (m Metadata) Add(key MetadataKey, value any) {
	if m == nil {
		m = Metadata{}
	}
	(m)[key] = value
}

// Value implements the driver.Valuer interface
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface
func (m *Metadata) Scan(value any) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, m)
}

// GetTypedFromMetadata retrieves the value associated with the given key and expected value type.
func GetTypedFromMetadata[T any](m Metadata, key MetadataKey) (T, bool) {
	var zero T
	if m == nil {
		return zero, false
	}

	val, exists := m[key]
	if !exists {
		return zero, false
	}

	typed, ok := val.(T)
	return typed, ok
}

// MetadataKey represents a key in the metadata map.
type MetadataKey string

const (
	// General keys

	// BaseFlowMetadataKey represents the base flow of current state.
	BaseFlowMetadataKey MetadataKey = "base_flow"
	// PageMetadataKey represents the current page number.
	PageMetadataKey MetadataKey = "page"

	// Balance related keys

	// BalanceIDMetadataKey represents the ID of the balance.
	BalanceIDMetadataKey MetadataKey = "balance_id"
	// BalanceNameMetadataKey represents the name of the balance.
	BalanceNameMetadataKey MetadataKey = "balance_name"
	// BalanceAmountMetadataKey represents the amount of the balance.
	BalanceAmountMetadataKey MetadataKey = "balance_amount"
	// BalanceFromMetadataKey represents the balance from which transfer was made.
	BalanceFromMetadataKey MetadataKey = "balance_from"
	// BalanceToMetadataKey represents the BalanceAmountMetadataKey.
	BalanceToMetadataKey MetadataKey = "balance_to"
	// CurrentBalanceNameMetadataKey represents the current balance name.
	CurrentBalanceNameMetadataKey MetadataKey = "current_balance_name"
	// MonthForBalanceStatisticsKey represents the month that was used for balance statistics.
	MonthForBalanceStatisticsKey MetadataKey = "month_for_balance_statistics"

	// Category related keys

	// PreviousCategoryIDMetadataKey represents the previous category ID.
	PreviousCategoryIDMetadataKey MetadataKey = "previous_category_id"
	// CategoryTitleMetadataKey represents the title of the category.
	CategoryTitleMetadataKey MetadataKey = "category_title"
	// CategoryIDMetadataKey represents the ID of the category.
	CategoryIDMetadataKey MetadataKey = "category_id"

	// Operation related keys

	// ExchangeRateMetadataKey represents the exchange rate.
	ExchangeRateMetadataKey MetadataKey = "exchange_rate"
	// OperationDescriptionMetadataKey represents the description of the operation.
	OperationDescriptionMetadataKey MetadataKey = "operation_description"
	// OperationAmountMetadataKey represents the amount of the operation.
	OperationAmountMetadataKey MetadataKey = "operation_amount"
	// OperationTypeMetadataKey represents the type of the operation.
	OperationTypeMetadataKey MetadataKey = "operation_type"
	// OperationIDMetadataKey represents the ID of the operation.
	OperationIDMetadataKey MetadataKey = "operation_id"
	// OperationCreationPeriodMetadataKey represents the creation period of the operation.
	OperationCreationPeriodMetadataKey MetadataKey = "operation_creation_period"

	// Balance subscription related keys

	// BalanceSubscriptionIDMetadataKey represents the ID of the balance subscription.
	BalanceSubscriptionIDMetadataKey MetadataKey = "balance_subscription_id"
	// BalanceSubscriptionNameMetadataKey represents the name of the balance subscription.
	BalanceSubscriptionNameMetadataKey MetadataKey = "balance_subscription_name"
	// BalanceSubscriptionPeriodMetadataKey represents the period of the balance subscription.
	BalanceSubscriptionPeriodMetadataKey MetadataKey = "balance_subscription_period"
	// BalanceSubscriptionAmountMetadataKey represents the amount of the balance subscription.
	BalanceSubscriptionAmountMetadataKey MetadataKey = "balance_subscription_amount"
)
