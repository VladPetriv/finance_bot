package models

// Balance represents a balance model.
type Balance struct {
	ID     string `bson:"_id,omitempty"`
	UserID string `bson:"userId,omitempty"`

	Name     string `bson:"name,omitempty"`
	Amount   string `bson:"amount,omitempty"`
	Currency string `bson:"currency,omitempty"`
}

// GetName returns the balance name.
func (b Balance) GetName() string {
	return b.Name
}
