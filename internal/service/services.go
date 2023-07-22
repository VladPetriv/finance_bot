package service

import (
	"context"
	"errors"
	"strings"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
)

// HandlerService provides functionally for handling events.
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
	// HandleEventOperationCreate is used to create an operation without amount.
	HandleEventOperationCreate(ctc context.Context, eventName event, messageData []byte) error
	// HandleEventUpdateOperationAmount get last transaction with empty amount from db and update his amount with user one.
	HandleEventUpdateOperationAmount(ctx context.Context, messageData []byte) error
	// HandleEventBack is used to reset bot buttons to default mode.
	HandleEventBack(ctx context.Context, messageData []byte) error
}

// HandleEventOperationCreate represents structure with all required info
// about message that needed for handling this event.

// EventService provides functionally for receiving an updates from bot and reacting on it.
type EventService interface {
	// Listen is used to receive all updates from bot.
	Listen(ctx context.Context, updates chan []byte, errs chan error)
	// ReactOnEven is used to react on event by his name.
	ReactOnEvent(ctx context.Context, eventName event, messageData []byte) error
}

type botMessage struct {
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

// GetUsername is used to get actual username from message.
func (h botMessage) GetUsername() string {
	if h.Message.From.Username != "" {
		return h.Message.From.Username
	}

	if h.CallbackQuery.From.Username != "" {
		return h.CallbackQuery.From.Username
	}

	return ""
}

// GetChatID is used to get actual chat id from message.
func (h botMessage) GetChatID() int64 {
	if h.Message.Chat.ID != 0 {
		return h.Message.Chat.ID
	}

	if h.CallbackQuery.Message.Chat.ID != 0 {
		return h.CallbackQuery.Message.Chat.ID
	}

	return 0
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
	botBackCommand                    string = "Back ❌"
	botCreateCategoryCommand          string = "Create Category 📊"
	botListCategoriesCommand          string = "List Categories 🗂️"
	botUpdateBalanceCommand           string = "Update Balance 💲"
	botUpdateBalanceAmountCommand     string = "Update Balance Amount 💵"
	botUpdateBalanceCurrencyCommand   string = "Update Balance Currency 💱"
	botGetBalanceCommand              string = "Get Balance Info 🏦"
	botCreateOperationCommand         string = "Create Operation 🤔"
	botCreateIncomingOperationCommand string = "Create Incoming Operation 🤑"
	botCreateSpendingOperationCommand string = "Create Spending Operation 💸"
	botUpdateOperationAmountCommand   string = "Update Operation Amount"
)

var availableCommands = []string{
	botStartCommand, botBackCommand, botCreateCategoryCommand,
	botListCategoriesCommand, botUpdateBalanceCommand, botUpdateBalanceAmountCommand,
	botCreateOperationCommand, botUpdateBalanceCurrencyCommand, botGetBalanceCommand, botCreateIncomingOperationCommand,
	botCreateIncomingOperationCommand, botCreateSpendingOperationCommand, botUpdateOperationAmountCommand,
}

// IsBotCommand is used to determine if incoming text a bot command or not.
func IsBotCommand(command string) bool {
	return strings.Contains(strings.Join(availableCommands, " "), command)
}

var commandToEvent = map[string]event{
	botStartCommand:                   startEvent,
	botBackCommand:                    backEvent,
	botCreateCategoryCommand:          createCategoryEvent,
	botListCategoriesCommand:          listCategoryEvent,
	botUpdateBalanceCommand:           updateBalanceEvent,
	botUpdateBalanceAmountCommand:     updateBalanceAmountEvent,
	botUpdateBalanceCurrencyCommand:   updateBalanceCurrencyEvent,
	botGetBalanceCommand:              getBalanceEvent,
	botCreateOperationCommand:         createOperationEvent,
	botCreateIncomingOperationCommand: createIncomingOperationEvent,
	botCreateSpendingOperationCommand: createSpendingOperationEvent,
	botUpdateOperationAmountCommand:   updateOperationAmountEvent,
}

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
		Buttons: []string{botCreateCategoryCommand, botListCategoriesCommand},
	},
	{
		Buttons: []string{botGetBalanceCommand, botUpdateBalanceCommand},
	},
	{
		Buttons: []string{botCreateOperationCommand},
	},
}

// UserService provides business logic for work with users.
type UserService interface {
	// CreateUser is used to create user if it's not exists.
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
	// CreateCategory creates a new category.
	CreateCategory(ctx context.Context, category *models.Category) error
	// ListCategories fetches a list of categories for the specified user ID.
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
