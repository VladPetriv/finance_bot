package store_test

import (
	"context"
	"testing"

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

var (
	amount300 = "300.00"
	amount400 = "400.00"
)

func TestBalance_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	balanceStore := store.NewBalance(db)

	balanceID := uuid.NewString()
	userID := uuid.NewString()

	testCases := []struct {
		desc          string
		preconditions *models.Balance
		input         service.GetBalanceFilter
		expected      *models.Balance
	}{
		{
			desc: "positive: balance received",
			preconditions: &models.Balance{
				ID:     balanceID,
				UserID: userID,
				Amount: amount300,
			},
			input: service.GetBalanceFilter{
				UserID: userID,
			},
			expected: &models.Balance{
				ID:     balanceID,
				UserID: userID,
				Amount: amount300,
			},
		},
		{
			desc: "negative: balance not found",
			input: service.GetBalanceFilter{
				UserID: uuid.NewString(),
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := balanceStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					_ = balanceStore.Delete(ctx, tc.preconditions.ID)
				}
			})

			got, err := balanceStore.Get(ctx, tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestBalance_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	balanceStore := store.NewBalance(db)

	balanceID := uuid.NewString()

	testCases := []struct {
		desc                 string
		preconditions        *models.Balance
		input                *models.Balance
		expectDuplicateError bool
	}{
		{
			desc: "positive: balance craeted",
			input: &models.Balance{
				ID:     uuid.NewString(),
				Amount: amount300,
			},
		},
		{
			desc: "positive: balance craeted",
			preconditions: &models.Balance{
				ID: balanceID,
			},
			input: &models.Balance{
				ID: balanceID,
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = balanceStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err = balanceStore.Delete(ctx, tc.input.ID)
				assert.NoError(t, err)
			})

			err := balanceStore.Create(ctx, tc.input)
			if tc.expectDuplicateError {
				assert.True(t, mongo.IsDuplicateKeyError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBalance_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	balanceStore := store.NewBalance(db)

	balanceID1 := uuid.NewString()
	balanceID2 := uuid.NewString()

	userID1 := uuid.NewString()
	userID2 := uuid.NewString()

	testCases := []struct {
		desc          string
		preconditions *models.Balance
		input         *models.Balance
		expected      *models.Balance
	}{
		{
			desc: "positive: balance updated",
			preconditions: &models.Balance{
				ID:     balanceID1,
				UserID: userID1,
				Amount: amount300,
			},
			input: &models.Balance{
				ID:     balanceID1,
				UserID: userID1,
				Amount: amount400,
			},
			expected: &models.Balance{
				ID:     balanceID1,
				UserID: userID1,
				Amount: amount400,
			},
		},
		{
			desc: "negative: balance not updated because of not existed id",
			preconditions: &models.Balance{
				ID:     balanceID2,
				UserID: userID2,
				Amount: amount300,
			},
			input: &models.Balance{
				ID:     uuid.NewString(),
				UserID: userID2,
				Amount: amount400,
			},
			expected: &models.Balance{
				ID:     balanceID2,
				UserID: userID2,
				Amount: amount300,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = balanceStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err = balanceStore.Delete(ctx, tc.preconditions.ID)
				assert.NoError(t, err)
			})

			err = balanceStore.Update(ctx, tc.input)
			assert.NoError(t, err)

			got, err := balanceStore.Get(ctx, service.GetBalanceFilter{UserID: tc.preconditions.UserID})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestBalance_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	balanceStore := store.NewBalance(db)

	balanceID := uuid.NewString()

	testCases := []struct {
		desc          string
		preconditions *models.Balance
		input         string
	}{
		{
			desc: "positive: balance deleted",
			preconditions: &models.Balance{
				ID: balanceID,
			},
			input: balanceID,
		},
		{
			desc: "negative: balance not deleted because of not existed id",
			preconditions: &models.Balance{
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
				err = balanceStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err = balanceStore.Delete(ctx, tc.preconditions.ID)
				assert.NoError(t, err)
			})

			err := balanceStore.Delete(ctx, tc.input)
			assert.NoError(t, err)

			// balance should not be deleted
			if tc.preconditions.ID != tc.input {
				var balance models.Balance

				err := db.DB.Collection("Balances").
					FindOne(ctx, bson.M{"_id": tc.preconditions.ID}).
					Decode(&balance)

				assert.NoError(t, err)
				assert.NotNil(t, balance)
			}
		})
	}
}
