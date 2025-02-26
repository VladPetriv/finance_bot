package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
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

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, "category_create_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.DB.Drop(ctx)
		assert.NoError(t, err)
	})

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

func TestCategory_Get(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, "category_get_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.DB.Drop(ctx)
		assert.NoError(t, err)
	})

	categoryStore := store.NewCategory(db)

	categoryID1, categoryID2, categoryID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()
	userID1 := uuid.NewString()
	title := "test_get"

	testCases := []struct {
		desc          string
		preconditions *models.Category
		input         service.GetCategoryFilter
		expected      *models.Category
	}{
		{
			desc: "positive: returned category by title",
			preconditions: &models.Category{
				ID:    categoryID1,
				Title: title,
			},
			input: service.GetCategoryFilter{
				Title: title,
			},
			expected: &models.Category{
				ID:    categoryID1,
				Title: title,
			},
		},
		{
			desc: "positive: returned category by id",
			preconditions: &models.Category{
				ID:    categoryID2,
				Title: "test_get",
			},
			input: service.GetCategoryFilter{
				ID: categoryID2,
			},
			expected: &models.Category{
				ID:    categoryID2,
				Title: "test_get",
			},
		},
		{
			desc: "positive: returned category by user id",
			preconditions: &models.Category{
				ID:     categoryID3,
				UserID: userID1,
				Title:  "test_get_user_id",
			},
			input: service.GetCategoryFilter{
				UserID: userID1,
			},
			expected: &models.Category{
				ID:     categoryID3,
				UserID: userID1,
				Title:  "test_get_user_id",
			},
		},
		{
			desc: "negative: category not found (by title)",
			input: service.GetCategoryFilter{
				Title: categoryID1,
			},
			expected: nil,
		},
		{
			desc: "negative: category not found (by id)",
			input: service.GetCategoryFilter{
				ID: title,
			},
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

			got, err := categoryStore.Get(ctx, tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestCategory_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, "category_delete_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.DB.Drop(ctx)
		assert.NoError(t, err)
	})

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

				err := db.DB.Collection("Categories").
					FindOne(ctx, bson.M{"_id": tc.preconditions.ID}).
					Decode(&category)

				assert.NoError(t, err)
				assert.NotNil(t, category)
			}
		})
	}
}

func TestCategory_Update(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, "category_update_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.DB.Drop(ctx)
		assert.NoError(t, err)
	})

	categoryStore := store.NewCategory(db)

	categoryID1 := uuid.NewString()
	categoryID2 := uuid.NewString()

	testCases := []struct {
		desc          string
		preconditions *models.Category
		input         *models.Category
		expected      *models.Category
	}{
		{
			desc: "positive: category updated",
			preconditions: &models.Category{
				ID:    categoryID1,
				Title: "old_title",
			},
			input: &models.Category{
				ID:    categoryID1,
				Title: "new_title",
			},
			expected: &models.Category{
				ID:    categoryID1,
				Title: "new_title",
			},
		},
		{
			desc: "negative: category not updated because of not existed id",
			preconditions: &models.Category{
				ID:    categoryID2,
				Title: "test_title",
			},
			input: &models.Category{
				ID:    uuid.NewString(),
				Title: "updated_title",
			},
			expected: &models.Category{
				ID:    categoryID2,
				Title: "test_title",
			},
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
				err = categoryStore.Delete(ctx, tc.preconditions.ID)
				assert.NoError(t, err)
			})

			err = categoryStore.Update(ctx, tc.input)
			assert.NoError(t, err)

			got, err := categoryStore.Get(ctx, service.GetCategoryFilter{ID: tc.preconditions.ID})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
