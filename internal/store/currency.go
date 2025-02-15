package store

import (
	"context"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
)

type currencyStore struct {
	*database.MongoDB
}

var collectionCurrency = "Currencies"

// NewCurrency creates a new currency store.
func NewCurrency(db *database.MongoDB) *currencyStore {
	return &currencyStore{
		db,
	}
}

func (c *currencyStore) Create(ctx context.Context, currency *models.Currency) error {
	_, err := c.DB.Collection(collectionCurrency).InsertOne(ctx, currency)
	if err != nil {
		return err
	}

	return nil
}

func (c *currencyStore) Count(ctx context.Context) (int, error) {
	count, err := c.DB.Collection(collectionCurrency).CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (c *currencyStore) List(ctx context.Context) ([]models.Currency, error) {
	cursor, err := c.DB.Collection(collectionCurrency).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var currencies []models.Currency
	err = cursor.All(ctx, &currencies)
	if err != nil {
		return nil, err
	}

	err = cursor.Close(ctx)
	if err != nil {
		return nil, err
	}

	return currencies, nil
}
