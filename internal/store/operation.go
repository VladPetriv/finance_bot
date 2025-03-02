package store

import (
	"context"
	"errors"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type operationStore struct {
	*database.MongoDB
}

var _ service.OperationStore = (*operationStore)(nil)

var collectionOperation = "Operations"

// NewOperation returns new instance of operation store.
func NewOperation(db *database.MongoDB) *operationStore {
	return &operationStore{
		db,
	}
}

func (o operationStore) List(ctx context.Context, filter service.ListOperationsFilter) ([]models.Operation, error) {
	stmt, findOptions := applyListOperationsFilter(filter)

	cursor, err := o.DB.Collection(collectionOperation).Find(ctx, stmt, findOptions)
	if err != nil {
		return nil, err
	}

	var operations []models.Operation
	err = cursor.All(ctx, &operations)
	if err != nil {
		return nil, err
	}

	err = cursor.Close(ctx)
	if err != nil {
		return nil, err
	}

	return operations, nil
}

func (o operationStore) Count(ctx context.Context, filter service.ListOperationsFilter) (int, error) {
	stmt, _ := applyListOperationsFilter(filter)

	count, err := o.DB.Collection(collectionOperation).CountDocuments(ctx, stmt)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func applyListOperationsFilter(filter service.ListOperationsFilter) (*bson.M, *options.FindOptions) {
	stmt := bson.M{}

	if filter.BalanceID != "" {
		stmt["balanceId"] = filter.BalanceID
	}

	if filter.CreationPeriod != "" {
		startDate, endDate := filter.CreationPeriod.CalculateTimeRange()
		stmt["createdAt"] = bson.M{
			"$gte": startDate,
			"$lte": endDate,
		}
	}

	if filter.Month != "" {
		startDate, endDate := filter.Month.GetTimeRange(time.Now())
		stmt["createdAt"] = bson.M{
			"$gte": startDate,
			"$lte": endDate,
		}
	}

	if !filter.CreatedAtLessThan.IsZero() {
		stmt["createdAt"] = bson.M{
			"$lt": filter.CreatedAtLessThan,
		}
	}

	findOptions := &options.FindOptions{}
	if filter.Limit != 0 {
		findOptions = findOptions.SetLimit(int64(filter.Limit))
	}

	if filter.SortByCreatedAtDesc {
		findOptions = findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	}

	return &stmt, findOptions
}

func (o operationStore) Get(ctx context.Context, filter service.GetOperationFilter) (*models.Operation, error) {
	stmt := bson.M{}

	if filter.ID != "" {
		stmt["_id"] = filter.ID
	}
	if filter.Type != "" {
		stmt["type"] = filter.Type
	}
	if filter.Amount != "" {
		stmt["amount"] = filter.Amount
	}
	if len(filter.BalanceIDs) != 0 {
		stmt["balanceId"] = bson.M{"$in": filter.BalanceIDs}
	}
	if !filter.CreateAtFrom.IsZero() {
		stmt["createdAt"] = bson.M{
			"$gte": filter.CreateAtFrom,
		}
	}
	if !filter.CreateAtTo.IsZero() {
		stmt["createdAt"] = bson.M{
			"$lt": filter.CreateAtTo,
		}
	}

	var operation models.Operation
	err := o.DB.Collection(collectionOperation).FindOne(ctx, stmt).Decode(&operation)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &operation, nil
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
