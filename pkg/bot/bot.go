package bot

// Bot represents a bot interface.
type Bot interface {
	// NewAPI returns a new instance of bot api.
	NewAPI() (API, error)
}

// API provides functionally to work with bot API.
type API interface {
	// ReadUpdates reads incoming updates from bot API and write them to channel.
	ReadUpdates(result chan []byte, errors chan error)
	// Send sends any information to the user(messages, keyboards).
	Send(opts *SendOptions) error
	// Close closes the connection with the bot API.
	Close() error
}

// SendOptions represents an input structure for Send method.
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
