package models

// Operation represent a financial operation.
type Operation struct {
	ID         string        `bson:"_id,omitempty"`
	Type       OperationType `bson:"type,ommitempty"`
	CategoryID string        `bson:"categoryId,omitempty"`
	BalanceID  string        `bson:"balanceId,omitempty"`
	Amount     string        `bson:"amount,omitempty"`
}

// OperationType represents the type of an operation, which can be either incoming or spending.
type OperationType string

const (
	// OperationTypeIncoming represents a incoming operation.
	OperationTypeIncoming OperationType = "incoming"
	// OperationTypeSpending represents a spending operation.
	OperationTypeSpending OperationType = "spending"
)
