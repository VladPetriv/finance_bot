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

// HandlerService provides functionally for handling bot events.
type HandlerService interface {
	// HandleError is used to send the user a message that something went wrong while processing the event.
	HandleError(ctx context.Context, err error, msg botMessage) error
	// HandleUnknown inform user that provided event is unknown and notify him about available events.
	HandleUnknown(msg botMessage) error

	// HandleStart initialize new user, his balance and send welcome message.
	HandleStart(ctx context.Context, msg botMessage) error
	// HandleBack resets user interface to main menu.
	HandleBack(ctx context.Context, msg botMessage) error
	// HandleWrappers processes main keyboard selections, where each button (Balance/Operations/Categories)
	// maps to corresponding model wrapper to handle its specific actions.
	HandleWrappers(ctx context.Context, event models.Event, msg botMessage) error

	// HandleBalanceCreate processes new balance entry creation
	HandleBalanceCreate(ctx context.Context, msg botMessage) error
	// HandleBalanceUpdate processes balance modification
	HandleBalanceUpdate(ctx context.Context, msg botMessage) error
	// HandleBalanceGet retrieves current balance information
	HandleBalanceGet(ctx context.Context, msg botMessage) error
	// HandleBalanceDelete processes balance entry removal
	HandleBalanceDelete(ctx context.Context, msg botMessage) error

	// HandleCategoryCreate processes new category creation
	HandleCategoryCreate(ctx context.Context, msg botMessage) error
	// HandleCategoryList retrieves all available categories
	HandleCategoryList(ctx context.Context, msg botMessage) error
	// HandleCategoryUpdate processes category modification
	HandleCategoryUpdate(ctx context.Context, msg botMessage) error
	// HandleCategoryDelete processes category removal
	HandleCategoryDelete(ctx context.Context, msg botMessage) error

	// HandleOperationCreate processes new operation creation
	HandleOperationCreate(ctx context.Context, msg botMessage) error
	// HandleOperationHistory retrieves operation transaction history
	HandleOperationHistory(ctx context.Context, msg botMessage) error
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
		Buttons: []string{models.BotBalanceCommand, models.BotCategoryCommand},
	},
	{
		Buttons: []string{models.BotOperationCommand},
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

const (
	// Balance related keys
	balanceIDMetadataKey              = "balance_id"
	balanceNameMetadataKey            = "balance_name"
	balanceFromMetadataKey            = "balance_from"
	balanceToMetadataKey              = "balance_to"
	currentBalanceNameMetadataKey     = "current_balance_name"
	currentBalanceCurrencyMetadataKey = "current_balance_currency"
	currentBalanceAmountMetadataKey   = "current_balance_amount"

	// Category related keys
	previousCategoryTitleMetadataKey = "previous_category_title"
	categoryTitleMetadataKey         = "category_title"

	// Operation related keys
	exchangeRateMetadataKey         = "exchange_rate"
	operationDescriptionMetadataKey = "operation_description"
	operationTypeMetadataKey        = "operation_type"
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
