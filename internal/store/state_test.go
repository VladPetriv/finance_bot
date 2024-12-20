package store_test

import (
	"context"
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

func TestState_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	stateStore := store.NewState(db)

	stateID := uuid.NewString()
	now := time.Now()

	testCases := []struct {
		desc                 string
		preconditions        *models.State
		input                *models.State
		expectDuplicateError bool
	}{
		{
			desc: "positive: state created",
			input: &models.State{
				ID:        uuid.NewString(),
				UserID:    uuid.NewString(),
				Flow:      models.StartFlow,
				Steps:     []models.FlowStep{models.StartFlowStep, models.CreateInitialBalanceFlowStep},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			desc: "negative: state not created because already exists",
			preconditions: &models.State{
				ID:        stateID,
				UserID:    uuid.NewString(),
				Flow:      models.StartFlow,
				Steps:     []models.FlowStep{models.StartFlowStep},
				CreatedAt: now,
				UpdatedAt: now,
			},
			input: &models.State{
				ID: stateID,
			},
			expectDuplicateError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = stateStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				_, err := stateStore.DB.Collection("States").DeleteOne(ctx, bson.M{"_id": tc.input.ID})
				assert.NoError(t, err)
			})

			err := stateStore.Create(ctx, tc.input)
			if tc.expectDuplicateError {
				assert.True(t, mongo.IsDuplicateKeyError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestState_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	stateStore := store.NewState(db)

	now := time.Now()

	testCases := []struct {
		desc          string
		preconditions *models.State
		input         service.GetStateFilter
		expected      *models.State
	}{
		{
			desc: "positive: state found by user ID",
			preconditions: &models.State{
				ID:        uuid.NewString(),
				UserID:    "user123",
				Flow:      models.StartFlow,
				Steps:     []models.FlowStep{models.StartFlowStep, models.CreateInitialBalanceFlowStep},
				CreatedAt: now,
				UpdatedAt: now,
			},
			input: service.GetStateFilter{
				UserID: "user123",
			},
			expected: &models.State{
				ID:        uuid.NewString(),
				UserID:    "user123",
				Flow:      models.StartFlow,
				Steps:     []models.FlowStep{models.StartFlowStep, models.CreateInitialBalanceFlowStep},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			desc: "negative: state not found",
			input: service.GetStateFilter{
				UserID: "nonexistent",
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = stateStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					_, err := stateStore.DB.Collection("States").DeleteOne(ctx, bson.M{"_id": tc.preconditions.ID})
					assert.NoError(t, err)
				}
			})

			actual, err := stateStore.Get(ctx, tc.input)
			assert.NoError(t, err)

			if tc.expected == nil {
				assert.Nil(t, actual)
			} else {
				assert.NotNil(t, actual)
				assert.Equal(t, tc.expected.UserID, actual.UserID)
				assert.Equal(t, tc.expected.Flow, actual.Flow)
				assert.Equal(t, tc.expected.Steps, actual.Steps)
			}
		})
	}
}

func TestState_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
	stateStore := store.NewState(db)

	now := time.Now()
	stateID := uuid.NewString()

	testCases := []struct {
		desc          string
		preconditions *models.State
		input         *models.State
		expected      *models.State
	}{
		{
			desc: "positive: state updated successfully",
			preconditions: &models.State{
				ID:        stateID,
				UserID:    "user123",
				Flow:      models.StartFlow,
				Steps:     []models.FlowStep{models.StartFlowStep},
				CreatedAt: now,
				UpdatedAt: now,
			},
			input: &models.State{
				ID:        stateID,
				UserID:    "user123",
				Flow:      models.StartFlow,
				Steps:     []models.FlowStep{models.StartFlowStep, models.CreateInitialBalanceFlowStep},
				CreatedAt: now,
				UpdatedAt: now.Add(1 * time.Minute),
			},
			expected: &models.State{
				ID:        stateID,
				UserID:    "user123",
				Flow:      models.StartFlow,
				Steps:     []models.FlowStep{models.StartFlowStep, models.CreateInitialBalanceFlowStep},
				CreatedAt: now,
				UpdatedAt: now.Add(1 * time.Minute),
			},
		},
		{
			desc: "negative: state not found for update",
			input: &models.State{
				ID: uuid.NewString(),
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = stateStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					_, err := stateStore.DB.Collection("States").DeleteOne(ctx, bson.M{"_id": tc.preconditions.ID})
					assert.NoError(t, err)
				}
			})

			actual, err := stateStore.Update(ctx, tc.input)
			assert.NoError(t, err)

			if tc.expected == nil {
				assert.Nil(t, actual)
			} else {
				assert.Equal(t, tc.expected.ID, actual.ID)
				assert.Equal(t, tc.expected.UserID, actual.UserID)
				assert.Equal(t, tc.expected.Flow, actual.Flow)
				assert.Equal(t, tc.expected.Steps, actual.Steps)
				assert.Equal(t, tc.expected.CreatedAt, actual.CreatedAt)
				assert.Equal(t, tc.expected.UpdatedAt, actual.UpdatedAt)
			}
		})
	}
}
