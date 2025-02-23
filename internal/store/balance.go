package store

import (
	"context"
	"fmt"

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

var collectionBalance = "Balances"

// NewBalance returns new instance of balance store.
func NewBalance(db *database.MongoDB) *balanceStore {
	return &balanceStore{
		db,
	}
}

func (b balanceStore) Get(ctx context.Context, filter service.GetBalanceFilter) (*models.Balance, error) {
	stmt := bson.M{}

	if filter.BalanceID != "" {
		stmt["_id"] = filter.BalanceID
	}
	if filter.Name != "" {
		stmt["name"] = filter.Name
	}
	if filter.UserID != "" {
		stmt["userId"] = filter.UserID
	}

	pipeLine := mongo.Pipeline{
		bson.D{{Key: "$match", Value: stmt}},
	}

	if filter.PreloadCurrency {
		pipeLine = append(pipeLine, bson.D{
			{
				Key: "$lookup",
				Value: bson.M{
					"from":         collectionCurrency,
					"localField":   "currencyId",
					"foreignField": "_id",
					"as":           "currency",
				},
			},
		})
		pipeLine = append(pipeLine, bson.D{
			{
				Key: "$unwind",
				Value: bson.M{
					"path":                       "$currency",
					"preserveNullAndEmptyArrays": true,
				},
			},
		})
	}

	cursor, err := b.DB.Collection(collectionBalance).Aggregate(ctx, pipeLine)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cursor.Close(ctx)
		if err != nil {
			fmt.Printf("error closing cursor: %v", err)
		}
	}()

	var matches []models.Balance
	err = cursor.All(ctx, &matches)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil
	}

	return &matches[0], nil
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
