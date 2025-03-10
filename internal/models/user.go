package models

// User represents an user model.
type User struct {
	ID       string `db:"id"`
	Username string `db:"username"`

	Balances []Balance
}

// GetBalancesIDs returns the balances IDs.
func (u *User) GetBalancesIDs() []string {
	ids := make([]string, 0, len(u.Balances))
	for _, balance := range u.Balances {
		ids = append(ids, balance.ID)
	}

	return ids
}

// GetBalance returns the balance by matching the input with a name or ID.
func (u *User) GetBalance(value string) *Balance {
	for _, balance := range u.Balances {
		if balance.Name == value || balance.ID == value {
			return &balance
		}
	}

	return nil
}
