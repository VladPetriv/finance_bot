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

type balanceStore struct {
	*database.MongoDB
}

var _ service.BalanceStore = (*balanceStore)(nil)

var collectionBalance = "Balance"

// NewBalance returns new instance of balance store.
func NewBalance(db *database.MongoDB) *balanceStore {
	return &balanceStore{
		db,
	}
}

func (b balanceStore) Get(ctx context.Context, balanceID string) (*models.Balance, error) {
	var balance models.Balance

	err := b.DB.Collection(collectionBalance).FindOne(ctx, bson.M{"_id": balanceID}).Decode(&balance)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &balance, nil
}

func (b balanceStore) Create(ctx context.Context, balance *models.Balance) error {
	_, err := b.DB.Collection(collectionBalance).InsertOne(ctx, balance)
	if err != nil {
		return err
	}

	return nil
}

func (b balanceStore) Update(ctx context.Context, balance *models.Balance) error {
	_, err := b.DB.Collection(collectionBalance).UpdateByID(ctx, balance.ID, bson.M{"$set": balance})
	if err != nil {
		return err
	}

	return nil
}

func (b balanceStore) Delete(ctx context.Context, balanceID string) error {
	_, err := b.DB.Collection(collectionBalance).DeleteOne(ctx, bson.M{"_id": balanceID})
	if err != nil {
		return err
	}

	return nil
}
