package service

import "github.com/VladPetriv/finance_bot/pkg/bot"

// Services contains all Services.
type Services struct {
	MessageService  MessageService
	KeyboardService KeyboardService
	EventService    EventService
}

// EventService provides functinally for receiving an updates from bot and reacting on it.
type EventService interface {
	// Listen is used to receive all updates from bot and react for them.
	Listen(updates chan []byte, errs chan error)
}

// BaseMessage represents a message with not detailed information.
// BaseMessage is used to determine which command to do.
type BaseMessage struct {
	Message struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text     string   `json:"text"`
		Entities []Entity `json:"entities"`
	} `json:"message"`
}

// Entity represents message entity that contains about message type.
type Entity struct {
	Type string `json:"type"`
}

type event string

const (
	startEvent   event = "start"
	stopEvent    event = "stop"
	unknownEvent event = "unknown"
)

// Commands that we can received from bot.
const (
	botStartCommand string = "/start"
	botStopCommand  string = "/stop"
)

// MessageService provides functinally for sending messages.
type MessageService interface {
	// SendMessage is used to send messages for specific chat.
	SendMessage(opts *SendMessageOptions) error
}

// SendMessageOptions represents input structure for CreateKeyboard method.
type SendMessageOptions struct {
	ChantID int64
	Text    string
}

// KeyboardService provides functinally rendering keyboard.
type KeyboardService interface {
	// CreateRowKeyboard is used to create all available keyboard.
	CreateKeyboard(opts *CreateKeyboardOptions) error
}

// CreateKeyboardOptions represents input structure for CreateKeyboard method.
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
