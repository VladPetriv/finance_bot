package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCategory_GetAll(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	db, err := database.NewMongoDB(ctx, "mongodb://localhost:27017", "api")
	require.NoError(t, err)

	categoryStore := store.NewCategory(db)

	categoryID1 := uuid.NewString()
	categoryID2 := uuid.NewString()
	categoryID3 := uuid.NewString()

	tests := []struct {
		desc                  string
		categoriesForCreation []models.Category
		want                  []models.Category
	}{
		{
			desc: "should return 3 categories",
			categoriesForCreation: []models.Category{
				{ID: categoryID1}, {ID: categoryID2}, {ID: categoryID3},
			},
			want: []models.Category{
				{ID: categoryID1}, {ID: categoryID2}, {ID: categoryID3},
			},
		},
		{
			desc: "should return nil because there are categories found",
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			for _, c := range tt.categoriesForCreation {
				err := categoryStore.Create(ctx, &c)
				assert.NoError(t, err)
			}

			got, err := categoryStore.GetAll(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

			t.Cleanup(func() {
				for _, c := range tt.categoriesForCreation {
					err := categoryStore.Delete(ctx, c.ID)
					assert.NoError(t, err)
				}
			})
		})
	}
}

func TestCategory_Create(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	db, err := database.NewMongoDB(ctx, "mongodb://localhost:27017", "api")
	require.NoError(t, err)

	categoryStore := store.NewCategory(db)

	categoryID := uuid.NewString()

	tests := []struct {
		desc  string
		input *models.Category
		want  models.Category
	}{
		{
			desc: "should create category",
			input: &models.Category{
				ID:    categoryID,
				Title: "Test",
			},
			want: models.Category{
				ID:    categoryID,
				Title: "Test",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			err := categoryStore.Create(ctx, tt.input)
			assert.NoError(t, err)

			var got models.Category
			result := categoryStore.DB.Collection("Category").FindOne(ctx, bson.M{"_id": tt.input.ID})

			err = result.Decode(&got)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, &got)

			t.Cleanup(func() {
				err := categoryStore.Delete(ctx, tt.input.ID)
				assert.NoError(t, err)
			})
		})
	}
}

func TestCategory_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	db, err := database.NewMongoDB(ctx, "mongodb://localhost:27017", "api")
	require.NoError(t, err)

	categoryStore := store.NewCategory(db)

	categoryID := uuid.NewString()

	tests := []struct {
		desc                string
		categoryForCreation *models.Category
		input               string
	}{
		{
			desc: "should delete category",
			categoryForCreation: &models.Category{
				ID:    categoryID,
				Title: "Test",
			},
			input: categoryID,
		},
		{
			desc:  "should not delete category",
			input: uuid.NewString(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			if tt.categoryForCreation != nil {
				err := categoryStore.Create(ctx, tt.categoryForCreation)
				assert.NoError(t, err)
			}

			err := categoryStore.Delete(ctx, tt.input)
			assert.NoError(t, err)

			var got models.Category
			result := categoryStore.DB.Collection("Category").FindOne(ctx, bson.M{"_id": tt.input})

			err = result.Decode(&got)
			assert.Equal(t, mongo.ErrNoDocuments, err)
			assert.Empty(t, got)
		})
	}
}
