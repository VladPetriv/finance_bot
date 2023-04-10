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

// NewOperationStore returns new instance of operations store.
func NewOperationStore(db *database.MongoDB) *operationStore {
	return &operationStore{
		db,
	}
}

func (o operationStore) GetAll(ctx context.Context) ([]models.Operation, error) {
	cursor, err := o.DB.Collection(collectionOperation).Find(ctx, &bson.M{})
	if err != nil {
		return nil, err
	}

	var operations []models.Operation

	if err := cursor.Decode(&operations); err != nil {
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

func (o operationStore) Delete(ctx context.Context, operationID string) error {
	_, err := o.DB.Collection(collectionOperation).DeleteOne(ctx, bson.M{"_id": operationID})
	if err != nil {
		return err
	}

	return nil
}
