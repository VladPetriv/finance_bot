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

func TestUser_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	userStore := store.NewUserStore(db)

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
				_, err := userStore.DB.Collection("User").DeleteOne(ctx, bson.M{"_id": tc.input.ID})
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

func TestUser_GetByUsername(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	userStore := store.NewUserStore(db)

	userID := uuid.NewString()

	testCases := []struct {
		desc          string
		preconditions *models.User
		input         string
		expected      *models.User
	}{
		{
			desc: "positive: user found",
			preconditions: &models.User{
				ID:       userID,
				Username: "test",
			},
			input: "test",
			expected: &models.User{
				ID:       userID,
				Username: "test",
			},
		},
		{
			desc:     "negative: user not found",
			input:    "not_found",
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
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					_, err := userStore.DB.Collection("User").DeleteOne(ctx, bson.M{"_id": tc.preconditions.ID})
					assert.NoError(t, err)
				}
			})

			actual, err := userStore.GetByUsername(ctx, tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}

}
