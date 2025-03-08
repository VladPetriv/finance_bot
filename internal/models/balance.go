package models

// Balance represents a balance model.
type Balance struct {
	ID         string `db:"id"`
	UserID     string `db:"user_id"`
	CurrencyID string `db:"currency_id"`

	Name   string `db:"name"`
	Amount string `db:"amount"`

	Currency Currency
}

// GetName returns the balance name.
func (b Balance) GetName() string {
	return b.Name
}

// GetCurrency returns information about the currency of the balance.
func (b Balance) GetCurrency() Currency {
	return b.Currency
}
