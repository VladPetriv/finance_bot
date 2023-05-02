package service

import "github.com/VladPetriv/finance_bot/pkg/bot"

// Services contains all Services.
type Services struct {
	MessageService  MessageService
	KeyboardService KeyboardService
	HandlerService  HandlerService
	EventService    EventService
}

// HandlerService provides functinally for handling bot commands.
type HandlerService interface {
	// HandleEventStart is used to handle event start.
	HandleEventStart(messageData []byte) error
	// HandleEventUnknown is used to handle event unknown.
	HandleEventUnknown(messageData []byte) error
}

// HandleEventStartMessage represents structure with all required info
// about message that needed for handling this event.
type HandleEventStartMessage struct {
	Message struct {
		Chat chat `json:"chat"`
		From from `json:"from"`
	} `json:"message"`
}

// HandleEventUnknownMessage represents structure with all required info
// about message that needed for handling this event.
type HandleEventUnknownMessage struct {
	Message struct {
		Chat chat `json:"chat"`
	} `json:"message"`
}

// EventService provides functinally for receiving an updates from bot and reacting on it.
type EventService interface {
	// Listen is used to receive all updates from bot and react for them.
	Listen(updates chan []byte, errs chan error)
	// ReactOnEven is used to
	ReactOnEvent(eventName event, messageData []byte) error
}

// BaseMessage represents a message with not detailed information.
// BaseMessage is used to determine which command to do.
type BaseMessage struct {
	Message struct {
		Chat     chat     `json:"chat"`
		Text     string   `json:"text"`
		Entities []Entity `json:"entities"`
	} `json:"message"`
}

type chat struct {
	ID int64 `json:"id"`
}

type from struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// Entity represents message entity that contains about message type.
type Entity struct {
	Type string `json:"type"`
}

type event string

const (
	startEvent   event = "start"
	unknownEvent event = "unknown"
)

// Commands that we can received from bot.
const (
	botStartCommand string = "/start"
)

// MessageService provides functinally for sending messages.
type MessageService interface {
	// SendMessage is used to send messages for specific chat.
	SendMessage(opts *SendMessageOptions) error
}

// SendMessageOptions represents input structure for CreateKeyboard method.
type SendMessageOptions struct {
	ChatID int64
	Text   string
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
