package store

import (
	"context"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"code": currency.Code}
	update := bson.M{
		"$setOnInsert": currency,
	}

	_, err := c.DB.Collection(collectionCurrency).UpdateOne(ctx, filter, update, opts)
	return err
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
