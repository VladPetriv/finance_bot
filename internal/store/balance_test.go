package store_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	amount300 = "300.00"
	amount400 = "400.00"
)

func TestBalance_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_create")
	balanceStore := store.NewBalance(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	currencyStore := store.NewCurrency(testCaseDB)

	balanceID, userID := uuid.NewString(), uuid.NewString()
	currency := &models.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc                 string
		preconditions        *models.Balance
		args                 *models.Balance
		expectDuplicateError bool
	}{
		{
			desc: "balance created",
			args: &models.Balance{
				ID:         uuid.NewString(),
				UserID:     userID,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
		},
		{
			desc: "duplicate key error because balance already exists",
			preconditions: &models.Balance{
				ID:         balanceID,
				UserID:     userID,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: &models.Balance{
				ID:         balanceID,
				UserID:     userID,
				CurrencyID: currency.ID,
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
				err = balanceStore.Delete(ctx, tc.args.ID)
				assert.NoError(t, err)
			})

			err := balanceStore.Create(ctx, tc.args)
			if tc.expectDuplicateError {
				assert.True(t, isDuplicateKeyError(err))
				return
			}

			assert.NoError(t, err)

			actual, err := balanceStore.Get(ctx, service.GetBalanceFilter{BalanceID: tc.args.ID})
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.args.ID, actual.ID)
			assert.Equal(t, tc.args.UserID, actual.UserID)
			assert.Equal(t, tc.args.CurrencyID, actual.CurrencyID)
			assert.Equal(t, tc.args.Name, actual.Name)
			assert.Equal(t, tc.args.Amount, actual.Amount)
		})
	}
}

func TestBalance_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_get")
	balanceStore := store.NewBalance(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	currencyStore := store.NewCurrency(testCaseDB)

	balanceID1, balanceID2, balanceID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()
	userID1, userID2, userID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()
	currency := &models.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	for _, userID := range [...]string{userID1, userID2, userID3} {
		err := userStore.Create(ctx, &models.User{
			ID:       userID,
			Username: "test" + userID,
		})
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)

		for _, userID := range [...]string{userID1, userID2, userID3} {
			err := deleteUserByID(testCaseDB.DB, userID)
			require.NoError(t, err)
		}
	})

	testCases := [...]struct {
		desc          string
		preconditions *models.Balance
		args          service.GetBalanceFilter
		expected      *models.Balance
	}{
		{
			desc: "balance received by user id",
			preconditions: &models.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: service.GetBalanceFilter{
				UserID: userID1,
			},
			expected: &models.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
		},
		{
			desc: "balance received by id with currency preload",
			preconditions: &models.Balance{
				ID:         balanceID2,
				UserID:     userID2,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: service.GetBalanceFilter{
				BalanceID:       balanceID2,
				PreloadCurrency: true,
			},
			expected: &models.Balance{
				ID:         balanceID2,
				UserID:     userID2,
				CurrencyID: currency.ID,
				Amount:     amount300,
				Currency:   *currency,
			},
		},
		{
			desc: "balance received by name",
			preconditions: &models.Balance{
				ID:         balanceID3,
				Name:       "test_x3",
				UserID:     userID3,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: service.GetBalanceFilter{
				Name: "test_x3",
			},
			expected: &models.Balance{
				ID:         balanceID3,
				Name:       "test_x3",
				CurrencyID: currency.ID,
				UserID:     userID3,
				Amount:     amount300,
			},
		},
		{
			desc: "balance not found",
			args: service.GetBalanceFilter{
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
					err := balanceStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := balanceStore.Get(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestBalance_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_update")
	balanceStore := store.NewBalance(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	currencyStore := store.NewCurrency(testCaseDB)

	userID1, userID2 := uuid.NewString(), uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	currency := &models.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	for _, userID := range [...]string{userID1, userID2} {
		err := userStore.Create(ctx, &models.User{
			ID:       userID,
			Username: "test" + userID,
		})
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)

		for _, userID := range [...]string{userID1, userID2} {
			err := deleteUserByID(testCaseDB.DB, userID)
			require.NoError(t, err)
		}
	})

	testCases := [...]struct {
		desc          string
		preconditions *models.Balance
		args          *models.Balance
		expected      *models.Balance
	}{
		{
			desc: "balance updated",
			preconditions: &models.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: &models.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount400,
			},
			expected: &models.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount400,
			},
		},
		{
			desc: "balance not updated because of not existed id",
			preconditions: &models.Balance{
				ID:         balanceID2,
				UserID:     userID2,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: &models.Balance{
				ID:         uuid.NewString(),
				UserID:     userID2,
				CurrencyID: currency.ID,
				Amount:     amount400,
			},
			expected: &models.Balance{
				ID:         balanceID2,
				UserID:     userID2,
				CurrencyID: currency.ID,
				Amount:     amount300,
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

			err := balanceStore.Update(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := balanceStore.Get(ctx, service.GetBalanceFilter{UserID: tc.preconditions.UserID})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestBalance_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "balance_delete")
	balanceStore := store.NewBalance(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	currencyStore := store.NewCurrency(testCaseDB)

	balanceID, userID := uuid.NewString(), uuid.NewString()
	currency := &models.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions *models.Balance
		args          string
	}{
		{
			desc: "balance deleted",
			preconditions: &models.Balance{
				ID:         balanceID,
				UserID:     userID,
				CurrencyID: currency.ID,
			},
			args: balanceID,
		},
		{
			desc: "balance not deleted because of not existed id",
			preconditions: &models.Balance{
				ID:         uuid.NewString(),
				UserID:     userID,
				CurrencyID: currency.ID,
			},
			args: uuid.NewString(),
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
				if tc.preconditions != nil {
					err = balanceStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			err := balanceStore.Delete(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := balanceStore.Get(ctx, service.GetBalanceFilter{BalanceID: tc.preconditions.ID})
			assert.NoError(t, err)

			// balance should not be deleted
			if tc.preconditions.ID != tc.args {
				assert.NotNil(t, actual)
				return
			}

			assert.Nil(t, actual)
		})
	}
}

func deleteBalanceByID(db *sqlx.DB, balanceID string) error {
	_, err := db.Exec("DELETE FROM balances WHERE id = $1;", balanceID)
	return err
}
