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

func TestUser_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	userStore := store.NewUser(db)

	userID := uuid.NewString()

	testCases := []struct {
		desc                 string
		preconditions        *models.User
		input                *models.User
		expectDuplicateError bool
	}{
		{
			desc: "positive: user created",
			input: &models.User{
				ID:       uuid.NewString(),
				Username: "test",
			},
		},
		{
			desc: "negative: user not created because already exist",
			preconditions: &models.User{
				ID: userID,
			},
			input: &models.User{
				ID: userID,
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = userStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				_, err := userStore.DB.Collection("Users").DeleteOne(ctx, bson.M{"_id": tc.input.ID})
				assert.NoError(t, err)
			})

			err := userStore.Create(ctx, tc.input)
			if tc.expectDuplicateError {
				assert.True(t, mongo.IsDuplicateKeyError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	userStore := store.NewUser(db)
	balanceStore := store.NewBalance(db)

	userID, userID2, balanceID := uuid.NewString(), uuid.NewString(), uuid.NewString()

	testCases := []struct {
		desc          string
		preconditions *models.User
		input         service.GetUserFilter
		expected      *models.User
	}{
		{
			desc: "positive: user by username found",
			preconditions: &models.User{
				ID:       userID,
				Username: "test",
			},
			input: service.GetUserFilter{
				Username: "test",
			},
			expected: &models.User{
				ID:       userID,
				Username: "test",
			},
		},
		{
			desc: "positive: user with balance preload by username found",
			preconditions: &models.User{
				ID:       userID2,
				Username: "test2",
				Balances: []models.Balance{
					{
						ID:       balanceID,
						UserID:   userID2,
						Amount:   "10",
						Currency: "$",
					},
				},
			},
			input: service.GetUserFilter{
				Username:        "test2",
				PreloadBalances: true,
			},
			expected: &models.User{
				ID:       userID2,
				Username: "test2",
				Balances: []models.Balance{
					{
						ID:       balanceID,
						UserID:   userID2,
						Amount:   "10",
						Currency: "$",
					},
				},
			},
		},
		{
			desc: "negative: user not found",
			input: service.GetUserFilter{
				Username: "not_found_user_test",
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := userStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)

				if len(tc.preconditions.Balances) != 0 {
					for _, balance := range tc.preconditions.Balances {
						balance.UserID = tc.preconditions.ID
						err = balanceStore.Create(ctx, &balance)
						require.NoError(t, err)
					}
				}
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					_, err := userStore.DB.Collection("Users").DeleteOne(ctx, bson.M{"_id": tc.preconditions.ID})
					assert.NoError(t, err)

					if len(tc.preconditions.Balances) != 0 {
						for _, balance := range tc.preconditions.Balances {
							_, err := balanceStore.DB.Collection("Balance").DeleteOne(ctx, bson.M{"_id": balance.ID})
							assert.NoError(t, err)
						}
					}
				}
			})

			actual, err := userStore.Get(ctx, tc.input)
			assert.NoError(t, err)

			if tc.preconditions != nil {
				assert.Equal(t, tc.expected.ID, actual.ID)
				assert.Equal(t, tc.expected.Username, actual.Username)

				// NOTE: We don't care about balances order, since in all test cases we have only one balance.
				for i := 0; i < len(tc.expected.Balances); i++ {
					assert.Equal(t, tc.expected.Balances[i].ID, actual.Balances[i].ID)
					assert.Equal(t, tc.expected.Balances[i].UserID, actual.Balances[i].UserID)
					assert.Equal(t, tc.expected.Balances[i].Amount, actual.Balances[i].Amount)
					assert.Equal(t, tc.expected.Balances[i].Currency, actual.Balances[i].Currency)
				}
			} else {
				assert.Nil(t, actual)
			}
		})
	}
}
