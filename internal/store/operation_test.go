package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestOperation_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	operationStore := store.NewOperation(db)

	operationID := uuid.NewString()

	testCases := []struct {
		desc                 string
		preconditions        *models.Operation
		input                *models.Operation
		expectDuplicateError bool
	}{
		{
			desc: "positive: operation created",
			input: &models.Operation{
				ID:         uuid.NewString(),
				Type:       models.OperationTypeIncoming,
				CategoryID: uuid.NewString(),
			},
		},
		{
			desc: "negative: operation not created because already exist",
			preconditions: &models.Operation{
				ID: operationID,
			},
			input: &models.Operation{
				ID: operationID,
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = operationStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err = operationStore.Delete(ctx, tc.input.ID)
				assert.NoError(t, err)
			})

			err := operationStore.Create(ctx, tc.input)
			if tc.expectDuplicateError {
				assert.True(t, mongo.IsDuplicateKeyError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOperation_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	operationStore := store.NewOperation(db)

	operationID := uuid.NewString()

	testCases := []struct {
		desc          string
		preconditions *models.Operation
		input         string
	}{
		{
			desc: "positive: operation deleted",
			preconditions: &models.Operation{
				ID: operationID,
			},
			input: operationID,
		},
		{
			desc: "negatie: operation not deleted because of not existed id",
			preconditions: &models.Operation{
				ID: uuid.NewString(),
			},
			input: uuid.NewString(),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := operationStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err := operationStore.Delete(ctx, tc.preconditions.ID)
				assert.NoError(t, err)
			})

			err := operationStore.Delete(ctx, tc.input)
			assert.NoError(t, err)

			// operation should not be deleted
			if tc.preconditions.ID != tc.input {
				var operation models.Operation

				err := db.DB.Collection("Operation").
					FindOne(ctx, bson.M{"_id": tc.preconditions.ID}).
					Decode(&operation)

				assert.NoError(t, err)
				assert.NotNil(t, operation)
			}
		})
	}
}
