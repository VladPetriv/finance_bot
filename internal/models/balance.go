package models

// Balance represents a Balance model.
type Balance struct {
	ID     string   `bson:"_id,omitempty"`
	Amount *float32 `bson:"amount,omitempty"`
}
