package models

// Category represents a category model.
type Category struct {
	ID     string `bson:"_id,omitempty"`
	UserID string `bson:"userid,omitempty"`
	Title  string `bson:"title,omitempty"`
}

// GetName returns the category title.
func (c Category) GetName() string {
	return c.Title
}
