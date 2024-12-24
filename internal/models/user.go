package models

// User represents an user model.
type User struct {
	ID       string `bson:"_id,omitempty"`
	Username string `bson:"username,omitempty"`

	Balances []Balance `bson:"balances,omitempty"`
}
