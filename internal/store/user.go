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

type userStore struct {
	*database.MongoDB
}

var _ service.UserStore = (*userStore)(nil)

var collectionUser = "Users"

// NewUser returns new instance of user store.
func NewUser(db *database.MongoDB) *userStore {
	return &userStore{
		db,
	}
}

func (u userStore) Create(ctx context.Context, user *models.User) error {
	_, err := u.DB.Collection(collectionUser).InsertOne(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (u userStore) Get(ctx context.Context, filter service.GetUserFilter) (*models.User, error) {
	stmt := bson.M{}

	if filter.Username != "" {
		stmt["username"] = filter.Username
	}

	pipeLine := mongo.Pipeline{
		bson.D{{Key: "$match", Value: stmt}},
	}

	if filter.PreloadBalances {
		pipeLine = append(pipeLine, bson.D{
			{
				Key: "$lookup",
				Value: bson.M{
					"from":         collectionBalance,
					"localField":   "_id",
					"foreignField": "userId",
					"as":           "balances",
				},
			},
		})
	}

	cursor, err := u.DB.Collection(collectionUser).Aggregate(ctx, pipeLine)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cursor.Close(ctx)
		if err != nil {
			fmt.Printf("error closing cursor: %v", err)
		}
	}()

	var matches []models.User
	err = cursor.All(ctx, &matches)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil
	}

	return &matches[0], nil
}
