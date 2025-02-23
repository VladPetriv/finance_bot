package models

// Balance represents a balance model.
type Balance struct {
	ID         string `bson:"_id,omitempty"`
	UserID     string `bson:"userId,omitempty"`
	CurrencyID string `bson:"currencyId,omitempty"`

	Name   string `bson:"name,omitempty"`
	Amount string `bson:"amount,omitempty"`

	Currency *Currency `bson:"currency"`
}

// GetName returns the balance name.
func (b Balance) GetName() string {
	return b.Name
}

// GetCurrency returns information about the currency of the balance.
func (b Balance) GetCurrency() Currency {
	if b.Currency == nil {
		return Currency{}
	}

	return *b.Currency
}
