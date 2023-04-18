package bot

// Bot represents a bot interface.
type Bot interface {
	// NewAPI returns a new instance of bot api.
	NewAPI() (API, error)
}

// API provides funtionally to work with bot API.
type API interface {
	// ReadUpdates is used to get all user actions.
	ReadUpdates(result chan []byte, errors chan error)
	// SendMessage is used to send specific messages to user.
	SendMessage(chatID int64, message string) error
}
