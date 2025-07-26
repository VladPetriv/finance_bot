package service

import (
	"context"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
)

// Stores represents all stores.
type Stores struct {
	Balance             BalanceStore
	Operation           OperationStore
	Category            CategoryStore
	User                UserStore
	State               StateStore
	Currency            CurrencyStore
	BalanceSubscription BalanceSubscriptionStore
}

// UserStore provides functionality for work with users store.
//
//go:generate mockery --dir . --name UserStore --output ./mocks
type UserStore interface {
	// Create creates a new user in store.
	Create(ctx context.Context, user *models.User) error
	// GetByUsername returns a user from store by username.
	Get(ctx context.Context, filters GetUserFilter) (*models.User, error)
	// CreateSettings creates a new user settings in store.
	CreateSettings(ctx context.Context, settings *models.UserSettings) error
}

// GetUserFilter represents a filters for GetUser method.
type GetUserFilter struct {
	Username        string
	BalanceID       string
	PreloadBalances bool
	PreloadSettings bool
}

// BalanceStore provides functionality for work with balance store.
//
//go:generate mockery --dir . --name BalanceStore --output ./mocks
type BalanceStore interface {
	// Create creates a new balance in store.
	Create(ctx context.Context, balance *models.Balance) error
	// Get returns a balance from store by user id.
	Get(ctx context.Context, filter GetBalanceFilter) (*models.Balance, error)
	// Update updates balance model in store.
	Update(ctx context.Context, balance *models.Balance) error
	// Delete deletes balance from store.
	Delete(ctx context.Context, balanceID string) error
}

// GetBalanceFilter represents a filters for GetBalance method.
type GetBalanceFilter struct {
	Name            string
	UserID          string
	BalanceID       string
	PreloadCurrency bool
}

// OperationStore provides functionality for work with operation store.
//
//go:generate mockery --dir . --name OperationStore --output ./mocks
type OperationStore interface {
	// Create creates a new operation.
	Create(ctx context.Context, operation *models.Operation) error
	// Get returns a operation from store by filter.
	Get(ctx context.Context, filter GetOperationFilter) (*models.Operation, error)
	// Count returns a count of operations from store by filter.
	Count(ctx context.Context, filter ListOperationsFilter) (int, error)
	// GetAll returns all operations from store by balance id.
	List(ctx context.Context, filter ListOperationsFilter) ([]models.Operation, error)
	// Update updates an operation in store.
	Update(ctx context.Context, operationID string, operation *models.Operation) error
	// Delete delete operation by his id.
	Delete(ctx context.Context, operationID string) error
}

// GetOperationFilter represents a filters for Get operation method.
type GetOperationFilter struct {
	ID           string
	Amount       string
	Type         models.OperationType
	CreateAtFrom time.Time
	CreateAtTo   time.Time
	BalanceIDs   []string
}

// ListOperationsFilter represents filters for list operations from store.
type ListOperationsFilter struct {
	BalanceID           string
	CreationPeriod      models.CreationPeriod
	Month               models.Month
	Limit               int
	CreatedAtLessThan   time.Time
	SortByCreatedAtDesc bool
}

// CategoryStore provides functionality for work with categories store.
//
//go:generate mockery --dir . --name CategoryStore --output ./mocks
type CategoryStore interface {
	// List returns a list of all categories from store.
	List(ctx context.Context, filters *ListCategoriesFilter) ([]models.Category, error)
	// Get returns a category by fileers.
	Get(ctx context.Context, filter GetCategoryFilter) (*models.Category, error)
	// Create creates new category in store.
	Create(ctx context.Context, category *models.Category) error
	// Update  updates category in store.
	Update(ctx context.Context, category *models.Category) error
	// Delete delete category from store.
	Delete(ctx context.Context, categoryID string) error
}

// ListCategoriesFilter represents a filters for GetAll method.
type ListCategoriesFilter struct {
	UserID string
}

// GetCategoryFilter represents a filters for Get method.
type GetCategoryFilter struct {
	ID     string
	UserID string
	Title  string
}

// StateStore represents a store for user states.
type StateStore interface {
	// Create creates a new state in store.
	Create(ctx context.Context, state *models.State) error
	// Get returns a state from store by user id.
	Get(ctx context.Context, filter GetStateFilter) (*models.State, error)
	// Update updates state model in store.
	Update(ctx context.Context, state *models.State) (*models.State, error)
	// Delete deletes state from store.
	Delete(ctx context.Context, ID string) error
}

// GetStateFilter represents a filters for StateStore.Get method.
type GetStateFilter struct {
	UserID string
}

// CurrencyStore represents a store for currencies.
type CurrencyStore interface {
	// Create creates a new currency in store(only in case if currency not exists).
	// The check for existence is based on currency code(models.Currency.Code).
	CreateIfNotExists(ctx context.Context, currency *models.Currency) error
	// Count returns a count of all currencies from store.
	Count(ctx context.Context) (int, error)
	// List returns a list of all currencies from store.
	List(ctx context.Context, filter ListCurrenciesFilter) ([]models.Currency, error)
	// Exists checks if currency exists in store based on input filter.
	Exists(ctx context.Context, opts ExistsCurrencyFilter) (bool, error)
}

// ListCurrenciesFilter represents a filter for CurrencyStore.List method.
type ListCurrenciesFilter struct {
	Pagination *Pagination
}

// ExistsCurrencyFilter represents a filter for CurrencyStore.Exists method.
type ExistsCurrencyFilter struct {
	ID string
}

// BalanceSubscriptionStore represents a store for balance subscriptions.
type BalanceSubscriptionStore interface {
	// Create creates a new balance subscription in store.
	Create(ctx context.Context, subscription models.BalanceSubscription) error
	// CreateScheduledOperation creates a new scheduled operation in store.
	CreateScheduledOperation(ctx context.Context, operation models.ScheduledOperation) error
	// Get retrieves balance subscription from store based on input filter.
	Get(ctx context.Context, filter GetBalanceSubscriptionFilter) (*models.BalanceSubscription, error)
	// Count returns a count of all balance subscriptions from store based on filter.
	Count(ctx context.Context, filter ListBalanceSubscriptionFilter) (int, error)
	// List returns a list of all balance subscriptions from store based on filter.
	List(ctx context.Context, filter ListBalanceSubscriptionFilter) ([]models.BalanceSubscription, error)
	// ListScheduledOperation returns a list of all scheduled operation based on input filters.
	ListScheduledOperation(ctx context.Context, filter ListScheduledOperation) ([]models.ScheduledOperation, error)
	// Update updates balance subscription model in store.
	Update(ctx context.Context, subscription *models.BalanceSubscription) error
	// MarkScheduledOperationAsNotified marks a scheduled operation as notified in store.
	MarkScheduledOperationAsNotified(ctx context.Context, scheduledOperationID string) error
	// Delete deletes balance subscription from store.
	Delete(ctx context.Context, subscriptionID string) error
	// DeleteScheduledOperation deletes scheduled operation from store.
	DeleteScheduledOperation(ctx context.Context, id string) error
}

// ListBalanceSubscriptionFilter represents a filter for store.List and store.Count methods.
type ListBalanceSubscriptionFilter struct {
	BalanceID                                                  string
	OrderByCreatedAtDesc                                       bool
	CreatedAtLessThan                                          time.Time
	Limit                                                      int
	SubscriptionsWithLastScheduledOperation                    bool
	SubscriptionsForUserWhoHasEnabledSubscriptionNotifications bool
}

// GetBalanceSubscriptionFilter represents a filter for store.Get method.
type GetBalanceSubscriptionFilter struct {
	ID   string
	Name string
}

// ListScheduledOperation represents a filter for store.ListScheduledOperation method.
type ListScheduledOperation struct {
	BetweenFilter          *BetweenFilter
	BalanceSubscriptionIDs []string
	NotNotified            bool
}

// BetweenFilter represents a time range filter with inclusive From and To boundaries
// for filtering data between two points in time.
type BetweenFilter struct {
	From time.Time
	To   time.Time
}

// Pagination represents a pagination filter for store.List methods.
type Pagination struct {
	Page  int
	Limit int
}
