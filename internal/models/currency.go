package models

// Currency represents currency
type Currency struct {
	ID     string `bson:"_id"`
	Name   string `bson:"name"`
	Code   string `bson:"code"`
	Symbol string `bson:"symbol"`
}
