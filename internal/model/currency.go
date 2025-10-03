package model

import "fmt"

// Currency represents currency model which contains currency name, code and symbol
type Currency struct {
	ID     string `db:"id"`
	Name   string `db:"name"`
	Code   string `db:"code"`
	Symbol string `db:"symbol"`
}

// GetID returns the currency data
func (c Currency) GetID() string {
	return c.ID
}

// GetName returns the currency text
func (c Currency) GetName() string {
	return fmt.Sprintf("%s (%s)", c.Name, c.Code)
}
