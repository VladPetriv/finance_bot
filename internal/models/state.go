package models

type State struct {
	ID     string `bson:"_id"`
	UserID string `bson:"userId"`

	Flow  Flow     `bson:"flow"`
	Steps []string `bson:"steps"`

	CreatedAt int64 `bson:"createdAt"`
	UpdatedAt int64 `bson:"updatedAt"`
}

func (s State) IsFinished() bool {
	return false
}

type Flow string

const (
	StartFlow Flow = "start"
)
