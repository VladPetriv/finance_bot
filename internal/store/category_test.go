package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCategory_Create(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	categoryStore := store.NewCategory(db)

	categoryID := uuid.NewString()

	testCases := []struct {
		desc                 string
		preconditions        *models.Category
		input                *models.Category
		expectDuplicateError bool
	}{
		{
			desc: "positive: created category",
			input: &models.Category{
				ID:    uuid.NewString(),
				Title: "test",
			},
		},
		{
			desc: "negative: category not created because already exists",
			preconditions: &models.Category{
				ID: categoryID,
			},
			input: &models.Category{
				ID: categoryID,
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = categoryStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err = categoryStore.Delete(ctx, tc.input.ID)
				assert.NoError(t, err)
			})

			err := categoryStore.Create(ctx, tc.input)
			if tc.expectDuplicateError {
				assert.True(t, mongo.IsDuplicateKeyError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCategory_GetByTitle(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	categoryStore := store.NewCategory(db)

	categoryID := uuid.NewString()

	testCases := []struct {
		desc          string
		preconditions *models.Category
		input         string
		expected      *models.Category
	}{
		{
			desc: "positive: returned category by Title",
			preconditions: &models.Category{
				ID:    categoryID,
				Title: "test_get",
			},
			input: "test_get",
			expected: &models.Category{
				ID:    categoryID,
				Title: "test_get",
			},
		},
		{
			desc:     "negative: category not found",
			input:    "not_found",
			expected: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = categoryStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err = categoryStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			got, err := categoryStore.GetByTitle(ctx, tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestCategory_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	categoryStore := store.NewCategory(db)

	categoryID := uuid.NewString()

	tests := []struct {
		desc          string
		preconditions *models.Category
		input         string
	}{
		{
			desc: "positive: category deleted",
			preconditions: &models.Category{
				ID: categoryID,
			},
			input: categoryID,
		},
		{
			desc: "negative: category not deleted because of not existed id",
			preconditions: &models.Category{
				ID: uuid.NewString(),
			},
			input: uuid.NewString(),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := categoryStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err := categoryStore.Delete(ctx, tc.preconditions.ID)
				assert.NoError(t, err)
			})

			err := categoryStore.Delete(ctx, tc.input)
			assert.NoError(t, err)

			// operation should not be deleted
			if tc.preconditions.ID != tc.input {
				var category models.Category

				err := db.DB.Collection("Category").
					FindOne(ctx, bson.M{"_id": tc.preconditions.ID}).
					Decode(&category)

				assert.NoError(t, err)
				assert.NotNil(t, category)
			}
		})
	}
}
