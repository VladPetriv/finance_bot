package models

// User represents an user model.
type User struct {
	ID       string `bson:"_id,omitempty"`
	Username string `bson:"username,omitempty"`

	Balances []Balance `bson:"balances,omitempty"`
}

// GetBalancesIDs returns the balances IDs.
func (u *User) GetBalancesIDs() []string {
	ids := make([]string, 0, len(u.Balances))
	for _, balance := range u.Balances {
		ids = append(ids, balance.ID)
	}

	return ids
}

// GetBalance returns the balance by trying to match it by input value with the name or an id.
func (u *User) GetBalance(value string) *Balance {
	for _, balance := range u.Balances {
		if balance.Name == value || balance.ID == value {
			return &balance
		}
	}

	return nil
}
