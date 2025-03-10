package models

// Category represents a category model.
type Category struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`
	Title  string `db:"title"`
}

// GetName returns the category title.
func (c Category) GetName() string {
	return c.Title
}
