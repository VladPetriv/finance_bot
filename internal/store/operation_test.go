package store_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
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

func TestOperation_GetAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)

	operationStore := store.NewOperation(db)

	type input struct {
		balanceID string
		filter    service.GetAllOperationsFilter
	}

	testCases := []struct {
		desc          string
		preconditions []models.Operation
		input         input
		expected      []models.Operation
	}{
		{
			desc: "positive: received all operations by only balance id",
			preconditions: []models.Operation{
				{ID: "1", BalanceID: "id"}, {ID: "2", BalanceID: "id"},
			},
			input: input{
				balanceID: "id",
			},
			expected: []models.Operation{
				{ID: "1", BalanceID: "id"}, {ID: "2", BalanceID: "id"},
			},
		},
		{
			desc: "positive: received all operations by day as a creation period",
			preconditions: []models.Operation{
				{ID: "1.1", BalanceID: "id1", CreatedAt: time.Now().Add(-23 * time.Hour)},
				{ID: "2.1", BalanceID: "id1", CreatedAt: time.Now()},
				{ID: "3.1", BalanceID: "id1", CreatedAt: time.Now().Add(-48 * time.Hour)},
			},
			input: input{
				balanceID: "id1",
				filter: service.GetAllOperationsFilter{
					CreationPeriod: &models.CreationPeriodDay,
				},
			},
			expected: []models.Operation{
				{ID: "1.1", BalanceID: "id1", CreatedAt: time.Now().Add(-23 * time.Hour)},
				{ID: "2.1", BalanceID: "id1", CreatedAt: time.Now()},
			},
		},
		{
			desc: "positive: received all operations by week as a creation period",
			preconditions: []models.Operation{
				{ID: "1.2", BalanceID: "id2", CreatedAt: time.Now().Add(-168 * time.Hour)},
				{ID: "2.2", BalanceID: "id2", CreatedAt: time.Now().Add(-100 * time.Hour)},
				{ID: "3.2", BalanceID: "id2", CreatedAt: time.Now().Add(-48 * time.Hour)},
			},
			input: input{
				balanceID: "id2",
				filter: service.GetAllOperationsFilter{
					CreationPeriod: &models.CreationPeriodWeek,
				},
			},
			expected: []models.Operation{
				{ID: "2.2", BalanceID: "id2", CreatedAt: time.Now().Add(-100 * time.Hour)},
				{ID: "3.2", BalanceID: "id2", CreatedAt: time.Now().Add(-48 * time.Hour)},
			},
		},
		{
			desc: "positive: received all operations by month as a creation period",
			preconditions: []models.Operation{
				{ID: "1.3", BalanceID: "id3", CreatedAt: time.Now().Add(-730 * time.Hour)},
				{ID: "2.3", BalanceID: "id3", CreatedAt: time.Now().Add(-200 * time.Hour)},
				{ID: "3.3", BalanceID: "id3", CreatedAt: time.Now().Add(-300 * time.Hour)},
				{ID: "4.3", BalanceID: "id3", CreatedAt: time.Now()},
			},
			input: input{
				balanceID: "id3",
				filter: service.GetAllOperationsFilter{
					CreationPeriod: &models.CreationPeriodMonth,
				},
			},
			expected: []models.Operation{
				{ID: "2.3", BalanceID: "id3", CreatedAt: time.Now().Add(-200 * time.Hour)},
				{ID: "3.3", BalanceID: "id3", CreatedAt: time.Now().Add(-300 * time.Hour)},
			},
		},
		{
			desc: "positive: received all operations by year as a creation period",
			preconditions: []models.Operation{
				{ID: "1.4", BalanceID: "id4", CreatedAt: time.Now().Add(-8760 * time.Hour)},
				{ID: "2.4", BalanceID: "id4", CreatedAt: time.Now().Add(-3500 * time.Hour)},
				{ID: "3.4", BalanceID: "id4", CreatedAt: time.Now().Add(-1000 * time.Hour)},
				{ID: "4.4", BalanceID: "id4", CreatedAt: time.Now()},
			},
			input: input{
				balanceID: "id4",
				filter: service.GetAllOperationsFilter{
					CreationPeriod: &models.CreationPeriodYear,
				},
			},
			expected: []models.Operation{
				{ID: "2.4", BalanceID: "id4", CreatedAt: time.Now().Add(-3500 * time.Hour)},
				{ID: "3.4", BalanceID: "id4", CreatedAt: time.Now().Add(-1000 * time.Hour)},
			},
		},
		{
			desc: "negative: operations not found",
			input: input{
				balanceID: "not_found",
				filter: service.GetAllOperationsFilter{
					CreationPeriod: &models.CreationPeriodYear,
				},
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				for _, o := range tc.preconditions {
					err := operationStore.Create(ctx, &o)
					require.NoError(t, err)
				}
			}

			t.Cleanup(func() {
				_, err := operationStore.DB.Collection("Operation").DeleteMany(ctx, bson.M{"balanceId": tc.input.balanceID})
				assert.NoError(t, err)
			})

			actual, err := operationStore.GetAll(ctx, tc.input.balanceID, tc.input.filter)
			assert.NoError(t, err)

			// NOTE: Sort both slices to get right order when compare.
			sort.Slice(actual, func(i, j int) bool {
				return actual[i].ID < actual[j].ID
			})
			sort.Slice(tc.expected, func(i, j int) bool {
				return tc.expected[i].ID < tc.expected[j].ID
			})

			for i := 0; i < len(tc.expected); i++ {
				assert.Equal(t, tc.expected[i].ID, actual[i].ID)
				assert.Equal(t, tc.expected[i].BalanceID, actual[i].BalanceID)
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
