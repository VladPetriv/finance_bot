package service

import (
	"context"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
)

// Services represents structure with all services.
type Services struct {
	Event   EventService
	Handler HandlerService
	State   StateService
}

// HandlerService provides functionally for handling bot events.
type HandlerService interface {
	// HandleError is used to send the user a message that something went wrong while processing the event.
	HandleError(ctx context.Context, err error, msg Message) error
	// HandleUnknown inform user that provided event is unknown and notify him about available events.
	HandleUnknown(msg Message) error

	// HandleStart initialize new user, his balance and send welcome message.
	HandleStart(ctx context.Context, msg Message) error
	// HandleBack resets user interface to main menu.
	HandleBack(ctx context.Context, msg Message) error
	// HandleWrappers processes main keyboard selections, where each button (Balance/Operations/Categories)
	// maps to corresponding model wrapper to handle its specific actions.
	HandleWrappers(ctx context.Context, event models.Event, msg Message) error

	// HandleBalanceCreate processes new balance entry creation
	HandleBalanceCreate(ctx context.Context, msg Message) error
	// HandleBalanceUpdate processes balance modification
	HandleBalanceUpdate(ctx context.Context, msg Message) error
	// HandleBalanceGet retrieves current balance information
	HandleBalanceGet(ctx context.Context, msg Message) error
	// HandleBalanceDelete processes balance entry removal
	HandleBalanceDelete(ctx context.Context, msg Message) error

	// HandleCategoryCreate processes new category creation
	HandleCategoryCreate(ctx context.Context, msg Message) error
	// HandleCategoryList retrieves all available categories
	HandleCategoryList(ctx context.Context, msg Message) error
	// HandleCategoryUpdate processes category modification
	HandleCategoryUpdate(ctx context.Context, msg Message) error
	// HandleCategoryDelete processes category removal
	HandleCategoryDelete(ctx context.Context, msg Message) error

	// HandleOperationCreate processes new operation creation
	HandleOperationCreate(ctx context.Context, msg Message) error
	// HandleOperationHistory retrieves operation transaction history
	HandleOperationHistory(ctx context.Context, msg Message) error
}

// EventService provides functionally for receiving an updates from bot and reacting on it.
type EventService interface {
	// Listen is used to receive all updates from bot.
	Listen(ctx context.Context)
	// ReactOnEven is used to react on event by his name.
	ReactOnEvent(ctx context.Context, eventName models.Event, msg Message) error
}

type contextFieldName string

const contextFieldNameState contextFieldName = "state"

var defaultKeyboardRows = []KeyboardRow{
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

// StateService represents a service for managing and handling complex bot flow using state.
type StateService interface {
	HandleState(ctx context.Context, message Message) (*HandleStateOutput, error)
}

// HandleStateOutput represents an output structure for StateService.HandleState method.
type HandleStateOutput struct {
	State *models.State
	Event models.Event
}
