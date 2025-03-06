package models

// Currency represents currency model which contains currency name, code and symbol
type Currency struct {
	ID     string `db:"id"`
	Name   string `db:"name"`
	Code   string `db:"code"`
	Symbol string `db:"symbol"`
}
