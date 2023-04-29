package service

import "github.com/VladPetriv/finance_bot/pkg/bot"

// Services contains all Services.
type Services struct {
	EventService   EventService
	MessageService MessageService
}

// EventService provides functinally for receiving an updates from bot and reacting on it.
type EventService interface {
	// Listen is used to receive all updates from bot and react for them.
	Listen() error
}

// MessageService provides functinally for sending messages.
type MessageService interface {
	// SendMessage is used to send messages for specific chat.
	SendMessage(chatID int64, message string) error
}

// KeyboardService provides functinally rendering keyboard.
type KeyboardService interface {
	// CreateRowKeyboard is used to create all available keyboard.
	CreateKeyboard(opts *CreateKeyboardOptions) error
}

// CreateKeyboardOptions represents input structure for CreateKeyboard method
type CreateKeyboardOptions struct {
	ChatID int64
	Type   KeyboardType
	Rows   []bot.KeyboardRow
}

// KeyboardType represents available keyboard types.
type KeyboardType string

const (
	keyboardTypeInline KeyboardType = "inline"
	keyboardTypeRow    KeyboardType = "row"
)
