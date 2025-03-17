package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	amount100 = "100.00"
	amount200 = "200.00"
)

func TestBalanceSubscription_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_subscription_create")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	balanceSubscriptionStore := store.NewBalanceSubscription(testCaseDB)

	userID := uuid.NewString()
	balanceID := uuid.NewString()
	categoryID := uuid.NewString()
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

	err = balanceStore.Create(ctx, &models.Balance{
		ID:         balanceID,
		UserID:     userID,
		CurrencyID: currency.ID,
	})
	assert.NoError(t, err)

	err = categoryStore.Create(ctx, &models.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err = balanceStore.Delete(ctx, balanceID)
		require.NoError(t, err)
		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc                 string
		preconditions        *models.BalanceSubscription
		args                 *models.BalanceSubscription
		expectDuplicateError bool
	}{
		{
			desc: "balance subscription created",
			args: &models.BalanceSubscription{
				ID:         uuid.NewString(),
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test",
				Amount:     amount100,
				Period:     models.SubscriptionPeriodMonthly,
				StartAt:    time.Now().Add(24 * time.Hour),
			},
		},
		{
			desc: "duplicate key error because balance already exists",
			preconditions: &models.BalanceSubscription{
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
				err = balanceSubscriptionStore.Create(ctx, *tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err = balanceSubscriptionStore.Delete(ctx, tc.args.ID)
				assert.NoError(t, err)
			})

			err := balanceSubscriptionStore.Create(ctx, *tc.args)
			if tc.expectDuplicateError {
				assert.True(t, isDuplicateKeyError(err))
				return
			}

			assert.NoError(t, err)

			var actual models.BalanceSubscription
			err = testCaseDB.DB.GetContext(ctx, "SELECT * FROM balance_subscriptions WHERE id = $1", tc.args.ID)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.args.ID, actual.ID)
			assert.Equal(t, tc.args.BalanceID, actual.BalanceID)
			assert.Equal(t, tc.args.CategoryID, actual.CategoryID)
			assert.Equal(t, tc.args.Name, actual.Name)
			assert.Equal(t, tc.args.Amount, actual.Amount)
			assert.Equal(t, tc.args.Period, actual.Period)
			assert.Equal(t, tc.args.StartAt, actual.StartAt)
		})
	}
}

// func TestBalance_Get(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background() //nolint: forbidigo

// 	testCaseDB := createTestDB(t, "balance_get")
// 	balanceStore := store.NewBalance(testCaseDB)
// 	userStore := store.NewUser(testCaseDB)
// 	currencyStore := store.NewCurrency(testCaseDB)

// 	balanceID1, balanceID2, balanceID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()
// 	userID1, userID2, userID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()
// 	currency := &models.Currency{
// 		ID:   uuid.NewString(),
// 		Code: "USD",
// 	}

// 	err := currencyStore.CreateIfNotExists(ctx, currency)
// 	require.NoError(t, err)

// 	for _, userID := range [...]string{userID1, userID2, userID3} {
// 		err := userStore.Create(ctx, &models.User{
// 			ID:       userID,
// 			Username: "test" + userID,
// 		})
// 		require.NoError(t, err)
// 	}
// 	t.Cleanup(func() {
// 		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
// 		require.NoError(t, err)

// 		for _, userID := range [...]string{userID1, userID2, userID3} {
// 			err := deleteUserByID(testCaseDB.DB, userID)
// 			require.NoError(t, err)
// 		}
// 	})

// 	testCases := [...]struct {
// 		desc          string
// 		preconditions *models.Balance
// 		args          service.GetBalanceFilter
// 		expected      *models.Balance
// 	}{
// 		{
// 			desc: "balance received by user id",
// 			preconditions: &models.Balance{
// 				ID:         balanceID1,
// 				UserID:     userID1,
// 				CurrencyID: currency.ID,
// 				Amount:     amount300,
// 			},
// 			args: service.GetBalanceFilter{
// 				UserID: userID1,
// 			},
// 			expected: &models.Balance{
// 				ID:         balanceID1,
// 				UserID:     userID1,
// 				CurrencyID: currency.ID,
// 				Amount:     amount300,
// 			},
// 		},
// 		{
// 			desc: "balance received by id with currency preload",
// 			preconditions: &models.Balance{
// 				ID:         balanceID2,
// 				UserID:     userID2,
// 				CurrencyID: currency.ID,
// 				Amount:     amount300,
// 			},
// 			args: service.GetBalanceFilter{
// 				BalanceID:       balanceID2,
// 				PreloadCurrency: true,
// 			},
// 			expected: &models.Balance{
// 				ID:         balanceID2,
// 				UserID:     userID2,
// 				CurrencyID: currency.ID,
// 				Amount:     amount300,
// 				Currency:   *currency,
// 			},
// 		},
// 		{
// 			desc: "balance received by name",
// 			preconditions: &models.Balance{
// 				ID:         balanceID3,
// 				Name:       "test_x3",
// 				UserID:     userID3,
// 				CurrencyID: currency.ID,
// 				Amount:     amount300,
// 			},
// 			args: service.GetBalanceFilter{
// 				Name: "test_x3",
// 			},
// 			expected: &models.Balance{
// 				ID:         balanceID3,
// 				Name:       "test_x3",
// 				CurrencyID: currency.ID,
// 				UserID:     userID3,
// 				Amount:     amount300,
// 			},
// 		},
// 		{
// 			desc: "balance not found",
// 			args: service.GetBalanceFilter{
// 				UserID: uuid.NewString(),
// 			},
// 			expected: nil,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.desc, func(t *testing.T) {
// 			t.Parallel()

// 			if tc.preconditions != nil {
// 				err := balanceStore.Create(ctx, tc.preconditions)
// 				assert.NoError(t, err)
// 			}

// 			t.Cleanup(func() {
// 				if tc.preconditions != nil {
// 					err := balanceStore.Delete(ctx, tc.preconditions.ID)
// 					assert.NoError(t, err)
// 				}
// 			})

// 			actual, err := balanceStore.Get(ctx, tc.args)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tc.expected, actual)
// 		})
// 	}
// }

// func TestBalance_Update(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background() //nolint: forbidigo

// 	testCaseDB := createTestDB(t, "balance_update")
// 	balanceStore := store.NewBalance(testCaseDB)
// 	userStore := store.NewUser(testCaseDB)
// 	currencyStore := store.NewCurrency(testCaseDB)

// 	userID1, userID2 := uuid.NewString(), uuid.NewString()
// 	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
// 	currency := &models.Currency{
// 		ID:   uuid.NewString(),
// 		Code: "USD",
// 	}

// 	err := currencyStore.CreateIfNotExists(ctx, currency)
// 	require.NoError(t, err)

// 	for _, userID := range [...]string{userID1, userID2} {
// 		err := userStore.Create(ctx, &models.User{
// 			ID:       userID,
// 			Username: "test" + userID,
// 		})
// 		require.NoError(t, err)
// 	}
// 	t.Cleanup(func() {
// 		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
// 		require.NoError(t, err)

// 		for _, userID := range [...]string{userID1, userID2} {
// 			err := deleteUserByID(testCaseDB.DB, userID)
// 			require.NoError(t, err)
// 		}
// 	})

// 	testCases := [...]struct {
// 		desc          string
// 		preconditions *models.Balance
// 		args          *models.Balance
// 		expected      *models.Balance
// 	}{
// 		{
// 			desc: "balance updated",
// 			preconditions: &models.Balance{
// 				ID:         balanceID1,
// 				UserID:     userID1,
// 				CurrencyID: currency.ID,
// 				Amount:     amount300,
// 			},
// 			args: &models.Balance{
// 				ID:         balanceID1,
// 				UserID:     userID1,
// 				CurrencyID: currency.ID,
// 				Amount:     amount400,
// 			},
// 			expected: &models.Balance{
// 				ID:         balanceID1,
// 				UserID:     userID1,
// 				CurrencyID: currency.ID,
// 				Amount:     amount400,
// 			},
// 		},
// 		{
// 			desc: "balance not updated because of not existed id",
// 			preconditions: &models.Balance{
// 				ID:         balanceID2,
// 				UserID:     userID2,
// 				CurrencyID: currency.ID,
// 				Amount:     amount300,
// 			},
// 			args: &models.Balance{
// 				ID:         uuid.NewString(),
// 				UserID:     userID2,
// 				CurrencyID: currency.ID,
// 				Amount:     amount400,
// 			},
// 			expected: &models.Balance{
// 				ID:         balanceID2,
// 				UserID:     userID2,
// 				CurrencyID: currency.ID,
// 				Amount:     amount300,
// 			},
// 		},
// 	}
// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.desc, func(t *testing.T) {
// 			t.Parallel()

// 			if tc.preconditions != nil {
// 				err = balanceStore.Create(ctx, tc.preconditions)
// 				assert.NoError(t, err)
// 			}

// 			t.Cleanup(func() {
// 				err = balanceStore.Delete(ctx, tc.preconditions.ID)
// 				assert.NoError(t, err)
// 			})

// 			err := balanceStore.Update(ctx, tc.args)
// 			assert.NoError(t, err)

// 			actual, err := balanceStore.Get(ctx, service.GetBalanceFilter{UserID: tc.preconditions.UserID})
// 			assert.NoError(t, err)
// 			assert.Equal(t, tc.expected, actual)
// 		})
// 	}
// }

// func TestBalance_Delete(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background() //nolint: forbidigo
// 	testCaseDB := createTestDB(t, "balance_delete")
// 	balanceStore := store.NewBalance(testCaseDB)
// 	userStore := store.NewUser(testCaseDB)
// 	currencyStore := store.NewCurrency(testCaseDB)

// 	balanceID, userID := uuid.NewString(), uuid.NewString()
// 	currency := &models.Currency{
// 		ID:   uuid.NewString(),
// 		Code: "USD",
// 	}

// 	err := currencyStore.CreateIfNotExists(ctx, currency)
// 	require.NoError(t, err)

// 	err = userStore.Create(ctx, &models.User{
// 		ID:       userID,
// 		Username: "test" + userID,
// 	})
// 	require.NoError(t, err)

// 	t.Cleanup(func() {
// 		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
// 		require.NoError(t, err)
// 		err = deleteUserByID(testCaseDB.DB, userID)
// 		require.NoError(t, err)
// 	})

// 	testCases := [...]struct {
// 		desc          string
// 		preconditions *models.Balance
// 		args          string
// 	}{
// 		{
// 			desc: "balance deleted",
// 			preconditions: &models.Balance{
// 				ID:         balanceID,
// 				UserID:     userID,
// 				CurrencyID: currency.ID,
// 			},
// 			args: balanceID,
// 		},
// 		{
// 			desc: "balance not deleted because of not existed id",
// 			preconditions: &models.Balance{
// 				ID:         uuid.NewString(),
// 				UserID:     userID,
// 				CurrencyID: currency.ID,
// 			},
// 			args: uuid.NewString(),
// 		},
// 	}
// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.desc, func(t *testing.T) {
// 			t.Parallel()

// 			if tc.preconditions != nil {
// 				err = balanceStore.Create(ctx, tc.preconditions)
// 				assert.NoError(t, err)
// 			}

// 			t.Cleanup(func() {
// 				if tc.preconditions != nil {
// 					err = balanceStore.Delete(ctx, tc.preconditions.ID)
// 					assert.NoError(t, err)
// 				}
// 			})

// 			err := balanceStore.Delete(ctx, tc.args)
// 			assert.NoError(t, err)

// 			actual, err := balanceStore.Get(ctx, service.GetBalanceFilter{BalanceID: tc.preconditions.ID})
// 			assert.NoError(t, err)

// 			// balance should not be deleted
// 			if tc.preconditions.ID != tc.args {
// 				assert.NotNil(t, actual)
// 				return
// 			}

// 			assert.Nil(t, actual)
// 		})
// 	}
// }

// func deleteBalanceByID(db *sqlx.DB, balanceID string) error {
// 	_, err := db.Exec("DELETE FROM balances WHERE id = $1;", balanceID)
// 	return err
// }
