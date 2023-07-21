package models

// Balance represents a balance model.
type Balance struct {
	ID       string `bson:"_id,omitempty"`
	UserID   string `bson:"userId,omitempty"`
	Amount   string `bson:"amount,omitempty"`
	Currency string `bson:"currency,omitempty"`
}
