package store

import (
	"context"
	"errors"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type stateStore struct {
	*database.MongoDB
}

var _ service.StateStore = (*stateStore)(nil)

var collectionState = "States"

// NewState returns new instance of state store.
func NewState(db *database.MongoDB) *stateStore {
	return &stateStore{
		db,
	}
}

func (s stateStore) Create(ctx context.Context, state *models.State) error {
	_, err := s.DB.Collection(collectionState).InsertOne(ctx, state)
	if err != nil {
		return err
	}

	return nil
}

func (s stateStore) Get(ctx context.Context, filter service.GetStateFilter) (*models.State, error) {
	stmt := bson.M{}

	if filter.UserID != "" {
		stmt["userId"] = filter.UserID
	}

	var state models.State
	err := s.DB.Collection(collectionState).FindOne(ctx, stmt).Decode(&state)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &state, nil
}

func (s stateStore) Update(ctx context.Context, state *models.State) (*models.State, error) {
	filter := bson.M{"_id": state.ID}
	update := bson.M{"$set": state}

	result := s.DB.Collection(collectionState).FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, result.Err()
	}

	var updatedState models.State
	err := result.Decode(&updatedState)
	if err != nil {
		return nil, err
	}

	return &updatedState, nil
}

func (s stateStore) Delete(ctx context.Context, ID string) error {
	_, err := s.DB.Collection(collectionState).DeleteOne(ctx, bson.M{"_id": ID})
	if err != nil {
		return err
	}

	return nil
}
