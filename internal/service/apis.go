package service

type APIs struct {
	Messenger Messenger
}

// Messenger handles messaging operations between the application and messaging platform.
type Messenger interface {
	// ReadUpdates retrieves new incoming updates/messages from the messaging platform.
	ReadUpdates(result chan Message, errors chan error)
	// SendMessage sends a text message to the specified chat.
	SendMessage(chatID int, text string) error
	// SendWithKeyboard sends a message with an attached keyboard (inline or reply).
	SendWithKeyboard(opts SendWithKeyboardOptions) error

	// Close closes the underlying connection to the messaging platform.
	Close() error
}

// SendWithKeyboardOptions represents options for sending a message with a keyboard.
type SendWithKeyboardOptions struct {
	ChatID         int
	Message        string
	Keyboard       []KeyboardRow
	InlineKeyboard []InlineKeyboardRow
}

// KeyboardRow represents keyboard row with buttons.
type KeyboardRow struct {
	Buttons []string
}

// InlineKeyboardRow represents inline keyboard row with buttons.
type InlineKeyboardRow struct {
	Buttons []InlineKeyboardButton
}

// InlineKeyboardButton represents an inline keyboard button with text and data.
type InlineKeyboardButton struct {
	Text string
	Data string
}

// Message represents a message that was received from the messaging platform.
type Message interface {
	// GetChatID returns the ID of the chat the message was sent to.
	GetChatID() int
	// GetText returns the text content of the message.
	GetText() string
	// GetSenderName returns the name of the user who sent the message.
	GetSenderName() string
}
