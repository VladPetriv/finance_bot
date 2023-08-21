package models

import "time"

// Operation represent a financial operation.
type Operation struct {
	ID         string        `bson:"_id,omitempty"`
	Type       OperationType `bson:"type,ommitempty"`
	CategoryID string        `bson:"categoryId,omitempty"`
	BalanceID  string        `bson:"balanceId,omitempty"`
	Amount     string        `bson:"amount,omitempty"`
	CreatedAt  time.Time     `bson:"createdAt,omitempty"`
}

// OperationType represents the type of an operation, which can be either incoming or spending.
type OperationType string

const (
	// OperationTypeIncoming represents a incoming operation.
	OperationTypeIncoming OperationType = "incoming"
	// OperationTypeSpending represents a spending operation.
	OperationTypeSpending OperationType = "spending"
)

// CreationPeriod defines constants representing different time periods for creation operations.
type CreationPeriod string

var (
	// CreationPeriodDay represents a daily time period.
	CreationPeriodDay CreationPeriod = "day"
	// CreationPeriodWeek represents a weekly time period.
	CreationPeriodWeek CreationPeriod = "week"
	// CreationPeriodMonth represents a monthly time period.
	CreationPeriodMonth CreationPeriod = "month"
	// CreationPeriodYear represents a yearly time period.
	CreationPeriodYear CreationPeriod = "year"
)
