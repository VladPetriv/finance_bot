package models

// Event represents the type of event that occurs during bot interaction
type Event string

const (
	// StartEvent represents the event when user starts interacting with the bot
	StartEvent Event = "start"
	// CancelEvent represents the event when user wants to stop the current floe
	CancelEvent Event = "cancel"
	// UnknownEvent represents an unrecognized or unsupported event
	UnknownEvent Event = "unknown"

	// BalanceEvent represents the event for receiving balance actions
	BalanceEvent Event = "balance/actions"
	// CategoryEvent represents the event for receiving category actions
	CategoryEvent Event = "category/actions"
	// OperationEvent represents the event for receiving operation actions
	OperationEvent Event = "operation/actions"
	// BalanceSubscriptionsEvent represents the event for receiving balance subscriptions actions
	BalanceSubscriptionsEvent Event = "balance/subscriptions/actions"

	// CreateBalanceEvent represents the event for creating a new balance
	CreateBalanceEvent Event = "balance/create"
	// UpdateBalanceEvent represents the event for updating a balance
	UpdateBalanceEvent Event = "balance/update"
	// GetBalanceEvent represents the event for getting a balance
	GetBalanceEvent Event = "balance/get"
	// DeleteBalanceEvent represents the event for deleting a balance
	DeleteBalanceEvent Event = "balance/delete"

	// CreateCategoryEvent represents the event for creating a new category
	CreateCategoryEvent Event = "category/create"
	// ListCategoriesEvent represents the event for listing all categories
	ListCategoriesEvent Event = "category/list"
	// UpdateCategoryEvent represents the event for updating a category
	UpdateCategoryEvent Event = "category/update"
	// DeleteCategoryEvent represents the event for deleting a category
	DeleteCategoryEvent Event = "category/delete"

	// CreateOperationEvent represents the event for creating a new operation
	CreateOperationEvent Event = "operation/create"
	// GetOperationsHistoryEvent represents the event for getting operations history
	GetOperationsHistoryEvent Event = "operation/get_history"
	// DeleteOperationEvent represents the event for deleting an operation
	DeleteOperationEvent Event = "operation/delete"
	// UpdateOperationEvent represents the event for updating an operation
	UpdateOperationEvent Event = "operation/update"
	// CreateOperationsThroughOneTimeInputEvent represents the event for creating operations through one-time input
	CreateOperationsThroughOneTimeInputEvent Event = "operation/create_through_one_time_input"

	// CreateBalanceSubscriptionEvent represents the event for creating a new balance subscription
	CreateBalanceSubscriptionEvent Event = "balance_subscription/create"
	// ListBalanceSubscriptionsEvent represents the event for listing all balance subscriptions
	ListBalanceSubscriptionsEvent Event = "balance_subscription/list"
	// UpdateBalanceSubscriptionEvent represents the event for updating a balance subscription
	UpdateBalanceSubscriptionEvent Event = "balance_subscription/update"
	// DeleteBalanceSubscriptionEvent represents the event for deleting a balance subscription
	DeleteBalanceSubscriptionEvent Event = "balance_subscription/delete"
)

// EventToFlow maps events to their corresponding flows
var EventToFlow = map[Event]Flow{
	// General
	StartEvent:  StartFlow,
	CancelEvent: CancelFlow,

	// Wrappers
	BalanceEvent:   BalanceFlow,
	CategoryEvent:  CategoryFlow,
	OperationEvent: OperationFlow,

	// Balance
	CreateBalanceEvent: CreateBalanceFlow,
	UpdateBalanceEvent: UpdateBalanceFlow,
	DeleteBalanceEvent: DeleteBalanceFlow,
	GetBalanceEvent:    GetBalanceFlow,

	// Category
	CreateCategoryEvent: CreateCategoryFlow,
	ListCategoriesEvent: ListCategoriesFlow,
	UpdateCategoryEvent: UpdateCategoryFlow,
	DeleteCategoryEvent: DeleteCategoryFlow,

	// Operation
	CreateOperationEvent:                     CreateOperationFlow,
	GetOperationsHistoryEvent:                GetOperationsHistoryFlow,
	DeleteOperationEvent:                     DeleteOperationFlow,
	UpdateOperationEvent:                     UpdateOperationFlow,
	CreateOperationsThroughOneTimeInputEvent: CreateOperationsThroughOneTimeInputFlow,

	// Balance subscriptions
	CreateBalanceSubscriptionEvent: CreateBalanceSubscriptionFlow,
	ListBalanceSubscriptionsEvent:  ListBalanceSubscriptionsFlow,
	UpdateBalanceSubscriptionEvent: UpdateBalanceSubscriptionFlow,
	DeleteBalanceSubscriptionEvent: DeleteBalanceSubscriptionFlow,
}
