package model

import (
	"fmt"
	"time"

	"github.com/VladPetriv/finance_bot/pkg/money"
)

// Balance represents a balance model.
type Balance struct {
	ID         string `db:"id"`
	UserID     string `db:"user_id"`
	CurrencyID string `db:"currency_id"`

	Name   string `db:"name"`
	Amount string `db:"amount"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Currency Currency
}

// GetID returns the balance ID.
func (b Balance) GetID() string {
	return b.ID
}

// GetName returns the balance name.
func (b Balance) GetName() string {
	return b.Name
}

// GetCurrency returns information about the currency of the balance.
func (b Balance) GetCurrency() Currency {
	return b.Currency
}

// BuildCurrencyConversionMessage creates a formatted message prompting the user
// for an exchange rate when transferring between different currencies.
// It includes source/destination balance info and an example conversion using a 4x rate.
func BuildCurrencyConversionMessage(balanceFrom, balanceTo *Balance) string {
	parsedAmount, _ := money.NewFromString(balanceFrom.Amount)
	parsedAmount.Mul(money.NewFromInt(4))

	return fmt.Sprintf(`⚠️ Different Currency Transfer ⚠️
Source Balance: %s
Currency: %s
Amount: %v %s

Destination Balance: %s
Currency: %s

To accurately convert your money, please provide the current exchange rate:

Formula: 1 %s = X %s
(How many %s you get for 1 %s)

Example:
- If 1 %s = 4 %s, enter: 4
- This means %v %s will be converted to %v %s

Please enter the current exchange rate:`,
		balanceFrom.Name,
		balanceFrom.GetCurrency().Symbol,
		balanceFrom.Amount,
		balanceFrom.GetCurrency().Symbol,
		balanceTo.Name,
		balanceTo.GetCurrency().Symbol,
		balanceFrom.GetCurrency().Symbol,
		balanceTo.GetCurrency().Symbol,
		balanceTo.GetCurrency().Symbol,
		balanceFrom.GetCurrency().Symbol,
		balanceFrom.GetCurrency().Symbol,
		balanceTo.GetCurrency().Symbol,
		balanceFrom.Amount,
		balanceFrom.GetCurrency().Symbol,
		parsedAmount.StringFixed(),
		balanceTo.GetCurrency().Symbol,
	)
}
