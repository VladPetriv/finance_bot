package service

import (
	"context"

	"github.com/VladPetriv/finance_bot/internal/models"
)

// Stores represents all stores.
type Stores struct {
	Balance   BalanceStore
	Operation OperationStore
	Category  CategoryStore
	User      UserStore
	State     StateStore
}

// UserStore provides functionality for work with users store.
//
//go:generate mockery --dir . --name UserStore --output ./mocks
type UserStore interface {
	// Create creates a new user in store.
	Create(ctx context.Context, user *models.User) error
	// GetByUsername returns a user from store by username.
	Get(ctx context.Context, filtera GetUserFilter) (*models.User, error)
}

// GetUserFilter represents a filters for GetUser method.
type GetUserFilter struct {
	Username        string
	PreloadBalances bool
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

type GetBalanceFilter struct {
	UserID    string
	BalanceID string
}

// OperationStore provides functionality for work with operation store.
//
//go:generate mockery --dir . --name OperationStore --output ./mocks
type OperationStore interface {
	// Create creates a new operation.
	Create(ctx context.Context, operation *models.Operation) error
	// GetAll returns all operations from store by balance id.
	GetAll(ctx context.Context, balanceID string, filters GetAllOperationsFilter) ([]models.Operation, error)
	// Update updates an operation in store.
	Update(ctx context.Context, operationID string, operation *models.Operation) error
	// Delete delete operation by his id.
	Delete(ctx context.Context, operationID string) error
}

// GetAllOperationsFilter represents filters for getting all operations from store.
type GetAllOperationsFilter struct {
	CreationPeriod *models.CreationPeriod
}

// CategoryStore provides functionality for work with categories store.
//
//go:generate mockery --dir . --name CategoryStore --output ./mocks
type CategoryStore interface {
	// GetAll returns a list of all categories from store.
	GetAll(ctx context.Context, filters *GetALlCategoriesFilter) ([]models.Category, error)

	// Get returns a category by fileers.
	Get(ctx context.Context, filter GetCategoryFilter) (*models.Category, error)
	// Create creates new category in store.
	Create(ctx context.Context, category *models.Category) error
	// Delete delete category from store.
	Delete(ctx context.Context, categoryID string) error
}

// GetALlCategoriesFilter represents a filters for GetAll method.
type GetALlCategoriesFilter struct {
	UserID *string
}

// GetCategoryFilter represents a filters for Get method.
type GetCategoryFilter struct {
	Title *string
	ID    *string
}

// StateStore represents a store for user states.
type StateStore interface {
	// Create creates a new state in store.
	Create(ctx context.Context, state *models.State) error
	// Get returns a state from store by user id.
	Get(ctx context.Context, filter GetStateFilter) (*models.State, error)
	// Update updates state model in store.
	Update(ctx context.Context, state *models.State) (*models.State, error)
	//Delete deletes state from store.
	Delete(ctx context.Context, ID string) error
}

// GetStateFilter represents a filters for StateStore.Get method.
type GetStateFilter struct {
	UserID string
}
