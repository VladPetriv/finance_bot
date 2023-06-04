package service

import (
	"context"
	"errors"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
)

// Services contains all Services.
type Services struct {
	MessageService  MessageService
	KeyboardService KeyboardService
	HandlerService  HandlerService
	EventService    EventService
	CategoryService CategoryService
	UserService     UserService
}

// HandlerService provides functionally for handling bot commands.
type HandlerService interface {
	// HandleEventStart is used to handle event start.
	HandleEventStart(ctx context.Context, messageData []byte) error
	// HandleEventUnknown is used to handle event unknown.
	HandleEventUnknown(messageData []byte) error
	// HandleEventCategoryCreate is used to handle category created event.
	HandleEventCategoryCreate(ctx context.Context, messageData []byte) error
	// HandleEventListCategories is used to handle lit categories event.
	HandleEventListCategories(ctx context.Context, messageData []byte) error
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

// HandleEventCategoryCreate represents structure with all required info
// about message that needed for handling this event.
type HandleEventCategoryCreate struct {
	Message struct {
		Chat     chat     `json:"chat"`
		Entities []entity `json:"entities"`
		Text     string   `json:"text"`
	} `json:"message"`
}

// HandleEventListCategories represents structure with all required info
// about message that needed for handling this event.
type HandleEventListCategories struct {
	Message struct {
		Chat chat `json:"chat"`
	} `json:"message"`
}

// EventService provides functionally for receiving an updates from bot and reacting on it.
type EventService interface {
	// Listen is used to receive all updates from bot and react for them.
	Listen(ctx context.Context, updates chan []byte, errs chan error)
	// ReactOnEven is used to
	ReactOnEvent(ctx context.Context, eventName event, messageData []byte) error
}

// BaseMessage represents a message with not detailed information.
// BaseMessage is used to determine which command to do.
type BaseMessage struct {
	Message struct {
		Chat     chat     `json:"chat"`
		Text     string   `json:"text"`
		Entities []entity `json:"entities"`
	} `json:"message"`
}

type chat struct {
	ID int64 `json:"id"`
}

type from struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type entity struct {
	Type string `json:"type"`
}

func (e entity) IsBotCommand() bool {
	return e.Type == botCommand
}

type event string

const (
	startEvent          event = "start"
	createCategoryEvent event = "create/category"
	listCategoryEvent   event = "list/categories"
	unknownEvent        event = "unknown"
)

var eventsWithInput = map[event]bool{
	createCategoryEvent: true,
}

// Commands that we can received from bot.
const (
	botStartCommand          string = "/start"
	botCreateCategoryCommand string = "/create_category"
	botListCategoriesCommand string = "/list-categories"
)

// MessageService provides functionally for sending messages.
type MessageService interface {
	// SendMessage is used to send messages for specific chat.
	SendMessage(opts *SendMessageOptions) error
}

// SendMessageOptions represents input structure for CreateKeyboard method.
type SendMessageOptions struct {
	ChatID int64
	Text   string
}

// KeyboardService provides functionally rendering keyboard.
type KeyboardService interface {
	// CreateRowKeyboard is used to create keyboard and send message with it..
	CreateKeyboard(opts *CreateKeyboardOptions) error
}

// CreateKeyboardOptions represents input structure for CreateKeyboard method.
type CreateKeyboardOptions struct {
	ChatID  int64
	Message string
	Type    KeyboardType
	Rows    []bot.KeyboardRow
}

// KeyboardType represents available keyboard types.
type KeyboardType string

const (
	keyboardTypeInline KeyboardType = "inline"
	keyboardTypeRow    KeyboardType = "row"
)

var defaultKeyboardRows = []bot.KeyboardRow{
	{
		Buttons: []string{"/create_category", "/list-categories"},
	},
}

// UserService provides business logic for work with users.
type UserService interface {
	// CreateUser is used to create user if it's not exists..
	CreateUser(ctx context.Context, user *models.User) error
}

// ErrUserAlreadyExists happens when user already exists in system.
var ErrUserAlreadyExists = errors.New("user already exists")

// CategoryService provides business logic for processing categories.
type CategoryService interface {
	CreateCategory(ctx context.Context, category *models.Category) error
	ListCategories(ctx context.Context) ([]models.Category, error)
}

var (
	// ErrCategoryAlreadyExists happens when try to create category that already exists.
	ErrCategoryAlreadyExists = errors.New("category already exist")
	// ErrCategoriesNotFound happens when received zero categories from store.
	ErrCategoriesNotFound = errors.New("categories not found")
)
