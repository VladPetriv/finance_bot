package models

// Balance represents a Balance model.
type Balance struct {
	ID     string `bson:"_id,omitempty"`
	UserID string `bson:"userId,omitempty"`
	Amount string `bson:"amount,omitempty"`
}
