package service

import (
	"context"

	"github.com/VladPetriv/finance_bot/internal/models"
)

// Stores represents all Storages.
type Stores struct {
	Balance   BalanceStore
	Operation OperationStore
	Category  CategoryStore
}

// UserStore provides functionality for work with users.
//
//go:generate mockery --dir . --name UserStore --output ./mocks
type UserStore interface {
	// Create creates a new user model in store.
	Create(ctx context.Context, user *models.User) error
	// GetByUsername returns a user from store by username.
	GetByUsername(ctx context.Context, username string) (*models.User, error)
}

// BalanceStore provides functionality for work with balance.
//
//go:generate mockery --dir . --name BalanceStore --output ./mocks
type BalanceStore interface {
	// Create creates a new balance model in store.
	Create(ctx context.Context, balance *models.Balance) error
	// Get returns a balance from store by id.
	Get(ctx context.Context, balanceID string) (*models.Balance, error)
	// Update updates a current balance model in store.
	Update(ctx context.Context, balance *models.Balance) error
	// Delete deletes a balance from store by id.
	Delete(ctx context.Context, balanceID string) error
}

// OperationStore provides functionality for work with operation.
//
//go:generate mockery --dir . --name OperationStore --output ./mocks
type OperationStore interface {
	// GetAll returns all operations from store.
	GetAll(ctx context.Context) ([]models.Operation, error)
	// Create creates a new operation.
	Create(ctx context.Context, operation *models.Operation) error
	// Delete delete operation by his id.
	Delete(ctx context.Context, operationID string) error
}

// CategoryStore provides functionality for work with categories.
//
//go:generate mockery --dir . --name CategoryStore --output ./mocks
type CategoryStore interface {
	// GetAll returns a list of all categories from store.
	GetAll(ctx context.Context) ([]models.Category, error)
	// GetByTitle returns a category by their title.
	GetByTitle(ctx context.Context, title string) (*models.Category, error)
	// Create creates new category model in store.
	Create(ctx context.Context, category *models.Category) error
	// Delete delete category from store by id.
	Delete(ctx context.Context, categoryID string) error
}
