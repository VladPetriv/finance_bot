package store

import (
	"context"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
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
	stmt := bson.M{}

	if filter.BalanceID != "" {
		stmt["balanceId"] = filter.BalanceID
	}

	if filter.CreationPeriod != nil {
		startDate, endDate := calculateTimeRange(*filter.CreationPeriod)
		stmt["createdAt"] = bson.M{
			"$gte": startDate,
			"$lte": endDate,
		}
	}

	cursor, err := o.DB.Collection(collectionOperation).Find(ctx, stmt)
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

// calculateTimeRange is used to calculate start and end times based on a given period
func calculateTimeRange(period models.CreationPeriod) (startTime, endTime time.Time) {
	now := time.Now()
	endTime = now

	switch period {
	case models.CreationPeriodDay:
		startTime = now.Add(-24 * time.Hour)
	case models.CreationPeriodWeek:
		startTime = now.Add(-7 * 24 * time.Hour)
	case models.CreationPeriodMonth:
		startTime = now.Add(-30 * 24 * time.Hour)
	case models.CreationPeriodYear:
		startTime = now.Add(-365 * 24 * time.Hour)
	}

	return startTime, endTime
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
