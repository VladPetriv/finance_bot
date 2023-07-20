package service

import (
	"context"
	"errors"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
)

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
	// HandleEventListCategories is used to handle update balance event.
	HandleEventUpdateBalance(ctx context.Context, eventName event, messageData []byte) error
	// HandleEventGetBalance is used to handle get balance event.
	HandleEventGetBalance(ctx context.Context, messageData []byte) error

	HandleEventOperationCreate(ctc context.Context, eventName event, messageData []byte) error

	HandleEventUpdateOperationAmount(ctx context.Context, messageData []byte) error

	// HandleEventBack is used to reset bot buttons to default mode.
	HandleEventBack(ctx context.Context, messageData []byte) error
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
		Chat chat   `json:"chat"`
		From from   `json:"from"`
		Text string `json:"text"`
	} `json:"message"`
}

// HandleEventUpdateBalance represents structure with all required info
// about message that needed for handling this event.
type HandleEventUpdateBalance struct {
	Message struct {
		Chat chat   `json:"chat"`
		From from   `json:"from"`
		Text string `json:"text"`
	} `json:"message"`
}

// HandleEventUpdateOperationAmount represents structure with all required info
// about message that needed for handling this event.
type HandleEventUpdateOperationAmount struct {
	Message struct {
		Chat chat   `json:"chat"`
		From from   `json:"from"`
		Text string `json:"text"`
	} `json:"message"`
}

// HandleEventGetBalance represents structure with all required info
// about message that needed for handling this event.
type HandleEventGetBalance struct {
	Message struct {
		Chat chat `json:"chat"`
		From from `json:"from"`
	} `json:"message"`
}

// HandleEventListCategories represents structure with all required info
// about message that needed for handling this event.
type HandleEventListCategories struct {
	Message struct {
		Chat chat `json:"chat"`
		From from `json:"from"`
	} `json:"message"`
}

// HandleEventOperationCreate represents structure with all required info
// about message that needed for handling this event.
type HandleEventOperationCreate struct {
	Message struct {
		Chat chat   `json:"chat"`
		From from   `json:"from"`
		Text string `json:"text"`
	} `json:"message"`
	CallbackQuery struct {
		ID      string `json:"id"`
		From    from   `json:"from"`
		Message struct {
			Chat chat `json:"chat"`
		} `json:"message"`
		Data string `json:"data"`
	} `json:"callback_query"`
}

// GetUsername is used to get user from message.
func (h HandleEventOperationCreate) GetUsername() string {
	if h.Message.From.Username != "" {
		return h.Message.From.Username
	}

	if h.CallbackQuery.From.Username != "" {
		return h.CallbackQuery.From.Username
	}

	return ""
}

// GetChatID is used to get chat id from message.
func (h HandleEventOperationCreate) GetChatID() int64 {
	if h.Message.Chat.ID != 0 {
		return h.Message.Chat.ID
	}

	if h.CallbackQuery.Message.Chat.ID != 0 {
		return h.CallbackQuery.Message.Chat.ID
	}

	return 0
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
		Chat chat   `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
	CallbackQuery struct {
		Data string `json:"data"`
	} `json:"callback_query"`
}

type chat struct {
	ID int64 `json:"id"`
}

type from struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type event string

const (
	startEvent                   event = "start"
	createCategoryEvent          event = "create/category"
	listCategoryEvent            event = "list/categories"
	updateBalanceEvent           event = "update/balance"
	updateBalanceAmountEvent     event = "update/balance/amount"
	updateBalanceCurrencyEvent   event = "update/balance/currency"
	getBalanceEvent              event = "get/balance"
	createOperationEvent         event = "create/operation"
	createIncomingOperationEvent event = "create/incoming/operation"
	createSpendingOperationEvent event = "create/spending/operation"
	updateOperationAmountEvent   event = "update/operation/amount"
	backEvent                    event = "back"
	unknownEvent                 event = "unknown"
)

var eventsWithInput = map[event]int{
	createCategoryEvent:          1,
	updateBalanceAmountEvent:     1,
	updateBalanceCurrencyEvent:   1,
	createIncomingOperationEvent: 1,
	createSpendingOperationEvent: 1,
	updateOperationAmountEvent:   1,
}

// Commands that we can received from bot.
const (
	botStartCommand                   string = "/start"
	botBackCommand                    string = "/back"
	botCreateCategoryCommand          string = "/create_category"
	botListCategoriesCommand          string = "/list-categories"
	botUpdateBalanceCommand           string = "/update-balance"
	botUpdateBalanceAmountCommand     string = "/update_balance_amount"
	botUpdateBalanceCurrencyCommand   string = "/update_balance_currency"
	botGetBalanceCommand              string = "/get_balance_info"
	botCreateOperationCommand         string = "/create_operation"
	botCreateIncomingOperationCommand string = "/create_incoming_operation"
	botCreateSpendingOperationCommand string = "/create_spending_operation"
	botUpdateOperationAmountCommand   string = "/update_operation_amount"
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
	{
		Buttons: []string{"/get_balance_info", "/update-balance"},
	},
	{
		Buttons: []string{botCreateOperationCommand},
	},
}

// UserService provides business logic for work with users.
type UserService interface {
	// CreateUser is used to create user if it's not exists..
	CreateUser(ctx context.Context, user *models.User) error
	// GetUserByUsername is used to get user by his username.
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

var (
	// ErrUserAlreadyExists happens when user already exists in system.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrUserNotFound happens when user not exists in system.
	ErrUserNotFound = errors.New("user not found")
)

// CategoryService provides business logic for processing categories.
type CategoryService interface {
	CreateCategory(ctx context.Context, category *models.Category) error
	ListCategories(ctx context.Context, userID string) ([]models.Category, error)
}

var (
	// ErrCategoryAlreadyExists happens when try to create category that already exists.
	ErrCategoryAlreadyExists = errors.New("category already exist")
	// ErrCategoriesNotFound happens when received zero categories from store.
	ErrCategoriesNotFound = errors.New("categories not found")
	// ErrCategoryNotFound happens when received not category from store.
	ErrCategoryNotFound = errors.New("categoriy not found")
)

// BalanceService provides business logic for processing balance.
type BalanceService interface {
	// 	GetBalanceInfo is used to get all balance related information by user id.
	GetBalanceInfo(ctx context.Context, userID string) (*models.Balance, error)
}

// ErrBalanceNotFound happens when don't receive balance from store.
var ErrBalanceNotFound = errors.New("balance not found")

// OperationService provides business logic for work with balance operations.
type OperationService interface {
	// CreateOperation is used to create new operation with change of user balance amount.
	CreateOperation(ctx context.Context, opts CreateOperationOptions) error
}

// CreateOperationOptions represents an input values for creating new operation.
type CreateOperationOptions struct {
	UserID    string
	Operation *models.Operation
}

// ErrInvalidAmountFormat happens when use enters amount with invalid format
var ErrInvalidAmountFormat = errors.New("invalid amount format")
