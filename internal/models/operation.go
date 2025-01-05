package models

import (
	"time"
)

// Operation represent a financial operation.
type Operation struct {
	ID         string `bson:"_id,omitempty"`
	CategoryID string `bson:"categoryId,omitempty"`
	BalanceID  string `bson:"balanceId,omitempty"`

	Type        OperationType `bson:"type,ommitempty"`
	Amount      string        `bson:"amount,omitempty"`
	Description string        `bson:"description,omitempty"`

	CreatedAt time.Time `bson:"createdAt,omitempty"`
}

// OperationType represents the type of an operation, which can be either incoming, spending or transfer.
type OperationType string

const (
	// OperationTypeIncoming represents a incoming operation.
	OperationTypeIncoming OperationType = "incoming"
	// OperationTypeSpending represents a spending operation.
	OperationTypeSpending OperationType = "spending"
	// OperationTypeTransfer represents a transfer operation.
	OperationTypeTransfer OperationType = "transfer"
	// OperationTypeTransferIn represents a transfer_in operation.
	OperationTypeTransferIn OperationType = "transfer_in"
	// OperationTypeTransferOut represents a transfer_out operation.
	OperationTypeTransferOut OperationType = "transfer_out"
)

// CreationPeriod defines constants representing different time periods for creation operations.
type CreationPeriod string

// GetCreationPeriodFromText checks if the input text matches any of the CreationPeriod enums.
// If there is no match, it returns nil. Otherwise, it returns the corresponding CreationPeriod type.
func GetCreationPeriodFromText(value string) *CreationPeriod {
	switch value {
	case string(CreationPeriodDay):
		return &CreationPeriodDay
	case string(CreationPeriodWeek):
		return &CreationPeriodWeek
	case string(CreationPeriodMonth):
		return &CreationPeriodMonth
	case string(CreationPeriodYear):
		return &CreationPeriodYear
	default:
		return nil
	}
}

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
