package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/service/mocks"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCategory_CreateCategory(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	categoryID := uuid.NewString()

	testCases := []struct {
		name     string
		mock     func(categoryStore *mocks.CategoryStore)
		args     *models.Category
		expected error
	}{
		{
			name: "positive: category created",
			mock: func(categoryStore *mocks.CategoryStore) {
				categoryStore.On("GetByTitle", ctx, "test").Return(nil, nil)
				categoryStore.On("Create", ctx, &models.Category{
					ID:    categoryID,
					Title: "test",
				}).Return(nil)
			},
			args: &models.Category{
				ID:    categoryID,
				Title: "test",
			},
			expected: nil,
		},
		{
			name: "negative: category already exists",
			mock: func(categoryStore *mocks.CategoryStore) {
				categoryStore.On("GetByTitle", ctx, "test").
					Return(&models.Category{ID: uuid.NewString(), Title: "test"}, nil)
			},
			args: &models.Category{
				ID:    uuid.NewString(),
				Title: "test",
			},
			expected: service.ErrCategoryAlreadyExists,
		},
		{
			name: "negative: got an error while get category by title",
			mock: func(categoryStore *mocks.CategoryStore) {
				categoryStore.On("GetByTitle", ctx, "test").
					Return(nil, fmt.Errorf("some error"))
			},
			args: &models.Category{
				ID:    uuid.NewString(),
				Title: "test",
			},
			expected: fmt.Errorf("get category by title: %w", fmt.Errorf("some error")),
		},
		{
			name: "negative: got an error while create category",
			mock: func(categoryStore *mocks.CategoryStore) {
				categoryStore.On("GetByTitle", ctx, "test").Return(nil, nil)
				categoryStore.On("Create", ctx, &models.Category{
					ID:    categoryID,
					Title: "test",
				}).Return(fmt.Errorf("some error"))
			},
			args: &models.Category{
				ID:    categoryID,
				Title: "test",
			},
			expected: fmt.Errorf("create category: %w", fmt.Errorf("some error")),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			categoryStoreMock := mocks.NewCategoryStore(t)
			tc.mock(categoryStoreMock)

			categoryService := service.NewCategory(logger.New("debug", ""), categoryStoreMock)

			got := categoryService.CreateCategory(ctx, tc.args)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestCategory_ListCategories(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	categories := []models.Category{
		{ID: uuid.NewString(), UserID: "1"},
		{ID: uuid.NewString(), UserID: "1"},
		{ID: uuid.NewString(), UserID: "2"},
		{ID: uuid.NewString(), UserID: "3"},
	}
	userID := "1"

	type expected struct {
		categories []models.Category
		err        error
	}

	testCases := []struct {
		name     string
		mock     func(categoryStore *mocks.CategoryStore)
		expected expected
	}{
		{
			name: "positive: categories found",
			mock: func(categoryStore *mocks.CategoryStore) {
				categoryStore.On("GetAll", ctx, &service.GetALlCategoriesFilter{
					UserID: &userID,
				}).Return(categories[:1], nil)
			},
			expected: expected{
				categories: categories[:1],
			},
		},
		{
			name: "negative: got an error while get all categories",
			mock: func(categoryStore *mocks.CategoryStore) {
				categoryStore.On("GetAll", ctx, &service.GetALlCategoriesFilter{
					UserID: &userID,
				}).Return(nil, fmt.Errorf("some error"))
			},
			expected: expected{
				err: fmt.Errorf("get all categories: %w", fmt.Errorf("some error")),
			},
		},
		{
			name: "negative: categories not found",
			mock: func(categoryStore *mocks.CategoryStore) {
				categoryStore.On("GetAll", ctx, &service.GetALlCategoriesFilter{
					UserID: &userID,
				}).Return(nil, nil)
			},
			expected: expected{
				err: service.ErrCategoriesNotFound,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			categoryStoreMock := mocks.NewCategoryStore(t)
			tc.mock(categoryStoreMock)

			categoryService := service.NewCategory(logger.New("debug", ""), categoryStoreMock)

			actual, err := categoryService.ListCategories(ctx, userID)
			assert.Equal(t, tc.expected.err, err)
			assert.Equal(t, tc.expected.categories, actual)
		})
	}
}
