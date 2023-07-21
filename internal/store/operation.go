package store

import (
	"context"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
)

type operationStore struct {
	*database.MongoDB
}

var _ service.OperationStore = (*operationStore)(nil)

var collectionOperation = "Operation"

// NewOperation returns new instance of operation store.
func NewOperation(db *database.MongoDB) *operationStore {
	return &operationStore{
		db,
	}
}

func (o operationStore) GetAll(ctx context.Context, balanceID string) ([]models.Operation, error) {
	cursor, err := o.DB.Collection(collectionOperation).Find(ctx, &bson.M{"balanceId": balanceID})
	if err != nil {
		return nil, err
	}

	var operations []models.Operation

	if err := cursor.All(ctx, &operations); err != nil {
		return nil, err
	}

	if err := cursor.Close(ctx); err != nil {
		return nil, err
	}

	return operations, nil
}

func (o operationStore) Create(ctx context.Context, operation *models.Operation) error {
	_, err := o.DB.Collection(collectionOperation).InsertOne(ctx, operation)
	if err != nil {
		return err
	}

	return nil
}

func (o operationStore) Update(ctx context.Context, operationID string, operation *models.Operation) error {
	_, err := o.DB.Collection(collectionOperation).UpdateByID(ctx, operationID, bson.M{"$set": operation})
	if err != nil {
		return err
	}

	return nil
}

func (o operationStore) Delete(ctx context.Context, operationID string) error {
	_, err := o.DB.Collection(collectionOperation).DeleteOne(ctx, bson.M{"_id": operationID})
	if err != nil {
		return err
	}

	return nil
}
