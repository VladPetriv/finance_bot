package models

// User represents an user model.
type User struct {
	ID       string `bson:"_id,omitempty"`
	Username string `bson:"username,omitempty"`

	Balances []Balance `bson:"balances,omitempty"`
}

// GetBalance returns the balance by the given name.
func (u *User) GetBalance(name string) *Balance {
	for _, balance := range u.Balances {
		if balance.Name == name {
			return &balance
		}
	}

	return nil
}
