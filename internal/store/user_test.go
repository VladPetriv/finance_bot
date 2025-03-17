package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "user_create")
	userStore := store.NewUser(testCaseDB)

	userID := uuid.NewString()

	testCases := [...]struct {
		desc                 string
		preconditions        *models.User
		args                 *models.User
		expectDuplicateError bool
	}{
		{
			desc: "user created",
			args: &models.User{
				ID:       uuid.NewString(),
				Username: "test",
			},
		},
		{
			desc: "user not created because already exist",
			preconditions: &models.User{
				ID:       userID,
				Username: "test_create_2",
			},
			args: &models.User{
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
				err := userStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err := deleteUserByID(testCaseDB.DB, tc.preconditions.ID)
					assert.NoError(t, err)
				}

				err := deleteUserByID(testCaseDB.DB, tc.args.ID)
				assert.NoError(t, err)
			})

			err := userStore.Create(ctx, tc.args)
			if tc.expectDuplicateError {
				assert.Error(t, err)
				assert.True(t, isDuplicateKeyError(err))
				return
			}

			assert.NoError(t, err)

			var createdUser models.User
			err = testCaseDB.DB.Get(&createdUser, "SELECT * FROM users WHERE id=$1;", tc.args.ID)
			assert.NoError(t, err)
			assert.Equal(t, tc.args.ID, createdUser.ID)
			assert.Equal(t, tc.args.Username, createdUser.Username)
		})
	}
}

func TestUser_CreateSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "user_settings_create")
	userStore := store.NewUser(testCaseDB)

	userSettingsID := uuid.NewString()
	user := &models.User{
		ID:       uuid.NewString(),
		Username: "test",
	}

	err := userStore.Create(ctx, user)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := deleteUserByID(testCaseDB.DB, user.ID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc                 string
		preconditions        *models.UserSettings
		args                 *models.UserSettings
		expectDuplicateError bool
	}{
		{
			desc: "user settings created",
			args: &models.UserSettings{
				ID:              uuid.NewString(),
				UserID:          user.ID,
				AIParserEnabled: true,
			},
		},
		{
			desc: "user settings not created because already exist",
			preconditions: &models.UserSettings{
				ID:              userSettingsID,
				UserID:          user.ID,
				AIParserEnabled: false,
			},
			args: &models.UserSettings{
				ID:              userSettingsID,
				UserID:          user.ID,
				AIParserEnabled: false,
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := userStore.CreateSettings(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err := deleteUserSettingsByID(testCaseDB.DB, tc.preconditions.ID)
					assert.NoError(t, err)
				}

				err := deleteUserSettingsByID(testCaseDB.DB, tc.args.ID)
				assert.NoError(t, err)
			})

			err := userStore.CreateSettings(ctx, tc.args)
			if tc.expectDuplicateError {
				assert.Error(t, err)
				assert.True(t, isDuplicateKeyError(err))
				return
			}

			assert.NoError(t, err)

			var createdUserSettings models.UserSettings
			err = testCaseDB.DB.Get(&createdUserSettings, "SELECT * FROM user_settings WHERE id=$1;", tc.args.ID)
			assert.NoError(t, err)
			assert.Equal(t, tc.args.ID, createdUserSettings.ID)
			assert.Equal(t, tc.args.UserID, createdUserSettings.UserID)
			assert.Equal(t, tc.args.AIParserEnabled, createdUserSettings.AIParserEnabled)
		})
	}
}

func TestUser_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "user_get")
	userStore := store.NewUser(testCaseDB)
	currencyStore := store.NewCurrency(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)

	userID, userID2, userID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()
	balanceID, currencyID, userSettingsID := uuid.NewString(), uuid.NewString(), uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &models.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	testCases := [...]struct {
		desc          string
		preconditions *models.User
		args          service.GetUserFilter
		expected      *models.User
	}{
		{
			desc: "found user by username",
			preconditions: &models.User{
				ID:       userID,
				Username: "test",
			},
			args: service.GetUserFilter{
				Username: "test",
			},
			expected: &models.User{
				ID:       userID,
				Username: "test",
			},
		},
		{
			desc: "user with balance preload by username found",
			preconditions: &models.User{
				ID:       userID2,
				Username: "test2",
				Balances: []models.Balance{
					{
						ID:         balanceID,
						UserID:     userID2,
						CurrencyID: currencyID,
						Amount:     "10",
					},
				},
			},
			args: service.GetUserFilter{
				Username:        "test2",
				PreloadBalances: true,
			},
			expected: &models.User{
				ID:       userID2,
				Username: "test2",
				Balances: []models.Balance{
					{
						ID:         balanceID,
						UserID:     userID2,
						CurrencyID: currencyID,
						Amount:     "10",
					},
				},
			},
		},
		{
			desc: "user with settings preload by username found",
			preconditions: &models.User{
				ID:       userID3,
				Username: "test3",
				Settings: &models.UserSettings{
					ID:              userSettingsID,
					UserID:          userID3,
					AIParserEnabled: false,
				},
			},
			args: service.GetUserFilter{
				Username:        "test3",
				PreloadSettings: true,
			},
			expected: &models.User{
				ID:       userID3,
				Username: "test3",
				Settings: &models.UserSettings{
					ID:              userSettingsID,
					UserID:          userID3,
					AIParserEnabled: false,
				},
			},
		},
		{
			desc: "user not found",
			args: service.GetUserFilter{
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

				if tc.preconditions.Settings != nil {
					err := userStore.CreateSettings(ctx, tc.preconditions.Settings)
					assert.NoError(t, err)
				}

				for _, balance := range tc.preconditions.Balances {
					balance.UserID = tc.preconditions.ID
					err = balanceStore.Create(ctx, &balance)
					assert.NoError(t, err)
				}
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					if tc.preconditions.Settings != nil {
						err := deleteUserSettingsByID(testCaseDB.DB, tc.preconditions.Settings.ID)
						assert.NoError(t, err)
					}

					for _, balance := range tc.preconditions.Balances {
						err := deleteBalanceByID(testCaseDB.DB, balance.ID)
						assert.NoError(t, err)
					}

					err := deleteUserByID(testCaseDB.DB, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := userStore.Get(ctx, tc.args)
			assert.NoError(t, err)

			if tc.preconditions == nil {
				assert.Nil(t, actual)
				return
			}

			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.Equal(t, tc.expected.Username, actual.Username)

			// NOTE: We don't care about balances order, since in all test cases we have only one balance.
			for i := range tc.expected.Balances {
				assert.Equal(t, tc.expected.Balances[i].ID, actual.Balances[i].ID)
				assert.Equal(t, tc.expected.Balances[i].UserID, actual.Balances[i].UserID)
				assert.Equal(t, tc.expected.Balances[i].Amount, actual.Balances[i].Amount)
				assert.Equal(t, tc.expected.Balances[i].Currency, actual.Balances[i].Currency)
			}

			if tc.preconditions.Settings != nil {
				assert.Equal(t, tc.expected.Settings.ID, actual.Settings.ID)
				assert.Equal(t, tc.expected.Settings.UserID, actual.Settings.UserID)
				assert.Equal(t, tc.expected.Settings.AIParserEnabled, actual.Settings.AIParserEnabled)
			}
		})
	}
}

func deleteUserByID(db *sqlx.DB, userID string) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1;", userID)
	return err
}

func deleteUserSettingsByID(db *sqlx.DB, userSettingsID string) error {
	_, err := db.Exec("DELETE FROM user_settings WHERE id = $1;", userSettingsID)
	return err
}
