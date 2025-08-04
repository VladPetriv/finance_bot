package service

import (
	"context"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
)

// Services represents structure with all services.
type Services struct {
	Event                     EventService
	Handler                   HandlerService
	State                     StateService
	Currency                  CurrencyService
	BalanceSubscriptionEngine BalanceSubscriptionEngine
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
	// HandleBack returns user to previous menu.
	HandleBack(ctx context.Context, msg Message) error
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
			Buttons: []string{models.BotOperationCommand, models.BotBalanceSubscriptionsCommand},
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
			Buttons: []string{models.BotBackCommand},
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
			Buttons: []string{models.BotBackCommand},
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
			Buttons: []string{models.BotBackCommand},
		},
	}

	balanceSubscriptionKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{models.BotCreateBalanceSubscriptionCommand, models.BotListBalanceSubscriptionsCommand},
		},
		{
			Buttons: []string{models.BotUpdateBalanceSubscriptionCommand, models.BotDeleteBalanceSubscriptionCommand},
		},
		{
			Buttons: []string{models.BotBackCommand},
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

	updateBalanceSubscriptionOptionsKeyboard = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateBalanceSubscriptionNameCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateBalanceSubscriptionAmountCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateBalanceSubscriptionCategoryCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotUpdateBalanceSubscriptionPeriodCommand,
				},
			},
		},
	}
	balanceSubscriptionFrequencyKeyboard = []KeyboardRow{
		{
			Buttons: []string{
				string(models.SubscriptionPeriodWeekly),
				string(models.SubscriptionPeriodMonthly),
				string(models.SubscriptionPeriodYearly),
			},
		},
		{Buttons: []string{models.BotCancelCommand}},
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
	ErrCategoryNotFound = errs.New("Category not found. Please try again!")
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

	// ErrNoBalanceSubscriptionsFound happens when don't receive balance subscriptions from store.
	ErrNoBalanceSubscriptionsFound = errs.New("No balance subscriptions found. Please try to select another balance.")
	// ErrBalanceSubscriptionNotFound happens when don't receive balance subscription from store.
	ErrBalanceSubscriptionNotFound = errs.New("Balance subscription not found. Please try to select another balance subscription.")
)

const (
	// General keys
	baseFlowKey     = "base_flow"
	pageMetadataKey = "page"

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
	categoryIDMetadataKey            = "category_id"

	// Operation related keys
	exchangeRateMetadataKey            = "exchange_rate"
	operationDescriptionMetadataKey    = "operation_description"
	operationAmountMetadataKey         = "operation_amount"
	operationTypeMetadataKey           = "operation_type"
	operationIDMetadataKey             = "operation_id"
	operationCreationPeriodMetadataKey = "operation_creation_period"

	// Balance subscription related keys
	balanceSubscriptionIDMetadataKey     = "balance_subscription_id"
	balanceSubscriptionNameMetadataKey   = "balance_subscription_name"
	balanceSubscriptionPeriodMetadataKey = "balance_subscription_period"
	balanceSubscriptionAmountMetadataKey = "balance_subscription_amount"
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

// BalanceSubscriptionEngine represents a service for processing balance subscriptions and operation creations based on their details.
type BalanceSubscriptionEngine interface {
	// ScheduleOperationsCreation creates scheduled operation entries for a balance subscription.
	// It generates future operation dates based on the subscription's frequency (period) and start date:
	//   - For weekly/monthly frequencies: schedules operations for the next quarter (3 months)
	//   - For yearly frequencies: schedules operations for the next two year
	ScheduleOperationsCreation(ctx context.Context, balanceSubscription models.BalanceSubscription)
	// ExtendScheduledOperations creates additional scheduled operations for an active balance subscription.
	// When a subscription reaches its last scheduled operation date, this method extends the timeline by
	// generating new scheduled operations for the upcoming billing period (quarter/year).
	// It only executes when the subscription is active and has reached its final scheduled operation.
	ExtendScheduledOperations(ctx context.Context)
	// CreateOperations creates operations based on balance subscriptions details.
	CreateOperations(ctx context.Context)
	// NotifyAboutSubscriptionPayment sends a notification a day before subscription payment.
	NotifyAboutSubscriptionPayment(ctx context.Context)
}
