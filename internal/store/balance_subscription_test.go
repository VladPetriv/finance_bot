package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
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
	currencyID := uuid.NewString()
	categoryID := uuid.NewString()
	balanceSubscriptionID := uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &models.Currency{
		ID:   currencyID,
		Code: "USD",
	})

	require.NoError(t, err)

	err = userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	err = balanceStore.Create(ctx, &models.Balance{
		ID:         balanceID,
		UserID:     userID,
		CurrencyID: currencyID,
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
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
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
			},
		},
		{
			desc: "duplicate key error because balance already exists",
			preconditions: &models.BalanceSubscription{
				ID:         balanceSubscriptionID,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test2",
				Period:     models.SubscriptionPeriodMonthly,
				Amount:     amount100,
			},
			args: &models.BalanceSubscription{
				ID:         balanceSubscriptionID,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test2",
				Period:     models.SubscriptionPeriodMonthly,
				Amount:     amount100,
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := balanceSubscriptionStore.Create(ctx, *tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err := balanceSubscriptionStore.Delete(ctx, tc.args.ID)
				assert.NoError(t, err)
			})

			err := balanceSubscriptionStore.Create(ctx, *tc.args)
			if tc.expectDuplicateError {
				assert.True(t, isDuplicateKeyError(err))
				return
			}

			assert.NoError(t, err)

			var actual models.BalanceSubscription
			err = testCaseDB.DB.GetContext(ctx, &actual, "SELECT * FROM balance_subscriptions WHERE id = $1;", tc.args.ID)
			assert.NoError(t, err)
			assert.Equal(t, tc.args.ID, actual.ID)
			assert.Equal(t, tc.args.BalanceID, actual.BalanceID)
			assert.Equal(t, tc.args.CategoryID, actual.CategoryID)
			assert.Equal(t, tc.args.Name, actual.Name)
			assert.Equal(t, tc.args.Amount, actual.Amount)
			assert.Equal(t, tc.args.Period, actual.Period)
		})
	}
}

func TestBalanceSubscription_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_subscription_get")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	balanceSubscriptionStore := store.NewBalanceSubscription(testCaseDB)

	userID := uuid.NewString()
	balanceID := uuid.NewString()
	currencyID := uuid.NewString()
	categoryID := uuid.NewString()
	balanceSubscriptionID := uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &models.Currency{
		ID:   currencyID,
		Code: "USD",
	})

	require.NoError(t, err)

	err = userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	err = balanceStore.Create(ctx, &models.Balance{
		ID:         balanceID,
		UserID:     userID,
		CurrencyID: currencyID,
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
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions *models.BalanceSubscription
		args          service.GetBalanceSubscriptionFilter
		expected      *models.BalanceSubscription
	}{
		{
			desc: "balance subscriptions received by id",
			preconditions: &models.BalanceSubscription{
				ID:         balanceSubscriptionID,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test",
				Amount:     amount100,
				Period:     models.SubscriptionPeriodMonthly,
			},
			args: service.GetBalanceSubscriptionFilter{
				ID: balanceSubscriptionID,
			},
			expected: &models.BalanceSubscription{
				ID:         balanceSubscriptionID,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test",
				Amount:     amount100,
				Period:     models.SubscriptionPeriodMonthly,
			},
		},
		{
			desc: "balance subscription not found",
			args: service.GetBalanceSubscriptionFilter{
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
				err := balanceSubscriptionStore.Create(ctx, *tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err := balanceSubscriptionStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := balanceSubscriptionStore.Get(ctx, tc.args)
			assert.NoError(t, err)

			if tc.expected == nil {
				assert.Nil(t, actual)
				return
			}

			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.Equal(t, tc.expected.BalanceID, actual.BalanceID)
			assert.Equal(t, tc.expected.CategoryID, actual.CategoryID)
			assert.Equal(t, tc.expected.Name, actual.Name)
			assert.Equal(t, tc.expected.Amount, actual.Amount)
			assert.Equal(t, tc.expected.Period, actual.Period)
		})
	}
}

func TestBalanceSubscription_Count(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_subscription_count")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	balanceSubscriptionStore := store.NewBalanceSubscription(testCaseDB)

	userID := uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	currencyID := uuid.NewString()
	categoryID := uuid.NewString()
	balanceSubscriptionID1, balanceSubscriptionID2, balanceSubscriptionID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &models.Currency{
		ID:   currencyID,
		Code: "USD",
	})

	require.NoError(t, err)

	err = userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range []string{balanceID1, balanceID2} {
		err = balanceStore.Create(ctx, &models.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &models.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, balanceID := range []string{balanceID1, balanceID2} {
			err = balanceStore.Delete(ctx, balanceID)
			assert.NoError(t, err)
		}

		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions []models.BalanceSubscription
		args          service.ListBalanceSubscriptionFilter
		expected      int
	}{
		{
			desc: "received a count of balance subscriptions by user id filter",
			preconditions: []models.BalanceSubscription{
				{
					ID:         balanceSubscriptionID1,
					BalanceID:  balanceID1,
					CategoryID: categoryID,
					Name:       "test1",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
				{
					ID:         balanceSubscriptionID2,
					BalanceID:  balanceID2,
					CategoryID: categoryID,
					Name:       "test2",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodYearly,
				},
				{
					ID:         balanceSubscriptionID3,
					BalanceID:  balanceID1,
					CategoryID: categoryID,
					Name:       "test3",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
			},
			args: service.ListBalanceSubscriptionFilter{
				BalanceID: balanceID1,
			},
			expected: 2,
		},
		{
			desc: "balance subscriptions not found",
			args: service.ListBalanceSubscriptionFilter{
				BalanceID: uuid.NewString(),
			},
			expected: 0,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			for _, subscription := range tc.preconditions {
				err := balanceSubscriptionStore.Create(ctx, subscription)
				assert.NoError(t, err)
			}
			t.Cleanup(func() {
				for _, subscription := range tc.preconditions {
					err := balanceSubscriptionStore.Delete(ctx, subscription.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := balanceSubscriptionStore.Count(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestBalanceSubscription_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_subscription_list")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	balanceSubscriptionStore := store.NewBalanceSubscription(testCaseDB)

	userID := uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	currencyID := uuid.NewString()
	categoryID := uuid.NewString()
	balanceSubscriptionID1, balanceSubscriptionID2, balanceSubscriptionID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &models.Currency{
		ID:   currencyID,
		Code: "USD",
	})

	require.NoError(t, err)

	err = userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range []string{balanceID1, balanceID2} {
		err = balanceStore.Create(ctx, &models.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &models.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, balanceID := range []string{balanceID1, balanceID2} {
			err = balanceStore.Delete(ctx, balanceID)
			assert.NoError(t, err)
		}

		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions []models.BalanceSubscription
		args          service.ListBalanceSubscriptionFilter
		expected      []models.BalanceSubscription
	}{
		{
			desc: "received a list of balance subscriptions by user id filter",
			preconditions: []models.BalanceSubscription{
				{
					ID:         balanceSubscriptionID1,
					BalanceID:  balanceID1,
					CategoryID: categoryID,
					Name:       "test1",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
				{
					ID:         balanceSubscriptionID2,
					BalanceID:  balanceID2,
					CategoryID: categoryID,
					Name:       "test2",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodYearly,
				},
				{
					ID:         balanceSubscriptionID3,
					BalanceID:  balanceID1,
					CategoryID: categoryID,
					Name:       "test3",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
			},
			args: service.ListBalanceSubscriptionFilter{
				BalanceID: balanceID1,
			},
			expected: []models.BalanceSubscription{
				{
					ID:         balanceSubscriptionID3,
					BalanceID:  balanceID1,
					CategoryID: categoryID,
					Name:       "test3",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
				{
					ID:         balanceSubscriptionID1,
					BalanceID:  balanceID1,
					CategoryID: categoryID,
					Name:       "test1",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
			},
		},
		{
			desc: "balance subscriptions not found",
			args: service.ListBalanceSubscriptionFilter{
				BalanceID: uuid.NewString(),
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			for _, subscription := range tc.preconditions {
				err := balanceSubscriptionStore.Create(ctx, subscription)
				assert.NoError(t, err)
			}
			t.Cleanup(func() {
				for _, subscription := range tc.preconditions {
					err := balanceSubscriptionStore.Delete(ctx, subscription.ID)
					assert.NoError(t, err)
				}
			})

			tc.args.OrderByCreatedAtDesc = true // Add order to simplify assertion
			actual, err := balanceSubscriptionStore.List(ctx, tc.args)
			assert.NoError(t, err)

			if tc.expected == nil {
				assert.Empty(t, actual)
				return
			}

			assert.Equal(t, len(tc.expected), len(actual))
			for i := range actual {
				assert.Equal(t, tc.expected[i].ID, actual[i].ID)
				assert.Equal(t, tc.expected[i].BalanceID, actual[i].BalanceID)
				assert.Equal(t, tc.expected[i].CategoryID, actual[i].CategoryID)
				assert.Equal(t, tc.expected[i].Name, actual[i].Name)
				assert.Equal(t, tc.expected[i].Amount, actual[i].Amount)
				assert.Equal(t, tc.expected[i].Period, actual[i].Period)
			}
		})
	}
}

func TestBalanceSubscription_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_subscription_update")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	balanceSubscriptionStore := store.NewBalanceSubscription(testCaseDB)

	userID := uuid.NewString()
	balanceID := uuid.NewString()
	currencyID := uuid.NewString()
	categoryID := uuid.NewString()
	balanceSubscriptionID1, balanceSubscriptionID2 := uuid.NewString(), uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &models.Currency{
		ID:   currencyID,
		Code: "USD",
	})

	require.NoError(t, err)

	err = userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	err = balanceStore.Create(ctx, &models.Balance{
		ID:         balanceID,
		UserID:     userID,
		CurrencyID: currencyID,
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
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions *models.BalanceSubscription
		args          *models.BalanceSubscription
		expected      *models.BalanceSubscription
	}{
		{
			desc: "balance subscription updated",
			preconditions: &models.BalanceSubscription{
				ID:         balanceSubscriptionID1,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test1",
				Amount:     amount100,
				Period:     models.SubscriptionPeriodMonthly,
			},
			args: &models.BalanceSubscription{
				ID:         balanceSubscriptionID1,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test1",
				Amount:     amount200,
				Period:     models.SubscriptionPeriodWeekly,
			},
			expected: &models.BalanceSubscription{
				ID:         balanceSubscriptionID1,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test1",
				Amount:     amount200,
				Period:     models.SubscriptionPeriodWeekly,
			},
		},
		{
			desc: "balance subscription not updated because of not existed id",
			preconditions: &models.BalanceSubscription{
				ID:         balanceSubscriptionID2,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test2",
				Amount:     amount100,
				Period:     models.SubscriptionPeriodMonthly,
			},
			args: &models.BalanceSubscription{
				ID:         uuid.NewString(),
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test2",
				Amount:     amount200,
				Period:     models.SubscriptionPeriodYearly,
			},
			expected: &models.BalanceSubscription{
				ID:         balanceSubscriptionID2,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test2",
				Amount:     amount100,
				Period:     models.SubscriptionPeriodMonthly,
			},
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
				err = balanceSubscriptionStore.Delete(ctx, tc.preconditions.ID)
				assert.NoError(t, err)
			})

			err := balanceSubscriptionStore.Update(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := balanceSubscriptionStore.Get(ctx, service.GetBalanceSubscriptionFilter{
				Name: tc.preconditions.Name,
			})
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.Equal(t, tc.expected.BalanceID, actual.BalanceID)
			assert.Equal(t, tc.expected.CategoryID, actual.CategoryID)
			assert.Equal(t, tc.expected.Name, actual.Name)
			assert.Equal(t, tc.expected.Amount, actual.Amount)
			assert.Equal(t, tc.expected.Period, actual.Period)
		})
	}
}

func TestBalanceSubscription_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_subscription_delete")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	balanceSubscriptionStore := store.NewBalanceSubscription(testCaseDB)

	userID := uuid.NewString()
	balanceID := uuid.NewString()
	currencyID := uuid.NewString()
	categoryID := uuid.NewString()
	balanceSubscriptionID := uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &models.Currency{
		ID:   currencyID,
		Code: "USD",
	})

	require.NoError(t, err)

	err = userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	err = balanceStore.Create(ctx, &models.Balance{
		ID:         balanceID,
		UserID:     userID,
		CurrencyID: currencyID,
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
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions *models.BalanceSubscription
		args          string
	}{
		{
			desc: "balance subscription deleted",
			preconditions: &models.BalanceSubscription{
				ID:         balanceSubscriptionID,
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test_delete",
				Amount:     amount100,
				Period:     models.SubscriptionPeriodMonthly,
				StartAt:    time.Now().Add(1 * time.Hour),
			},
			args: balanceSubscriptionID,
		},
		{
			desc: "balance not deleted because of not existed id",
			preconditions: &models.BalanceSubscription{
				ID:         uuid.NewString(),
				BalanceID:  balanceID,
				CategoryID: categoryID,
				Name:       "test_delete",
				Amount:     amount100,
				Period:     models.SubscriptionPeriodMonthly,
				StartAt:    time.Now().Add(1 * time.Hour),
			},
			args: uuid.NewString(),
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
				if tc.preconditions != nil {
					err = balanceSubscriptionStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			err := balanceSubscriptionStore.Delete(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := balanceSubscriptionStore.Get(ctx, service.GetBalanceSubscriptionFilter{
				ID: tc.preconditions.ID,
			})
			assert.NoError(t, err)

			// balance subscription should not be deleted
			if tc.preconditions.ID != tc.args {
				assert.NotNil(t, actual)
				return
			}

			assert.Nil(t, actual)
		})
	}
}
