package store

import (
	"context"
	"errors"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type categoryStore struct {
	*database.MongoDB
}

var _ service.CategoryStore = (*categoryStore)(nil)

var collectionCategory = "Categories"

// NewCategory returns a new instance of category store.
func NewCategory(db *database.MongoDB) *categoryStore {
	return &categoryStore{
		db,
	}
}

func (c categoryStore) List(ctx context.Context, filter *service.ListCategoriesFilter) ([]models.Category, error) {
	stmt := bson.M{}

	if filter.UserID != "" {
		stmt = bson.M{"userId": filter.UserID}
	}

	cursor, err := c.DB.Collection(collectionCategory).Find(ctx, stmt)
	if err != nil {
		return nil, err
	}

	var categories []models.Category
	err = cursor.All(ctx, &categories)
	if err != nil {
		return nil, err
	}

	err = cursor.Close(ctx)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (c categoryStore) Get(ctx context.Context, filter service.GetCategoryFilter) (*models.Category, error) {
	request := bson.M{}

	if filter.ID != "" {
		request["_id"] = filter.ID
	}
	if filter.Title != "" {
		request["title"] = filter.Title
	}

	var category models.Category
	err := c.DB.Collection(collectionCategory).FindOne(ctx, request).Decode(&category)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &category, nil
}

func (c categoryStore) Create(ctx context.Context, category *models.Category) error {
	_, err := c.DB.Collection(collectionCategory).InsertOne(ctx, category)
	if err != nil {
		return err
	}

	return nil
}

func (c categoryStore) Delete(ctx context.Context, categoryID string) error {
	_, err := c.DB.Collection(collectionCategory).DeleteOne(ctx, bson.M{"_id": categoryID})
	if err != nil {
		return err
	}

	return nil
}
