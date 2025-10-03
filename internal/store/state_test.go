package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "state_create")
	userStore := store.NewUser(testCaseDB)
	stateStore := store.NewState(testCaseDB)

	userID1, userID2 := uuid.NewString(), uuid.NewString()
	stateID := uuid.NewString()

	for _, userID := range [...]string{userID1, userID2} {
		err := userStore.Create(ctx, &model.User{
			ID:       userID,
			Username: "test_state_create" + userID,
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, userID := range [...]string{userID1, userID2} {
			err := deleteUserByID(testCaseDB.DB, userID)
			require.NoError(t, err)
		}
	})

	testCases := [...]struct {
		desc                 string
		preconditions        *model.State
		args                 *model.State
		expectDuplicateError bool
	}{
		{
			desc: "state created",
			args: &model.State{
				ID:     uuid.NewString(),
				UserID: "test_state_create" + userID1,
				Flow:   model.StartFlow,
				Steps:  []model.FlowStep{model.StartFlowStep, model.CreateInitialBalanceFlowStep},
				Metedata: map[string]any{
					"string": "test",
					"bool":   true,
				},
			},
		},
		{
			desc: "state not created because already exists",
			preconditions: &model.State{
				ID:     stateID,
				UserID: "test_state_create" + userID2,
			},
			args: &model.State{
				ID:     stateID,
				UserID: "test_state_create" + userID2,
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := stateStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err := stateStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
				if tc.args != nil {
					err := stateStore.Delete(ctx, tc.args.ID)
					assert.NoError(t, err)
				}
			})

			err := stateStore.Create(ctx, tc.args)
			if tc.expectDuplicateError {
				assert.True(t, isDuplicateKeyError(err))
				return
			}

			assert.NoError(t, err)

			actual, err := stateStore.Get(ctx, service.GetStateFilter{UserID: tc.args.UserID})
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.args.ID, actual.ID)
			assert.Equal(t, tc.args.UserID, actual.UserID)
			assert.Equal(t, tc.args.Flow, actual.Flow)
			if len(tc.args.Steps) != 0 {
				assert.Equal(t, tc.args.Steps, actual.Steps)
			}
			if tc.args.Metedata != nil {
				assert.Equal(t, tc.args.Metedata["string"].(string), actual.Metedata["string"].(string))
				assert.Equal(t, tc.args.Metedata["bool"].(bool), actual.Metedata["bool"].(bool))
			}
		})
	}
}

func TestState_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "state_get")
	userStore := store.NewUser(testCaseDB)
	stateStore := store.NewState(testCaseDB)

	userID := uuid.NewString()
	stateID := uuid.NewString()
	err := userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test_state_get" + userID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions *model.State
		args          service.GetStateFilter
		expected      *model.State
	}{
		{
			desc: "state found by user ID",
			preconditions: &model.State{
				ID:     stateID,
				UserID: "test_state_get" + userID,
				Flow:   model.StartFlow,
				Steps:  []model.FlowStep{model.StartFlowStep, model.CreateInitialBalanceFlowStep},
			},
			args: service.GetStateFilter{
				UserID: "test_state_get" + userID,
			},
			expected: &model.State{
				ID:     stateID,
				UserID: "test_state_get" + userID,
				Flow:   model.StartFlow,
				Steps:  []model.FlowStep{model.StartFlowStep, model.CreateInitialBalanceFlowStep},
			},
		},
		{
			desc: "state  by user id not found",
			args: service.GetStateFilter{
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
					err := stateStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := stateStore.Get(ctx, tc.args)
			assert.NoError(t, err)

			if tc.expected == nil {
				assert.Nil(t, actual)
				return
			}

			assert.NotNil(t, actual)
			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.Equal(t, tc.expected.UserID, actual.UserID)
			assert.Equal(t, tc.expected.Flow, actual.Flow)
			assert.Equal(t, tc.expected.Steps, actual.Steps)
		})
	}
}

func TestState_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "state_update")
	userStore := store.NewUser(testCaseDB)
	stateStore := store.NewState(testCaseDB)

	userID := uuid.NewString()
	stateID := uuid.NewString()
	err := userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test_state_update" + userID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions *model.State
		args          *model.State
		expected      *model.State
	}{
		{
			desc: "state updated successfully",
			preconditions: &model.State{
				ID:     stateID,
				UserID: "test_state_update" + userID,
				Flow:   model.StartFlow,
				Steps:  []model.FlowStep{model.StartFlowStep},
			},
			args: &model.State{
				ID:     stateID,
				UserID: "test_state_update" + userID,
				Flow:   model.StartFlow,
				Steps:  []model.FlowStep{model.StartFlowStep, model.CreateInitialBalanceFlowStep},
				Metedata: map[string]any{
					"updated_flow_blabla": "test",
				},
			},
			expected: &model.State{
				ID:     stateID,
				UserID: "test_state_update" + userID,
				Flow:   model.StartFlow,
				Steps:  []model.FlowStep{model.StartFlowStep, model.CreateInitialBalanceFlowStep},
				Metedata: map[string]any{
					"updated_flow_blabla": "test",
				},
			},
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
					err := stateStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := stateStore.Update(ctx, tc.args)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.Equal(t, tc.expected.UserID, actual.UserID)
			assert.Equal(t, tc.expected.Flow, actual.Flow)
			assert.Equal(t, tc.expected.Steps, actual.Steps)
			assert.Equal(t, tc.expected.Metedata, actual.Metedata)
		})
	}
}

func TestState_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	testCaseDB := createTestDB(t, "state_delete")
	userStore := store.NewUser(testCaseDB)
	stateStore := store.NewState(testCaseDB)

	userID1, userID2 := uuid.NewString(), uuid.NewString()
	stateID := uuid.NewString()

	for _, userID := range [...]string{userID1, userID2} {
		err := userStore.Create(ctx, &model.User{
			ID:       userID,
			Username: "test_state_delete" + userID,
		})
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		for _, userID := range [...]string{userID1, userID2} {
			err := deleteUserByID(testCaseDB.DB, userID)
			require.NoError(t, err)
		}
	})

	testCases := [...]struct {
		desc          string
		preconditions *model.State
		args          string
	}{
		{
			desc: "state deleted",
			preconditions: &model.State{
				ID:     stateID,
				UserID: "test_state_delete" + userID1,
				Flow:   model.StartFlow,
			},
			args: stateID,
		},
		{
			desc: "state not deleted because of not existed id",
			preconditions: &model.State{
				ID:     uuid.NewString(),
				UserID: "test_state_delete" + userID2,
				Flow:   model.CancelFlow,
			},
			args: uuid.NewString(),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := stateStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err := stateStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			err := stateStore.Delete(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := stateStore.Get(ctx, service.GetStateFilter{UserID: tc.preconditions.UserID})
			assert.NoError(t, err)

			// operation should not be deleted
			if tc.preconditions.ID != tc.args {
				assert.NotNil(t, actual)
				return
			}

			assert.Nil(t, actual)
		})
	}
}
