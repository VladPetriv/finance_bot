package money

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

var errUnsupportedScanType = errors.New("unsupported scan type for money")

var errMinorUnitsOverflow = errors.New("amount in minor units does not fit into int64")

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

// NewFromFloat returns decimal from float number.
func NewFromFloat(i float64) Money {
	d := decimal.NewFromFloat(i)
	return Money{d}
}

// Inc increments left amount by right.
// Same as left = left + right; left+=right
func (m *Money) Inc(right Money) {
	m.decimal = m.decimal.Add(right.decimal)
}

// Sub decrements left - right amounts.
// Same as left = left - right; left-=right
func (m *Money) Sub(right Money) {
	m.decimal = m.decimal.Sub(right.decimal)
}

// Mul multiplies this Money value with another and updates the result in place
func (m *Money) Mul(right Money) {
	m.decimal = m.decimal.Mul(right.decimal)
}

// Div divides this Money value by another and updates the result in place
func (m *Money) Div(right Money) {
	m.decimal = m.decimal.Div(right.decimal)
}

// Equal returns true if current value is equal to input one.
func (m *Money) Equal(right Money) bool {
	return m.decimal.Equal(right.decimal)
}

// GreaterThan returns true if current value is greater than input one.
func (m *Money) GreaterThan(right Money) bool {
	return m.decimal.GreaterThan(right.decimal)
}

// LessThan returns true if current value is less than input one.
func (m *Money) LessThan(right Money) bool {
	return m.decimal.LessThan(right.decimal)
}

// StringFixed returns string representation of float with 2 places after digit.
// Resulting string will be rounded to nearest.
func (m Money) StringFixed() string {
	return m.decimal.StringFixed(2)
}

// String returns string representation of float with unlimited places after digit.
// Resulting string will be rounded to nearest.
func (m Money) String() string {
	return m.decimal.String()
}

// Round rounds the value in place to 2 places, half away from zero.
func (m *Money) Round() {
	m.decimal = m.decimal.Round(minorUnitsExponent)
}

// Set sets the value of the Money instance to the provided value.
func (m *Money) Set(value Money) {
	m.decimal = value.decimal
}

const minorUnitsExponent = 2

// NewFromMinorUnits returns money from an integer number of minor units (1/100).
func NewFromMinorUnits(minorUnits int64) Money {
	return Money{
		decimal: decimal.New(minorUnits, -minorUnitsExponent),
	}
}

// MinorUnits returns the amount in minor units, rounded half away from zero to 2 places.
func (m Money) MinorUnits() (int64, error) {
	minorUnits := m.decimal.Round(minorUnitsExponent).Shift(minorUnitsExponent).BigInt()
	if !minorUnits.IsInt64() {
		return 0, errMinorUnitsOverflow
	}

	return minorUnits.Int64(), nil
}

// Value implements driver.Valuer and stores the amount in minor units.
func (m Money) Value() (driver.Value, error) {
	minorUnits, err := m.MinorUnits()
	if err != nil {
		return nil, fmt.Errorf("convert to minor units: %w", err)
	}

	return minorUnits, nil
}

// Scan implements sql.Scanner and reads the amount from minor units.
func (m *Money) Scan(src any) error {
	minorUnits, ok := src.(int64)
	if !ok {
		return fmt.Errorf("%w: %T", errUnsupportedScanType, src)
	}

	*m = NewFromMinorUnits(minorUnits)

	return nil
}
