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
	// Send is used to send information to user.
	Send(opts SendOptions) error
}

type SendOptions struct {
	ChatID         int64
	Message        string
	Keyboard       []KeyboardRow
	InlineKeyboard []KeyboardRow
}

// KeyboardRow represents keyboard row with buttons.
type KeyboardRow struct {
	Buttons []string
}
