package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
	balanceID1, balanceID2,
		balanceID3, balanceID4 := uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString()
	currencyID := uuid.NewString()
	categoryID := uuid.NewString()
	balanceSubscriptionID1, balanceSubscriptionID2, balanceSubscriptionID3,
		balanceSubscriptionID4, balanceSubscriptionID5, balanceSubscriptionID6,
		balanceSubscriptionID7, balanceSubscriptionID8, balanceSubscriptionID9,
		balanceSubscriptionID10, balanceSubscriptionID11 := uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString()

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

	for _, balanceID := range []string{balanceID1, balanceID2, balanceID3, balanceID4} {
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
		for _, balanceID := range []string{balanceID1, balanceID2, balanceID3, balanceID4} {
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
			desc: "received a list of balance subscriptions with pagination: page: 1, limit: 2, total: 4",
			preconditions: []models.BalanceSubscription{
				{
					ID:         balanceSubscriptionID4,
					BalanceID:  balanceID3,
					CategoryID: categoryID,
					Name:       "test1",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
				{
					ID:         balanceSubscriptionID5,
					BalanceID:  balanceID3,
					CategoryID: categoryID,
					Name:       "test2",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodYearly,
				},
				{
					ID:         balanceSubscriptionID6,
					BalanceID:  balanceID3,
					CategoryID: categoryID,
					Name:       "test3",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
				{
					ID:         balanceSubscriptionID7,
					BalanceID:  balanceID3,
					CategoryID: categoryID,
					Name:       "test4",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
			},
			args: service.ListBalanceSubscriptionFilter{
				BalanceID: balanceID3,
				Pagination: &service.Pagination{
					Page:  1,
					Limit: 2,
				},
			},
			expected: []models.BalanceSubscription{
				{
					ID:         balanceSubscriptionID7,
					BalanceID:  balanceID3,
					CategoryID: categoryID,
					Name:       "test4",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
				{
					ID:         balanceSubscriptionID6,
					BalanceID:  balanceID3,
					CategoryID: categoryID,
					Name:       "test3",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
			},
		},
		{
			desc: "received a list of balance subscriptions with pagination: page: 2, limit: 2, total: 4",
			preconditions: []models.BalanceSubscription{
				{
					ID:         balanceSubscriptionID8,
					BalanceID:  balanceID4,
					CategoryID: categoryID,
					Name:       "test1",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
				{
					ID:         balanceSubscriptionID9,
					BalanceID:  balanceID4,
					CategoryID: categoryID,
					Name:       "test2",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodYearly,
				},
				{
					ID:         balanceSubscriptionID10,
					BalanceID:  balanceID4,
					CategoryID: categoryID,
					Name:       "test3",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
				{
					ID:         balanceSubscriptionID11,
					BalanceID:  balanceID4,
					CategoryID: categoryID,
					Name:       "test4",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodWeekly,
				},
			},
			args: service.ListBalanceSubscriptionFilter{
				BalanceID: balanceID4,
				Pagination: &service.Pagination{
					Page:  2,
					Limit: 2,
				},
			},
			expected: []models.BalanceSubscription{
				{
					ID:         balanceSubscriptionID9,
					BalanceID:  balanceID4,
					CategoryID: categoryID,
					Name:       "test2",
					Amount:     amount100,
					Period:     models.SubscriptionPeriodYearly,
				},
				{
					ID:         balanceSubscriptionID8,
					BalanceID:  balanceID4,
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

func TestBalanceSubscription_CreateScheduledOperation(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_subscription_create_scheduled_operations")
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
	scheduledOperationID := uuid.NewString()

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

	err = balanceSubscriptionStore.Create(ctx, models.BalanceSubscription{
		ID:         balanceSubscriptionID,
		BalanceID:  balanceID,
		CategoryID: categoryID,
		Name:       "test",
		Amount:     amount100,
		Period:     models.SubscriptionPeriodMonthly,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err = balanceSubscriptionStore.Delete(ctx, balanceSubscriptionID)
		require.NoError(t, err)
		err = balanceStore.Delete(ctx, balanceID)
		require.NoError(t, err)
		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
		err = balanceSubscriptionStore.Delete(ctx, balanceSubscriptionID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc                 string
		preconditions        *models.ScheduledOperation
		args                 *models.ScheduledOperation
		expectDuplicateError bool
	}{
		{
			desc: "scheduled operation created",
			args: &models.ScheduledOperation{
				ID:             uuid.NewString(),
				SubscriptionID: balanceSubscriptionID,
				CreationDate:   time.Date(2025, time.May, 2, 13, 12, 0, 0, time.UTC),
			},
		},
		{
			desc: "duplicate key error because scheduled operation already exists",
			preconditions: &models.ScheduledOperation{
				ID:             scheduledOperationID,
				SubscriptionID: balanceSubscriptionID,
				CreationDate:   time.Date(2025, time.May, 1, 12, 12, 0, 0, time.UTC),
			},
			args: &models.ScheduledOperation{
				ID:             scheduledOperationID,
				SubscriptionID: balanceSubscriptionID,
				CreationDate:   time.Date(2025, time.May, 1, 12, 12, 0, 0, time.UTC),
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := balanceSubscriptionStore.CreateScheduledOperation(ctx, *tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err = balanceSubscriptionStore.DeleteScheduledOperation(ctx, tc.args.ID)
				assert.NoError(t, err)
			})

			err := balanceSubscriptionStore.CreateScheduledOperation(ctx, *tc.args)
			if tc.expectDuplicateError {
				assert.True(t, isDuplicateKeyError(err))
				return
			}

			assert.NoError(t, err)

			var actual models.ScheduledOperation
			err = testCaseDB.DB.GetContext(ctx, &actual, "SELECT * FROM scheduled_operations WHERE id = $1;", tc.args.ID)
			assert.NoError(t, err)
			assert.Equal(t, tc.args.ID, actual.ID)
			assert.Equal(t, tc.args.SubscriptionID, actual.SubscriptionID)
			assert.Equal(t, tc.args.CreationDate, actual.CreationDate.UTC())
		})
	}
}

func TestBalanceSubscription_ListScheduledOperation(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_subscription_list_scheduled_operations")
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
	scheduledOperationID1, scheduledOperationID2,
		scheduledOperationID3, scheduledOperationID4 := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()

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

	err = balanceSubscriptionStore.Create(ctx, models.BalanceSubscription{
		ID:         balanceSubscriptionID,
		BalanceID:  balanceID,
		CategoryID: categoryID,
		Name:       "test",
		Amount:     amount100,
		Period:     models.SubscriptionPeriodMonthly,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err = balanceSubscriptionStore.Delete(ctx, balanceSubscriptionID)
		require.NoError(t, err)
		err = balanceStore.Delete(ctx, balanceID)
		require.NoError(t, err)
		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
		err = balanceSubscriptionStore.Delete(ctx, balanceSubscriptionID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions []models.ScheduledOperation
		args          service.ListScheduledOperation
		expected      []models.ScheduledOperation
	}{
		{
			desc: "received scheduled operation with date greater than filter",
			preconditions: []models.ScheduledOperation{
				{
					ID:             scheduledOperationID1,
					SubscriptionID: balanceSubscriptionID,
					CreationDate:   time.Date(2025, time.March, 11, 10, 0, 0, 0, time.UTC),
				},
				{
					ID:             scheduledOperationID2,
					SubscriptionID: balanceSubscriptionID,
					CreationDate:   time.Date(2025, time.March, 11, 11, 0, 0, 0, time.UTC),
				},
				{
					ID:             scheduledOperationID3,
					SubscriptionID: balanceSubscriptionID,
					CreationDate:   time.Date(2025, time.March, 11, 13, 0, 0, 0, time.UTC),
				},
				{
					ID:             scheduledOperationID4,
					SubscriptionID: balanceSubscriptionID,
					CreationDate:   time.Date(2025, time.March, 11, 14, 0, 0, 0, time.UTC),
				},
			},
			args: service.ListScheduledOperation{
				BetweenFilter: &service.BetweenFilter{
					From: time.Date(2025, time.March, 11, 11, 0, 0, 0, time.UTC),
					To:   time.Date(2025, time.March, 11, 13, 0, 0, 0, time.UTC),
				},
			},
			expected: []models.ScheduledOperation{
				{
					ID:             scheduledOperationID2,
					SubscriptionID: balanceSubscriptionID,
					CreationDate:   time.Date(2025, time.March, 11, 11, 0, 0, 0, time.UTC),
				},
				{
					ID:             scheduledOperationID3,
					SubscriptionID: balanceSubscriptionID,
					CreationDate:   time.Date(2025, time.March, 11, 13, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			for _, scheduledOperation := range tc.preconditions {
				err := balanceSubscriptionStore.CreateScheduledOperation(ctx, scheduledOperation)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				for _, scheduledOperation := range tc.preconditions {
					err = balanceSubscriptionStore.DeleteScheduledOperation(ctx, scheduledOperation.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := balanceSubscriptionStore.ListScheduledOperation(ctx, tc.args)
			assert.NoError(t, err)

			for i := range actual {
				assert.Equal(t, tc.expected[i].ID, actual[i].ID)
				assert.Equal(t, tc.expected[i].SubscriptionID, actual[i].SubscriptionID)
				assert.Equal(t, tc.expected[i].CreationDate, actual[i].CreationDate.UTC())
			}
		})
	}
}

func TestBalanceSubscription_DeleteScheduledOperation(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_subscription_delete_scheduled_operations")
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
	scheduledOperationID := uuid.NewString()

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
	require.NoError(t, err)

	err = categoryStore.Create(ctx, &models.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	err = balanceSubscriptionStore.Create(ctx, models.BalanceSubscription{
		ID:         balanceSubscriptionID,
		BalanceID:  balanceID,
		CategoryID: categoryID,
		Name:       "test",
		Amount:     amount100,
		Period:     models.SubscriptionPeriodMonthly,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err = balanceSubscriptionStore.Delete(ctx, balanceSubscriptionID)
		require.NoError(t, err)
		err = balanceStore.Delete(ctx, balanceID)
		require.NoError(t, err)
		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
		err = balanceSubscriptionStore.Delete(ctx, balanceSubscriptionID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions *models.ScheduledOperation
		args          string
	}{
		{
			desc: "scheduled operation deleted",
			preconditions: &models.ScheduledOperation{
				ID:             scheduledOperationID,
				SubscriptionID: balanceSubscriptionID,
			},
			args: balanceSubscriptionID,
		},
		{
			desc: "scheduled operation not deleted because of not existed id",
			preconditions: &models.ScheduledOperation{
				ID:             uuid.NewString(),
				SubscriptionID: balanceSubscriptionID,
			},
			args: uuid.NewString(),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = balanceSubscriptionStore.CreateScheduledOperation(ctx, *tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err = balanceSubscriptionStore.DeleteScheduledOperation(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			err := balanceSubscriptionStore.DeleteScheduledOperation(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := getScheledOperationByID(testCaseDB.DB, tc.preconditions.ID)
			assert.NoError(t, err)

			// scheduled opration should not be deleted
			if tc.preconditions.ID != tc.args {
				assert.NotNil(t, actual)
				return
			}

			assert.Nil(t, actual)
		})
	}
}

func getScheledOperationByID(db *sqlx.DB, id string) (*models.ScheduledOperation, error) {
	var scheduleOperation models.ScheduledOperation
	err := db.Get(&scheduleOperation, "SELECT * FROM scheduled_operations WHERE id = $1;", id)
	return &scheduleOperation, err
}
