package models

// Category represents a category model.
type Category struct {
	ID    string `bson:"_id,omitempty"`
	Title string `bson:"title,omitempty"`
}
