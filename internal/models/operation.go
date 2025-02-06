package models

import (
	"fmt"
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

var typeShort = map[OperationType]string{
	OperationTypeIncoming:    "IN",
	OperationTypeSpending:    "OUT",
	OperationTypeTransfer:    "TRF",
	OperationTypeTransferIn:  "TRF-IN",
	OperationTypeTransferOut: "TRF-OUT",
}

// GetName returns a string representation of the operation.
func (o *Operation) GetName() string {
	truncatedDescription := o.Description
	if len(o.Description) > 10 {
		truncatedDescription = o.Description[:10] + "..."
	}

	return fmt.Sprintf("%s, %s, %s, %s",
		o.CreatedAt.Format("02.01 15:04:05"),
		o.Amount,
		typeShort[o.Type],
		truncatedDescription,
	)
}

// GetDeletionMessage generates a confirmation message for operation deletion,
// including operation details and warnings about balance impacts based on the
// operation type (transfer, spending, or incoming).
func (o *Operation) GetDeletionMessage() string {
	var balanceImpactMsg string
	switch o.Type {
	case OperationTypeTransferIn, OperationTypeTransferOut:
		balanceImpactMsg = fmt.Sprintf(
			"⚠️ Warning: Deleting this transfer operation will:\n"+
				"• Increase balance of source account by %s\n"+
				"• Decrease balance of destination account by %s",
			o.Amount,
			o.Amount)
	case OperationTypeSpending, OperationTypeIncoming:
		var effect string
		if o.Type == "income" {
			effect = "decrease"
		} else {
			effect = "increase"
		}
		balanceImpactMsg = fmt.Sprintf(
			"⚠️ Warning: Deleting this %s operation will %s your balance by %s",
			o.Type,
			effect,
			o.Amount)
	}

	return fmt.Sprintf(
		"Are you sure you want to proceed with the selected operation?\n\n"+
			"Operation Details:\n"+
			"Type: %s\n"+
			"Amount: %s\n"+
			"Description: %s\n\n"+
			"%s\n\n"+
			"This action cannot be undone.",
		o.Type, o.Amount, o.Description,
		balanceImpactMsg,
	)
}

// GetDetails returns the operation details in string format.
func (o *Operation) GetDetails() string {
	return fmt.Sprintf(
		"Operation Details:\nType: %s\nAmount: %s\nDescription: %s",
		o.Type, o.Amount, o.Description,
	)
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
