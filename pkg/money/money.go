package money

import (
	"github.com/shopspring/decimal"
)

// Money represents custom typo for processing money.
type Money struct {
	decimal decimal.Decimal
}

// Zero represents zero (0) amount.
// Zero always equals to 0 and to 0.0...N.
var Zero = NewFromInt(0)

// NewFromString parses string and returns decimal amount.
// Returns truncated off amount always with precision=2. If s is zero,
// will be returned Zero decimal without throwing an error.
func NewFromString(s string) (Money, error) {
	if len(s) == 0 {
		return Zero, nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Zero, err
	}
	return Money{d}, nil
}

// NewFromInt returns decimal from integer number.
func NewFromInt(i int64) Money {
	d := decimal.NewFromInt(i)
	return Money{d}
}

// Sub returns left - right amounts.
func (m Money) Sub(right Money) Money {
	return Money{m.decimal.Sub(right.decimal)}
}

// Inc increments left amount by right.
// Same as left = left + right; left+=right
func (m *Money) Inc(right Money) {
	m.decimal = m.decimal.Add(right.decimal)
}

// String returns string representation of float with 2 places after digit.
// Resulting string will be rounded to nearest.
func (m Money) String() string {
	return m.decimal.StringFixed(2)
}
