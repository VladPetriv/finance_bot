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
	// CreateRowKeyboard ...
	CreateKeyboard(keyboardType keyboardType, opts *bot.SendOptions)
}

type keyboardType string

const (
	keyboardTypeInline keyboardType = "inline"
	keyboardTypeRow    keyboardType = "row"
)
