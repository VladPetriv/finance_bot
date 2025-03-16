package service

import (
	"context"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
)

// Services represents structure with all services.
type Services struct {
	Event    EventService
	Handler  HandlerService
	State    StateService
	Currency CurrencyService
}

// HandlerService provides functionally for handling bot events.
type HandlerService interface {
	// RegisterHandlers registers handlers for all flows and steps.
	RegisterHandlers()

	// HandleError is used to send the user a message that something went wrong while processing the event.
	HandleError(ctx context.Context, opts HandleErrorOptions) error
	// HandleUnknown inform user that provided event is unknown and notify him about available events.
	HandleUnknown(msg Message) error

	// HandleStart initialize new user, his balance and send welcome message.
	HandleStart(ctx context.Context, msg Message) error
	// HandleCancel cancel current user flow and returns the default keyboard
	HandleCancel(ctx context.Context, msg Message) error
	// HandleWrappers processes main keyboard selections, where each button (Balance/Operations/Categories)
	// maps to corresponding model wrapper to handle its specific actions.
	HandleWrappers(ctx context.Context, event models.Event, msg Message) error

	// HandleAction process user actions with app entities (balance, category, operation).
	HandleAction(ctx context.Context, msg Message) error
}

type flowProcessingOptions struct {
	user          *models.User
	message       Message
	state         *models.State
	stateMetaData map[string]any
}

// HandleErrorOptions represents input structure for HandleError method.
type HandleErrorOptions struct {
	Err                 error
	Msg                 Message
	SendDefaultKeyboard bool
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

var (
	defaultKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{models.BotBalanceCommand, models.BotCategoryCommand},
		},
		{
			Buttons: []string{models.BotOperationCommand},
		},
	}

	rowKeyboardWithCancelButtonOnly = []KeyboardRow{
		{
			Buttons: []string{models.BotCancelCommand},
		},
	}

	balanceKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{models.BotCreateBalanceCommand, models.BotGetBalanceCommand},
		},
		{
			Buttons: []string{models.BotUpdateBalanceCommand, models.BotDeleteBalanceCommand},
		},
		{
			Buttons: []string{models.BotCancelCommand},
		},
	}

	categoryKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{models.BotCreateCategoryCommand, models.BotListCategoriesCommand},
		},
		{
			Buttons: []string{models.BotUpdateCategoryCommand, models.BotDeleteCategoryCommand},
		},
		{
			Buttons: []string{models.BotCancelCommand},
		},
	}

	operationKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{models.BotCreateOperationCommand, models.BotGetOperationsHistory},
		},
		{
			Buttons: []string{models.BotUpdateOperationCommand, models.BotDeleteOperationCommand},
		},
		{
			Buttons: []string{models.BotCancelCommand},
		},
	}

	updateOperationOptionsKeyboardForIncomingAndSpendingOperations = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateOperationAmountCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateOperationDescriptionCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateOperationCategoryCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateOperationDateCommand,
				},
			},
		},
	}

	updateOperationOptionsKeyboardForTransferOperations = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateOperationAmountCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateOperationDateCommand,
				},
			},
		},
	}

	updateBalanceOptionsKeyboard = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateBalanceNameCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateBalanceAmountCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateBalanceCurrencyCommand,
				},
			},
		},
	}
)

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
	// ErrNotEnoughCategories happens when received 0 categories after filtering.
	ErrNotEnoughCategories = errs.New("Not enough categories.")

	// ErrBalanceNotFound happens when don't receive balance from store.
	ErrBalanceNotFound = errs.New("Balance not found")
	// ErrBalanceAlreadyExists happens when try to create balance that already exists.
	ErrBalanceAlreadyExists = errs.New("Balance already exist. Please use another name.")

	// ErrOperationsNotFound happens when don't receive operations from store.
	ErrOperationsNotFound = errs.New("Operations not found")
	// ErrOperationNotFound happens when don't receive operation from store.
	ErrOperationNotFound = errs.New("Operation not found. Please try to select another operation.")

	// ErrInvalidAmountFormat happens when use enters amount with invalid format
	ErrInvalidAmountFormat = errs.New("Invalid amount format! Please try again.")
	// ErrInvalidDateFormat happens when user enters date with invalid format
	ErrInvalidDateFormat = errs.New("Invalid date format! Please try again.")

	// ErrInvalidExchangeRateFormat happens when user enters exchange rate with invalid format
	ErrInvalidExchangeRateFormat = errs.New("Invalid exchange rate format! Please try again.")

	// ErrCurrencyNotFound happens when don't receive currency from store.
	ErrCurrencyNotFound = errs.New("Currency not found. Please try to select another currency.")
)

const (
	// Balance related keys
	balanceIDMetadataKey              = "balance_id"
	balanceNameMetadataKey            = "balance_name"
	balanceAmountMetadataKey          = "balance_amount"
	balanceFromMetadataKey            = "balance_from"
	balanceToMetadataKey              = "balance_to"
	currentBalanceNameMetadataKey     = "current_balance_name"
	currentBalanceCurrencyMetadataKey = "current_balance_currency"
	currentBalanceAmountMetadataKey   = "current_balance_amount"
	monthForBalanceStatisticsKey      = "month_for_balance_statistics"

	// Category related keys
	previousCategoryTitleMetadataKey = "previous_category_title"
	categoryTitleMetadataKey         = "category_title"

	// Operation related keys
	exchangeRateMetadataKey         = "exchange_rate"
	operationDescriptionMetadataKey = "operation_description"
	operationTypeMetadataKey        = "operation_type"
	lastOperationDateMetadataKey    = "last_operation_date"
	operationIDMetadataKey          = "operation_id"
)

// StateService represents a service for managing and handling complex bot flow using state.
type StateService interface {
	HandleState(ctx context.Context, message Message) (*HandleStateOutput, error)
	DeleteState(ctx context.Context, message Message) error
}

// HandleStateOutput represents an output structure for StateService.HandleState method.
type HandleStateOutput struct {
	State *models.State
	Event models.Event
}

// CurrencyService represents a service for managing and handling currencies.
type CurrencyService interface {
	// InitCurrencies is used to initialize currencies by parsing them from the CurrencyExchanger API and saving them to the database.
	InitCurrencies(ctx context.Context) error
	// Convert is used to convert operations amount from base currency to target currency
	Convert(ctx context.Context, opts ConvertCurrencyOptions) (*money.Money, error)
}

// ConvertCurrencyOptions represents options for converting currency.
type ConvertCurrencyOptions struct {
	BaseCurrency   string
	TargetCurrency string
	Amount         money.Money
}
