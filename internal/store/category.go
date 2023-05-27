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

var collectionCategory = "Category"

// NewCategory returns a new instance of the category store.
func NewCategory(db *database.MongoDB) *categoryStore {
	return &categoryStore{
		db,
	}
}

func (c categoryStore) GetAll(ctx context.Context) ([]models.Category, error) {
	cursor, err := c.DB.Collection(collectionCategory).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var categories []models.Category

	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	if err := cursor.Close(ctx); err != nil {
		return nil, err
	}

	return categories, nil
}

func (c categoryStore) GetByTitle(ctx context.Context, title string) (*models.Category, error) {
	var category models.Category

	err := c.DB.Collection(collectionCategory).FindOne(ctx, bson.M{"title": title}).Decode(&category)
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
