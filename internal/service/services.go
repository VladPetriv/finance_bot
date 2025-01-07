package service

import (
	"context"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/errs"
)

// Services represents structure with all services.
type Services struct {
	Event    EventService
	Handler  HandlerService
	Message  MessageService
	Keyboard KeyboardService
	State    StateService
}

// HandlerService provides functionally for handling events.
type HandlerService interface {
	// HandleError is used to send the user a message that something went wrong while processing the command.
	HandleError(ctx context.Context, err error, msg botMessage) error
	// HandleEventUnknown is used to handle event unknown.
	HandleEventUnknown(msg botMessage) error
	// HandleEventStart is used to handle event start.
	HandleEventStart(ctx context.Context, msg botMessage) error
	// HandleEventBack is used to reset bot buttons to default mode.
	HandleEventBack(ctx context.Context, msg botMessage) error

	// HandleEventBalanceCreated is used to handle update balance event.
	HandleEventBalanceCreated(ctx context.Context, msg botMessage) error
	// HandleEventBalanceUpdated is used to handle update balance event.
	HandleEventBalanceUpdated(ctx context.Context, msg botMessage) error
	// HandleEventGetBalance is used to handle get balance event.
	HandleEventGetBalance(ctx context.Context, msg botMessage) error

	// HandleEventCategoryCreate is used to handle category created event.
	HandleEventCategoryCreated(ctx context.Context, msg botMessage) error
	// HandleEventListCategories is used to handle lit categories event.
	HandleEventListCategories(ctx context.Context, msg botMessage) error
	// HandleEventCategoryUpdated is used to handle update category event.
	HandleEventCategoryUpdated(ctx context.Context, msg botMessage) error
	// HandleEventCategoryDeleted is used to handle delete category event.
	HandleEventCategoryDeleted(ctx context.Context, msg botMessage) error

	// HandleEventOperationCreated is used to create an operation.
	HandleEventOperationCreated(ctc context.Context, msg botMessage) error
	// HandleEventGetOperationsHistory is used to get operations history.
	HandleEventGetOperationsHistory(ctx context.Context, msg botMessage) error
}

// EventService provides functionally for receiving an updates from bot and reacting on it.
type EventService interface {
	// Listen is used to receive all updates from bot.
	Listen(ctx context.Context)
	// ReactOnEven is used to react on event by his name.
	ReactOnEvent(ctx context.Context, eventName models.Event, msg botMessage) error
}

type contextFieldName string

const contextFieldNameState contextFieldName = "state"

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
		Buttons: []string{models.BotCreateBalanceCommand, models.BotUpdateBalanceCommand, models.BotGetBalanceCommand},
	},
	{
		Buttons: []string{models.BotCreateCategoryCommand, models.BotListCategoriesCommand, models.BotUpdateCategoryCommand},
	},
	{
		Buttons: []string{models.BotDeleteCategoryCommand},
	},
	{
		Buttons: []string{models.BotCreateOperationCommand, models.BotGetOperationsHistory},
	},
}

var (
	// ErrUserAlreadyExists happens when user already exists in system.
	ErrUserAlreadyExists = errs.New("user already exists")
	// ErrUserNotFound happens when user not exists in system.
	ErrUserNotFound = errs.New("User not found")

	// ErrCategoryAlreadyExists happens when try to create category that already exists.
	ErrCategoryAlreadyExists = errs.New("Category already exist. Please use another name.")
	// ErrCategoriesNotFound happens when received zero categories from store.
	ErrCategoriesNotFound = errs.New("Categories not found")
	// ErrCategoryNotFound happens when received not category from store.
	ErrCategoryNotFound = errs.New("Category not found")

	// ErrBalanceNotFound happens when don't receive balance from store.
	ErrBalanceNotFound = errs.New("Balance not found")
	// ErrBalanceAlreadyExists happens when try to create balance that already exists.
	ErrBalanceAlreadyExists = errs.New("Balance already exist. Please use another name.")

	// ErrOperationsNotFound happens when don't receive operations from store.
	ErrOperationsNotFound = errs.New("Operations not found")

	// ErrInvalidAmountFormat happens when use enters amount with invalid format
	ErrInvalidAmountFormat = errs.New("Invalid amount format! Please try again.")

	// ErrInvalidExchangeRateFormat happens when user enters exchange rate with invalid format
	ErrInvalidExchangeRateFormat = errs.New("Invalid exchange rate format! Please try again.")
)

// StateService represents a service for managing and handling complex bot flow using statesstates.
type StateService interface {
	HandleState(ctx context.Context, message botMessage) (*HandleStateOutput, error)
}

// HandleStateOutput represents an output structure for StateService.HandleState method.
type HandleStateOutput struct {
	State *models.State
	Event models.Event
}
