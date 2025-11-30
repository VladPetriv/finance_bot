package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutomaticReport_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "automatic_report_create")

	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	automaticReportStore := store.NewAutomaticReport(testCaseDB)

	userID := uuid.NewString()
	currencyID := uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	categoryID1, categoryID2 := uuid.NewString(), uuid.NewString()
	automaticReportID := uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &model.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range []string{balanceID1, balanceID2} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}

	for _, categoryID := range []string{categoryID1, categoryID2} {
		err = categoryStore.Create(ctx, &model.Category{
			ID:     categoryID,
			UserID: userID,
			Title:  "test_category",
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, categoryID := range []string{categoryID1, categoryID2} {
			err = categoryStore.Delete(ctx, categoryID)
			require.NoError(t, err)
		}
		for _, balanceID := range []string{balanceID1, balanceID2} {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc                 string
		precondition         *model.AutomaticReport
		args                 *model.AutomaticReport
		expectDuplicateError bool
	}{
		{
			desc: "Automatic report with one balance and one category created",
			args: &model.AutomaticReport{
				ID:          uuid.NewString(),
				BalanceIDs:  []string{balanceID1},
				CategoryIDs: []string{categoryID1},
				Name:        "test_automatic_report_create_1",
				Period:      model.AutomaticReportPeriodDaily,
			},
		},
		{
			desc: "Automatic report with two balances and two categories created",
			args: &model.AutomaticReport{
				ID:          uuid.NewString(),
				BalanceIDs:  []string{balanceID1, balanceID2},
				CategoryIDs: []string{categoryID1, categoryID2},
				Name:        "test_automatic_report_create_2",
				Period:      model.AutomaticReportPeriodWeekly,
			},
		},
		{
			desc: "Automatic report not created due to duplicate key error",
			precondition: &model.AutomaticReport{
				ID:          automaticReportID,
				BalanceIDs:  []string{balanceID1},
				CategoryIDs: []string{categoryID1},
				Name:        "test_automatic_report_create_3",
				Period:      model.AutomaticReportPeriodMonthly,
			},
			args: &model.AutomaticReport{
				ID:          automaticReportID,
				BalanceIDs:  []string{balanceID1},
				CategoryIDs: []string{categoryID1},
				Name:        "test_automatic_report_create_3",
				Period:      model.AutomaticReportPeriodMonthly,
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.precondition != nil {
				err := automaticReportStore.Create(ctx, tc.precondition)
				assert.NoError(t, err)
			}
			t.Cleanup(func() {
				err := automaticReportStore.Delete(ctx, tc.args.ID)
				assert.NoError(t, err)
			})

			err := automaticReportStore.Create(ctx, tc.args)
			if tc.expectDuplicateError {
				assert.True(t, isDuplicateKeyError(err))
				return
			}
			assert.NoError(t, err)

			actual, err := automaticReportStore.Get(ctx, service.GetAutomaticReportFilter{ID: tc.args.ID})
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.args.ID, actual.ID)
			assert.ElementsMatch(t, tc.args.BalanceIDs, actual.BalanceIDs)
			assert.ElementsMatch(t, tc.args.CategoryIDs, actual.CategoryIDs)
			assert.Equal(t, tc.args.Name, actual.Name)
			assert.Equal(t, tc.args.Period, actual.Period)
		})
	}
}

func TestAutomaticReport_CreateScheduledReportExecution(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "automatic_report_create_scheduled_report_execution")

	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	automaticReportStore := store.NewAutomaticReport(testCaseDB)

	userID := uuid.NewString()
	currencyID := uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	categoryID1, categoryID2 := uuid.NewString(), uuid.NewString()
	automaticReportID := uuid.NewString()
	scheduledReportExecutionID := uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &model.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range []string{balanceID1, balanceID2} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}

	for _, categoryID := range []string{categoryID1, categoryID2} {
		err = categoryStore.Create(ctx, &model.Category{
			ID:     categoryID,
			UserID: userID,
			Title:  "test_category",
		})
		require.NoError(t, err)
	}

	err = automaticReportStore.Create(ctx, &model.AutomaticReport{
		ID:          automaticReportID,
		BalanceIDs:  []string{balanceID1, balanceID2},
		CategoryIDs: []string{categoryID1, categoryID2},
		Name:        "test",
		Period:      model.AutomaticReportPeriodDaily,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := automaticReportStore.Delete(ctx, automaticReportID)
		require.NoError(t, err)
		for _, categoryID := range []string{categoryID1, categoryID2} {
			err = categoryStore.Delete(ctx, categoryID)
			require.NoError(t, err)
		}
		for _, balanceID := range []string{balanceID1, balanceID2} {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err = deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc                 string
		precondition         *model.ScheduledReportExecution
		args                 *model.ScheduledReportExecution
		expectDuplicateError bool
	}{
		{
			desc: "Scheduled report execution created",
			args: &model.ScheduledReportExecution{
				ID:                uuid.NewString(),
				AutomaticReportID: automaticReportID,
				ExecutionDate:     time.Now().UTC().Add(24 * time.Hour),
			},
		},
		{
			desc: "Scheduled report execution not created due to duplicate key error",
			precondition: &model.ScheduledReportExecution{
				ID:                scheduledReportExecutionID,
				AutomaticReportID: automaticReportID,
				ExecutionDate:     time.Now().Add(24 * time.Hour),
			},
			args: &model.ScheduledReportExecution{
				ID:                scheduledReportExecutionID,
				AutomaticReportID: automaticReportID,
				ExecutionDate:     time.Now().Add(24 * time.Hour),
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.precondition != nil {
				err := automaticReportStore.CreateScheduledReportExecution(ctx, tc.precondition)
				assert.NoError(t, err)
			}
			t.Cleanup(func() {
				err := automaticReportStore.DeleteScheduledReportExecution(ctx, tc.args.ID)
				assert.NoError(t, err)
			})

			err := automaticReportStore.CreateScheduledReportExecution(ctx, tc.args)
			if tc.expectDuplicateError {
				assert.True(t, isDuplicateKeyError(err))
				return
			}
			assert.NoError(t, err)

			var actual model.ScheduledReportExecution
			err = testCaseDB.DB.Get(&actual, "SELECT * FROM scheduled_report_executions WHERE id = $1", tc.args.ID)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.args.ID, actual.ID)
			assert.Equal(t, tc.args.AutomaticReportID, actual.AutomaticReportID)
			assert.Equal(t, tc.args.ExecutionDate.UTC(), actual.ExecutionDate.UTC())
		})
	}
}

func TestAutomaticReport_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "automatic_report_get")

	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	automaticReportStore := store.NewAutomaticReport(testCaseDB)

	userID := uuid.NewString()
	currencyID := uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	categoryID1, categoryID2 := uuid.NewString(), uuid.NewString()
	automaticReportID1, automaticReportID2 := uuid.NewString(), uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &model.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range []string{balanceID1, balanceID2} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}

	for _, categoryID := range []string{categoryID1, categoryID2} {
		err = categoryStore.Create(ctx, &model.Category{
			ID:     categoryID,
			UserID: userID,
			Title:  "test_category",
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, categoryID := range []string{categoryID1, categoryID2} {
			err = categoryStore.Delete(ctx, categoryID)
			require.NoError(t, err)
		}
		for _, balanceID := range []string{balanceID1, balanceID2} {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc         string
		precondition *model.AutomaticReport
		args         service.GetAutomaticReportFilter
		expected     *model.AutomaticReport
	}{
		{
			desc: "Automatic report found by id filter",
			precondition: &model.AutomaticReport{
				ID:          automaticReportID1,
				BalanceIDs:  []string{balanceID1},
				CategoryIDs: []string{categoryID1},
				Name:        "test_get_1",
				Period:      model.AutomaticReportPeriodDaily,
			},
			args: service.GetAutomaticReportFilter{
				ID: automaticReportID1,
			},
			expected: &model.AutomaticReport{
				ID:          automaticReportID1,
				BalanceIDs:  []string{balanceID1},
				CategoryIDs: []string{categoryID1},
				Name:        "test_get_1",
				Period:      model.AutomaticReportPeriodDaily,
			},
		},
		{
			desc: "Automatic report found by name filter",
			precondition: &model.AutomaticReport{
				ID:          automaticReportID2,
				BalanceIDs:  []string{balanceID2},
				CategoryIDs: []string{categoryID2},
				Name:        "test_get_2",
				Period:      model.AutomaticReportPeriodDaily,
			},
			args: service.GetAutomaticReportFilter{
				Name: "test_get_2",
			},
			expected: &model.AutomaticReport{
				ID:          automaticReportID2,
				BalanceIDs:  []string{balanceID2},
				CategoryIDs: []string{categoryID2},
				Name:        "test_get_2",
				Period:      model.AutomaticReportPeriodDaily,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.precondition != nil {
				err := automaticReportStore.Create(ctx, tc.precondition)
				assert.NoError(t, err)

				t.Cleanup(func() {
					err := automaticReportStore.Delete(ctx, tc.precondition.ID)
					assert.NoError(t, err)
				})
			}

			actual, err := automaticReportStore.Get(ctx, tc.args)
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.ElementsMatch(t, tc.expected.BalanceIDs, actual.BalanceIDs)
			assert.ElementsMatch(t, tc.expected.CategoryIDs, actual.CategoryIDs)
			assert.Equal(t, tc.expected.Name, actual.Name)
			assert.Equal(t, tc.expected.Period, actual.Period)
		})
	}
}

func TestAutomaticReport_Count(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "automatic_report_count")

	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	automaticReportStore := store.NewAutomaticReport(testCaseDB)

	userID := uuid.NewString()
	currencyID := uuid.NewString()
	balanceID1, balanceID2, balanceID3,
		balanceID4, balanceID5, balanceID6 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString()
	categoryID1, categoryID2, categoryID3,
		categoryID4, categoryID5, categoryID6,
		categoryID7 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString()
	automaticReportID1, automaticReportID2, automaticReportID3,
		automaticReportID4, automaticReportID5, automaticReportID6,
		automaticReportID7, automaticReportID8, automaticReportID9 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &model.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range []string{
		balanceID1, balanceID2, balanceID3,
		balanceID4, balanceID5, balanceID6,
	} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}
	for _, categoryID := range []string{
		categoryID1, categoryID2, categoryID3, categoryID4,
		categoryID5, categoryID6, categoryID7,
	} {
		err = categoryStore.Create(ctx, &model.Category{
			ID:     categoryID,
			UserID: userID,
			Title:  "test_category",
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, categoryID := range []string{
			categoryID1, categoryID2, categoryID3, categoryID4,
			categoryID5, categoryID6, categoryID7,
		} {
			err = categoryStore.Delete(ctx, categoryID)
			require.NoError(t, err)
		}
		for _, balanceID := range []string{
			balanceID1, balanceID2, balanceID3,
			balanceID4, balanceID5, balanceID6,
		} {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc         string
		precondition []model.AutomaticReport
		args         service.ListAutomaticReportFilter
		expected     int
	}{
		{
			desc: "Should return count of 2 automatic reports with balance id filter",
			precondition: []model.AutomaticReport{
				{
					ID:          automaticReportID1,
					BalanceIDs:  []string{balanceID1},
					CategoryIDs: []string{categoryID1},
					Name:        "test_count_1",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID2,
					BalanceIDs:  []string{balanceID2},
					CategoryIDs: []string{categoryID2},
					Name:        "test_count_2",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID3,
					BalanceIDs:  []string{balanceID1, balanceID2},
					CategoryIDs: []string{categoryID1, categoryID2},
					Name:        "test_count_3",
					Period:      model.AutomaticReportPeriodDaily,
				},
			},
			args: service.ListAutomaticReportFilter{
				BalanceIDs: []string{balanceID1},
			},
			expected: 2,
		},
		{
			desc: "Should return count of 2 automatic reports with category id filter",
			precondition: []model.AutomaticReport{
				{
					ID:          automaticReportID4,
					BalanceIDs:  []string{balanceID3},
					CategoryIDs: []string{categoryID3},
					Name:        "test_count_4",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID5,
					BalanceIDs:  []string{balanceID3},
					CategoryIDs: []string{categoryID4},
					Name:        "test_count_5",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID6,
					BalanceIDs:  []string{balanceID3},
					CategoryIDs: []string{categoryID3, categoryID4},
					Name:        "test_count_6",
					Period:      model.AutomaticReportPeriodDaily,
				},
			},
			args: service.ListAutomaticReportFilter{
				CategoryIDs: []string{categoryID3},
			},
			expected: 2,
		},
		{
			desc: "Should return count of 2 automatic reports with both category and balance ids filter",
			precondition: []model.AutomaticReport{
				{
					ID:          automaticReportID7,
					BalanceIDs:  []string{balanceID6},
					CategoryIDs: []string{categoryID5, categoryID6},
					Name:        "test_count_7",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID8,
					BalanceIDs:  []string{balanceID5},
					CategoryIDs: []string{categoryID5},
					Name:        "test_count_8",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID9,
					BalanceIDs:  []string{balanceID6},
					CategoryIDs: []string{categoryID5, categoryID7},
					Name:        "test_count_9",
					Period:      model.AutomaticReportPeriodDaily,
				},
			},
			args: service.ListAutomaticReportFilter{
				CategoryIDs: []string{categoryID5, categoryID7},
				BalanceIDs:  []string{balanceID6},
			},
			expected: 2,
		},
		{
			desc: "Should return 0 as a count of automatic reports with random ids",
			args: service.ListAutomaticReportFilter{
				CategoryIDs: []string{uuid.NewString()},
				BalanceIDs:  []string{uuid.NewString()},
			},
			expected: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			for _, precondition := range tc.precondition {
				err := automaticReportStore.Create(ctx, &precondition)
				assert.NoError(t, err)

				t.Cleanup(func() {
					err := automaticReportStore.Delete(ctx, precondition.ID)
					assert.NoError(t, err)
				})
			}

			actual, err := automaticReportStore.Count(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestAutomaticReport_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "automatic_report_list")

	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	automaticReportStore := store.NewAutomaticReport(testCaseDB)

	userID := uuid.NewString()
	currencyID := uuid.NewString()

	balanceID1, balanceID2, balanceID3, balanceID4,
		balanceID5, balanceID6, balanceID7, balanceID8,
		balanceID9, balanceID10, balanceID11, balanceID12,
		balanceID13, balanceID14 := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString()

	categoryID1, categoryID2, categoryID3, categoryID4,
		categoryID5, categoryID6, categoryID7, categoryID8,
		categoryID9, categoryID10, categoryID11, categoryID12,
		categoryID13, categoryID14 := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString()

	automaticReportID1, automaticReportID2, automaticReportID3,
		automaticReportID4, automaticReportID5, automaticReportID6,
		automaticReportID7, automaticReportID8, automaticReportID9,
		automaticReportID10, automaticReportID11, automaticReportID12 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &model.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	allBalanceIDs := []string{
		balanceID1, balanceID2, balanceID3, balanceID4,
		balanceID5, balanceID6, balanceID7, balanceID8,
		balanceID9, balanceID10, balanceID11, balanceID12,
		balanceID13, balanceID14,
	}
	for _, balanceID := range allBalanceIDs {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}

	allCategoryIDs := []string{
		categoryID1, categoryID2, categoryID3, categoryID4,
		categoryID5, categoryID6, categoryID7, categoryID8,
		categoryID9, categoryID10, categoryID11, categoryID12,
		categoryID13, categoryID14,
	}
	for _, categoryID := range allCategoryIDs {
		err = categoryStore.Create(ctx, &model.Category{
			ID:     categoryID,
			UserID: userID,
			Title:  "test_category",
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, categoryID := range allCategoryIDs {
			err = categoryStore.Delete(ctx, categoryID)
			require.NoError(t, err)
		}
		for _, balanceID := range allBalanceIDs {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc         string
		precondition []model.AutomaticReport
		args         service.ListAutomaticReportFilter
		expected     []model.AutomaticReport
	}{
		{
			desc: "Should return list of 2 automatic reports with balance id filter",
			precondition: []model.AutomaticReport{
				{
					ID:          automaticReportID1,
					BalanceIDs:  []string{balanceID1},
					CategoryIDs: []string{categoryID1},
					Name:        "test_list_1",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID2,
					BalanceIDs:  []string{balanceID2},
					CategoryIDs: []string{categoryID2},
					Name:        "test_list_2",
					Period:      model.AutomaticReportPeriodWeekly,
				},
				{
					ID:          automaticReportID3,
					BalanceIDs:  []string{balanceID1, balanceID2},
					CategoryIDs: []string{categoryID1, categoryID2},
					Name:        "test_list_3",
					Period:      model.AutomaticReportPeriodMonthly,
				},
			},
			args: service.ListAutomaticReportFilter{
				BalanceIDs: []string{balanceID1},
			},
			expected: []model.AutomaticReport{
				{
					ID:          automaticReportID3,
					BalanceIDs:  []string{balanceID1, balanceID2},
					CategoryIDs: []string{categoryID1, categoryID2},
					Name:        "test_list_3",
					Period:      model.AutomaticReportPeriodMonthly,
				},
				{
					ID:          automaticReportID1,
					BalanceIDs:  []string{balanceID1},
					CategoryIDs: []string{categoryID1},
					Name:        "test_list_1",
					Period:      model.AutomaticReportPeriodDaily,
				},
			},
		},
		{
			desc: "Should return list of 2 automatic reports with category id filter",
			precondition: []model.AutomaticReport{
				{
					ID:          automaticReportID4,
					BalanceIDs:  []string{balanceID3},
					CategoryIDs: []string{categoryID3},
					Name:        "test_list_4",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID5,
					BalanceIDs:  []string{balanceID4},
					CategoryIDs: []string{categoryID4},
					Name:        "test_list_5",
					Period:      model.AutomaticReportPeriodWeekly,
				},
				{
					ID:          automaticReportID6,
					BalanceIDs:  []string{balanceID5},
					CategoryIDs: []string{categoryID3, categoryID5},
					Name:        "test_list_6",
					Period:      model.AutomaticReportPeriodMonthly,
				},
			},
			args: service.ListAutomaticReportFilter{
				CategoryIDs: []string{categoryID3},
			},
			expected: []model.AutomaticReport{
				{
					ID:          automaticReportID6,
					BalanceIDs:  []string{balanceID5},
					CategoryIDs: []string{categoryID3, categoryID5},
					Name:        "test_list_6",
					Period:      model.AutomaticReportPeriodMonthly,
				},
				{
					ID:          automaticReportID4,
					BalanceIDs:  []string{balanceID3},
					CategoryIDs: []string{categoryID3},
					Name:        "test_list_4",
					Period:      model.AutomaticReportPeriodDaily,
				},
			},
		},
		{
			desc: "Should return paginated list with limit 2",
			precondition: []model.AutomaticReport{
				{
					ID:          automaticReportID7,
					BalanceIDs:  []string{balanceID6},
					CategoryIDs: []string{categoryID6},
					Name:        "test_pagination_1",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID8,
					BalanceIDs:  []string{balanceID7},
					CategoryIDs: []string{categoryID7},
					Name:        "test_pagination_2",
					Period:      model.AutomaticReportPeriodWeekly,
				},
				{
					ID:          automaticReportID9,
					BalanceIDs:  []string{balanceID8},
					CategoryIDs: []string{categoryID8},
					Name:        "test_pagination_3",
					Period:      model.AutomaticReportPeriodMonthly,
				},
			},
			args: service.ListAutomaticReportFilter{
				BalanceIDs: []string{balanceID6, balanceID7, balanceID8},
				Pagination: &service.Pagination{
					Page:  1,
					Limit: 2,
				},
			},
			expected: []model.AutomaticReport{
				{
					ID:          automaticReportID9,
					BalanceIDs:  []string{balanceID8},
					CategoryIDs: []string{categoryID8},
					Name:        "test_pagination_3",
					Period:      model.AutomaticReportPeriodMonthly,
				},
				{
					ID:          automaticReportID8,
					BalanceIDs:  []string{balanceID7},
					CategoryIDs: []string{categoryID7},
					Name:        "test_pagination_2",
					Period:      model.AutomaticReportPeriodWeekly,
				},
			},
		},
		{
			desc: "Should return second page with limit 2",
			precondition: []model.AutomaticReport{
				{
					ID:          automaticReportID10,
					BalanceIDs:  []string{balanceID9},
					CategoryIDs: []string{categoryID9},
					Name:        "test_pagination_page2_1",
					Period:      model.AutomaticReportPeriodDaily,
				},
				{
					ID:          automaticReportID11,
					BalanceIDs:  []string{balanceID10},
					CategoryIDs: []string{categoryID10},
					Name:        "test_pagination_page2_2",
					Period:      model.AutomaticReportPeriodWeekly,
				},
				{
					ID:          automaticReportID12,
					BalanceIDs:  []string{balanceID11},
					CategoryIDs: []string{categoryID11},
					Name:        "test_pagination_page2_3",
					Period:      model.AutomaticReportPeriodMonthly,
				},
			},
			args: service.ListAutomaticReportFilter{
				BalanceIDs: []string{balanceID9, balanceID10, balanceID11},
				Pagination: &service.Pagination{
					Page:  2,
					Limit: 2,
				},
			},
			expected: []model.AutomaticReport{
				{
					ID:          automaticReportID10,
					BalanceIDs:  []string{balanceID9},
					CategoryIDs: []string{categoryID9},
					Name:        "test_pagination_page2_1",
					Period:      model.AutomaticReportPeriodDaily,
				},
			},
		},
		{
			desc: "Should return empty list as result due to random filters",
			args: service.ListAutomaticReportFilter{
				BalanceIDs:  []string{uuid.NewString()},
				CategoryIDs: []string{uuid.NewString()},
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			for _, precondition := range tc.precondition {
				err := automaticReportStore.Create(ctx, &precondition)
				assert.NoError(t, err)
			}
			t.Cleanup(func() {
				for _, precondition := range tc.precondition {
					err := automaticReportStore.Delete(ctx, precondition.ID)
					assert.NoError(t, err)
				}
			})

			tc.args.OrderByCreatedAtDesc = true // Add order to simplify assertion
			actual, err := automaticReportStore.List(ctx, tc.args)
			assert.NoError(t, err)

			if tc.expected == nil {
				assert.Empty(t, actual)
				return
			}

			assert.Equal(t, len(tc.expected), len(actual))
			for i := range actual {
				assert.Equal(t, tc.expected[i].ID, actual[i].ID)
				assert.ElementsMatch(t, tc.expected[i].BalanceIDs, actual[i].BalanceIDs)
				assert.ElementsMatch(t, tc.expected[i].CategoryIDs, actual[i].CategoryIDs)
				assert.Equal(t, tc.expected[i].Name, actual[i].Name)
				assert.Equal(t, tc.expected[i].Period, actual[i].Period)
			}
		})
	}
}

func TestAutomaticReport_ListScheduledReportExecutions(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "automatic_report_list_scheduled_report_executions")

	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	automaticReportStore := store.NewAutomaticReport(testCaseDB)

	userID := uuid.NewString()
	currencyID := uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	categoryID1, categoryID2 := uuid.NewString(), uuid.NewString()
	automaticReportID1, automaticReportID2, automaticReportID3,
		automaticReportID4 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString()
	scheduledReportExecutionID1, scheduledReportExecutionID2, scheduledReportExecutionID3,
		scheduledReportExecutionID4, scheduledReportExecutionID5, scheduledReportExecutionID6,
		scheduledReportExecutionID7, scheduledReportExecutionID8, scheduledReportExecutionID9,
		scheduledReportExecutionID10 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &model.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range []string{balanceID1, balanceID2} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}

	for _, categoryID := range []string{categoryID1, categoryID2} {
		err = categoryStore.Create(ctx, &model.Category{
			ID:     categoryID,
			UserID: userID,
			Title:  "test_category",
		})
		require.NoError(t, err)
	}

	for _, automaticReportID := range []string{
		automaticReportID1, automaticReportID2, automaticReportID3,
		automaticReportID4,
	} {
		err = automaticReportStore.Create(ctx, &model.AutomaticReport{
			ID:          automaticReportID,
			BalanceIDs:  []string{balanceID1},
			CategoryIDs: []string{categoryID1},
			Name:        "test_" + automaticReportID,
			Period:      model.AutomaticReportPeriodDaily,
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, automaticReportID := range []string{
			automaticReportID1, automaticReportID2, automaticReportID3,
			automaticReportID4,
		} {
			err := automaticReportStore.Delete(ctx, automaticReportID)
			require.NoError(t, err)
		}
		for _, categoryID := range []string{categoryID1, categoryID2} {
			err := categoryStore.Delete(ctx, categoryID)
			require.NoError(t, err)
		}
		for _, balanceID := range []string{balanceID1, balanceID2} {
			err := balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	now := time.Now().UTC()

	testCases := [...]struct {
		desc         string
		precondition []model.ScheduledReportExecution
		args         service.ListScheduledReportExecutionFilter
		expected     []model.ScheduledReportExecution
	}{
		{
			desc: "Should return all scheduled executions for specific automatic report",
			precondition: []model.ScheduledReportExecution{
				{
					ID:                scheduledReportExecutionID1,
					AutomaticReportID: automaticReportID1,
					ExecutionDate:     now.Add(1 * time.Hour),
				},
				{
					ID:                scheduledReportExecutionID2,
					AutomaticReportID: automaticReportID1,
					ExecutionDate:     now.Add(2 * time.Hour),
				},
				{
					ID:                scheduledReportExecutionID3,
					AutomaticReportID: automaticReportID2,
					ExecutionDate:     now.Add(3 * time.Hour),
				},
			},
			args: service.ListScheduledReportExecutionFilter{
				AutomaticReportID: automaticReportID1,
			},
			expected: []model.ScheduledReportExecution{
				{
					ID:                scheduledReportExecutionID1,
					AutomaticReportID: automaticReportID1,
					ExecutionDate:     now.Add(1 * time.Hour),
				},
				{
					ID:                scheduledReportExecutionID2,
					AutomaticReportID: automaticReportID1,
					ExecutionDate:     now.Add(2 * time.Hour),
				},
			},
		},
		{
			desc: "Should return scheduled executions filtered by date range",
			precondition: []model.ScheduledReportExecution{
				{
					ID:                scheduledReportExecutionID4,
					AutomaticReportID: automaticReportID3,
					ExecutionDate:     now.Add(5 * time.Hour),
				},
				{
					ID:                scheduledReportExecutionID5,
					AutomaticReportID: automaticReportID3,
					ExecutionDate:     now.Add(10 * time.Hour),
				},
				{
					ID:                scheduledReportExecutionID6,
					AutomaticReportID: automaticReportID3,
					ExecutionDate:     now.Add(20 * time.Hour),
				},
			},
			args: service.ListScheduledReportExecutionFilter{
				BetweenFilter: &service.BetweenFilter{
					From: now.Add(8 * time.Hour),
					To:   now.Add(15 * time.Hour),
				},
			},
			expected: []model.ScheduledReportExecution{
				{
					ID:                scheduledReportExecutionID5,
					AutomaticReportID: automaticReportID3,
					ExecutionDate:     now.Add(10 * time.Hour),
				},
			},
		},
		{
			desc: "Should return scheduled executions with both automatic report ID and date range filters",
			precondition: []model.ScheduledReportExecution{
				{
					ID:                scheduledReportExecutionID7,
					AutomaticReportID: automaticReportID4,
					ExecutionDate:     now.Add(25 * time.Hour),
				},
				{
					ID:                scheduledReportExecutionID8,
					AutomaticReportID: automaticReportID4,
					ExecutionDate:     now.Add(30 * time.Hour),
				},
				{
					ID:                scheduledReportExecutionID9,
					AutomaticReportID: automaticReportID4,
					ExecutionDate:     now.Add(35 * time.Hour),
				},
				{
					ID:                scheduledReportExecutionID10,
					AutomaticReportID: automaticReportID4,
					ExecutionDate:     now.Add(40 * time.Hour),
				},
			},
			args: service.ListScheduledReportExecutionFilter{
				AutomaticReportID: automaticReportID4,
				BetweenFilter: &service.BetweenFilter{
					From: now.Add(28 * time.Hour),
					To:   now.Add(38 * time.Hour),
				},
			},
			expected: []model.ScheduledReportExecution{
				{
					ID:                scheduledReportExecutionID8,
					AutomaticReportID: automaticReportID4,
					ExecutionDate:     now.Add(30 * time.Hour),
				},
				{
					ID:                scheduledReportExecutionID9,
					AutomaticReportID: automaticReportID4,
					ExecutionDate:     now.Add(35 * time.Hour),
				},
			},
		},
		{
			desc: "Should return empty list when no executions exist",
			args: service.ListScheduledReportExecutionFilter{
				AutomaticReportID: uuid.NewString(),
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			for _, precondition := range tc.precondition {
				err := automaticReportStore.CreateScheduledReportExecution(ctx, &precondition)
				assert.NoError(t, err)
			}
			t.Cleanup(func() {
				for _, precondition := range tc.precondition {
					err := automaticReportStore.DeleteScheduledReportExecution(ctx, precondition.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := automaticReportStore.ListScheduledReportExecutions(ctx, tc.args)
			assert.NoError(t, err)

			if len(tc.expected) == 0 {
				assert.Empty(t, actual)
				return
			}

			assert.Equal(t, len(tc.expected), len(actual))
			assert.Equal(t, len(tc.expected), len(actual))
			for i := range actual {
				assert.Equal(t, tc.expected[i].ID, actual[i].ID)
				assert.Equal(t, tc.expected[i].AutomaticReportID, actual[i].AutomaticReportID)
				assert.Equal(t, tc.expected[i].ExecutionDate.UTC(), actual[i].ExecutionDate.UTC())
			}
		})
	}
}

func TestAutomaticReport_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "automatic_report_update")

	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	automaticReportStore := store.NewAutomaticReport(testCaseDB)

	userID := uuid.NewString()
	currencyID := uuid.NewString()
	balanceID1, balanceID2, balanceID3,
		balanceID4, balanceID5 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString()
	categoryID1, categoryID2, categoryID3,
		categoryID4, categoryID5 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString()
	automaticReportID1, automaticReportID2,
		automaticReportID3, automaticReportID4 := uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &model.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	allBalanceIDs := []string{balanceID1, balanceID2, balanceID3, balanceID4, balanceID5}
	for _, balanceID := range allBalanceIDs {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}

	allCategoryIDs := []string{categoryID1, categoryID2, categoryID3, categoryID4, categoryID5}
	for _, categoryID := range allCategoryIDs {
		err = categoryStore.Create(ctx, &model.Category{
			ID:     categoryID,
			UserID: userID,
			Title:  "test_category",
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, categoryID := range allCategoryIDs {
			err = categoryStore.Delete(ctx, categoryID)
			require.NoError(t, err)
		}
		for _, balanceID := range allBalanceIDs {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc         string
		precondition *model.AutomaticReport
		args         *model.AutomaticReport
		expected     *model.AutomaticReport
	}{
		{
			desc: "Should update automatic report name and period",
			precondition: &model.AutomaticReport{
				ID:          automaticReportID1,
				BalanceIDs:  []string{balanceID1},
				CategoryIDs: []string{categoryID1},
				Name:        "test_update_original_1",
				Period:      model.AutomaticReportPeriodDaily,
			},
			args: &model.AutomaticReport{
				ID:          automaticReportID1,
				BalanceIDs:  []string{balanceID1},
				CategoryIDs: []string{categoryID1},
				Name:        "test_update_modified_1",
				Period:      model.AutomaticReportPeriodWeekly,
			},
			expected: &model.AutomaticReport{
				ID:          automaticReportID1,
				BalanceIDs:  []string{balanceID1},
				CategoryIDs: []string{categoryID1},
				Name:        "test_update_modified_1",
				Period:      model.AutomaticReportPeriodWeekly,
			},
		},
		{
			desc: "Should update automatic report with new balance IDs",
			precondition: &model.AutomaticReport{
				ID:          automaticReportID2,
				BalanceIDs:  []string{balanceID2},
				CategoryIDs: []string{categoryID2},
				Name:        "test_update_original_2",
				Period:      model.AutomaticReportPeriodDaily,
			},
			args: &model.AutomaticReport{
				ID:          automaticReportID2,
				BalanceIDs:  []string{balanceID2, balanceID3},
				CategoryIDs: []string{categoryID2},
				Name:        "test_update_original_2",
				Period:      model.AutomaticReportPeriodDaily,
			},
			expected: &model.AutomaticReport{
				ID:          automaticReportID2,
				BalanceIDs:  []string{balanceID2, balanceID3},
				CategoryIDs: []string{categoryID2},
				Name:        "test_update_original_2",
				Period:      model.AutomaticReportPeriodDaily,
			},
		},
		{
			desc: "Should update automatic report with new category IDs",
			precondition: &model.AutomaticReport{
				ID:          automaticReportID3,
				BalanceIDs:  []string{balanceID4},
				CategoryIDs: []string{categoryID3},
				Name:        "test_update_original_3",
				Period:      model.AutomaticReportPeriodMonthly,
			},
			args: &model.AutomaticReport{
				ID:          automaticReportID3,
				BalanceIDs:  []string{balanceID4},
				CategoryIDs: []string{categoryID3, categoryID4},
				Name:        "test_update_original_3",
				Period:      model.AutomaticReportPeriodMonthly,
			},
			expected: &model.AutomaticReport{
				ID:          automaticReportID3,
				BalanceIDs:  []string{balanceID4},
				CategoryIDs: []string{categoryID3, categoryID4},
				Name:        "test_update_original_3",
				Period:      model.AutomaticReportPeriodMonthly,
			},
		},
		{
			desc: "Should update automatic report with completely different balance and category IDs",
			precondition: &model.AutomaticReport{
				ID:          automaticReportID4,
				BalanceIDs:  []string{balanceID1, balanceID2},
				CategoryIDs: []string{categoryID1, categoryID2},
				Name:        "test_update_original_4",
				Period:      model.AutomaticReportPeriodDaily,
			},
			args: &model.AutomaticReport{
				ID:          automaticReportID4,
				BalanceIDs:  []string{balanceID4, balanceID5},
				CategoryIDs: []string{categoryID4, categoryID5},
				Name:        "test_update_modified_4",
				Period:      model.AutomaticReportPeriodQuarterly,
			},
			expected: &model.AutomaticReport{
				ID:          automaticReportID4,
				BalanceIDs:  []string{balanceID4, balanceID5},
				CategoryIDs: []string{categoryID4, categoryID5},
				Name:        "test_update_modified_4",
				Period:      model.AutomaticReportPeriodQuarterly,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			err := automaticReportStore.Create(ctx, tc.precondition)
			assert.NoError(t, err)
			t.Cleanup(func() {
				err := automaticReportStore.Delete(ctx, tc.precondition.ID)
				assert.NoError(t, err)
			})

			err = automaticReportStore.Update(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := automaticReportStore.Get(ctx, service.GetAutomaticReportFilter{ID: tc.expected.ID})
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.ElementsMatch(t, tc.expected.BalanceIDs, actual.BalanceIDs)
			assert.ElementsMatch(t, tc.expected.CategoryIDs, actual.CategoryIDs)
			assert.Equal(t, tc.expected.Name, actual.Name)
			assert.Equal(t, tc.expected.Period, actual.Period)
		})
	}
}

func TestAutomaticReport_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "automatic_report_delete")

	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	automaticReportStore := store.NewAutomaticReport(testCaseDB)

	userID := uuid.NewString()
	currencyID := uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	categoryID1, categoryID2 := uuid.NewString(), uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &model.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range []string{balanceID1, balanceID2} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currencyID,
		})
		assert.NoError(t, err)
	}

	for _, categoryID := range []string{categoryID1, categoryID2} {
		err = categoryStore.Create(ctx, &model.Category{
			ID:     categoryID,
			UserID: userID,
			Title:  "test_category",
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, categoryID := range []string{categoryID1, categoryID2} {
			err = categoryStore.Delete(ctx, categoryID)
			require.NoError(t, err)
		}
		for _, balanceID := range []string{balanceID1, balanceID2} {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err := deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc         string
		precondition *model.AutomaticReport
		args         string
	}{
		{
			desc: "Should delete existing automatic report",
			precondition: &model.AutomaticReport{
				ID:          uuid.NewString(),
				BalanceIDs:  []string{balanceID1},
				CategoryIDs: []string{categoryID1},
				Name:        "test_delete_1",
				Period:      model.AutomaticReportPeriodDaily,
			},
		},
		{
			desc: "Should delete automatic report with multiple balances and categories",
			precondition: &model.AutomaticReport{
				ID:          uuid.NewString(),
				BalanceIDs:  []string{balanceID1, balanceID2},
				CategoryIDs: []string{categoryID1, categoryID2},
				Name:        "test_delete_2",
				Period:      model.AutomaticReportPeriodWeekly,
			},
		},
		{
			desc:         "Should not fail when deleting non-existent automatic report",
			precondition: nil,
			args:         uuid.NewString(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			reportID := tc.args
			if tc.precondition != nil {
				reportID = tc.precondition.ID
				err := automaticReportStore.Create(ctx, tc.precondition)
				assert.NoError(t, err)
			}

			err := automaticReportStore.Delete(ctx, reportID)
			assert.NoError(t, err)

			actual, err := automaticReportStore.Get(ctx, service.GetAutomaticReportFilter{ID: reportID})
			assert.NoError(t, err)
			assert.Nil(t, actual)
		})
	}
}

func TestAutomaticReport_DeleteScheduledReportExecution(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "automatic_report_delete_scheduled_report_execution")

	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	automaticReportStore := store.NewAutomaticReport(testCaseDB)

	userID := uuid.NewString()
	currencyID := uuid.NewString()
	balanceID := uuid.NewString()
	categoryID := uuid.NewString()
	automaticReportID := uuid.NewString()

	err := currencyStore.CreateIfNotExists(ctx, &model.Currency{
		ID:   currencyID,
		Code: "USD",
	})
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	err = balanceStore.Create(ctx, &model.Balance{
		ID:         balanceID,
		UserID:     userID,
		CurrencyID: currencyID,
	})
	require.NoError(t, err)

	err = categoryStore.Create(ctx, &model.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	err = automaticReportStore.Create(ctx, &model.AutomaticReport{
		ID:          automaticReportID,
		BalanceIDs:  []string{balanceID},
		CategoryIDs: []string{categoryID},
		Name:        "test_report",
		Period:      model.AutomaticReportPeriodDaily,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := automaticReportStore.Delete(ctx, automaticReportID)
		require.NoError(t, err)
		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err = balanceStore.Delete(ctx, balanceID)
		require.NoError(t, err)
		err = deleteCurrencyByID(testCaseDB.DB, currencyID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	now := time.Now().UTC()

	testCases := [...]struct {
		desc         string
		precondition *model.ScheduledReportExecution
		args         string
	}{
		{
			desc: "Should delete existing scheduled report execution",
			precondition: &model.ScheduledReportExecution{
				ID:                uuid.NewString(),
				AutomaticReportID: automaticReportID,
				ExecutionDate:     now.Add(24 * time.Hour),
			},
		},
		{
			desc: "Should delete scheduled report execution with different date",
			precondition: &model.ScheduledReportExecution{
				ID:                uuid.NewString(),
				AutomaticReportID: automaticReportID,
				ExecutionDate:     now.Add(48 * time.Hour),
			},
		},
		{
			desc:         "Should not fail when deleting non-existent scheduled report execution",
			precondition: nil,
			args:         uuid.NewString(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			executionID := tc.args
			if tc.precondition != nil {
				executionID = tc.precondition.ID
				err := automaticReportStore.CreateScheduledReportExecution(ctx, tc.precondition)
				assert.NoError(t, err)
			}

			err := automaticReportStore.DeleteScheduledReportExecution(ctx, executionID)
			assert.NoError(t, err)

			var actual model.ScheduledReportExecution
			err = testCaseDB.DB.Get(&actual, "SELECT * FROM scheduled_report_executions WHERE id = $1", executionID)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, sql.ErrNoRows))
		})
	}
}
