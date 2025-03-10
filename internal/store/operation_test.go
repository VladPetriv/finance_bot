package store_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperation_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "operation_create")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	operationStore := store.NewOperation(testCaseDB)

	userID := uuid.NewString()
	balanceID := uuid.NewString()
	categoryID := uuid.NewString()
	operationID1, operationID2 := uuid.NewString(), uuid.NewString()
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
		preconditions        *models.Operation
		args                 *models.Operation
		expectDuplicateError bool
	}{
		{
			desc: "operation created",
			args: &models.Operation{
				ID:          operationID1,
				CategoryID:  categoryID,
				BalanceID:   balanceID,
				Type:        models.OperationTypeIncoming,
				Amount:      "100",
				Description: "test_create_1",
			},
		},
		{
			desc: "operation not created because already exist",
			preconditions: &models.Operation{
				ID:          operationID2,
				CategoryID:  categoryID,
				BalanceID:   balanceID,
				Type:        models.OperationTypeIncoming,
				Amount:      "100",
				Description: "test_create_2",
			},
			args: &models.Operation{
				ID:          operationID2,
				CategoryID:  categoryID,
				BalanceID:   balanceID,
				Type:        models.OperationTypeIncoming,
				Amount:      "100",
				Description: "test_create_2",
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = operationStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err = operationStore.Delete(ctx, tc.args.ID)
					assert.NoError(t, err)
				}
				if tc.args != nil {
					err = operationStore.Delete(ctx, tc.args.ID)
					assert.NoError(t, err)
				}
			})

			err := operationStore.Create(ctx, tc.args)
			if tc.expectDuplicateError {
				assert.True(t, isDuplicateKeyError(err))
				return
			}

			assert.NoError(t, err)

			actual, err := operationStore.Get(ctx, service.GetOperationFilter{ID: tc.args.ID})
			assert.NoError(t, err)
			assert.Equal(t, tc.args.ID, actual.ID)
			assert.Equal(t, tc.args.CategoryID, actual.CategoryID)
			assert.Equal(t, tc.args.BalanceID, actual.BalanceID)
			assert.Equal(t, tc.args.Type, actual.Type)
			assert.Equal(t, tc.args.Amount, actual.Amount)
			assert.Equal(t, tc.args.Description, actual.Description)
		})
	}
}

func TestOperation_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "operation_get")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	operationStore := store.NewOperation(testCaseDB)

	userID := uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	categoryID := uuid.NewString()
	operationID1,
		operationID2,
		operationID3,
		operationID4,
		operationID5 := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
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

	for _, balanceID := range [...]string{balanceID1, balanceID2} {
		err = balanceStore.Create(ctx, &models.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currency.ID,
		})
		require.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &models.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, balanceID := range [...]string{balanceID1, balanceID2} {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	now := time.Now()

	testCases := [...]struct {
		desc          string
		preconditions *models.Operation
		args          service.GetOperationFilter
		expected      *models.Operation
	}{
		{
			desc: "found operation by id",
			preconditions: &models.Operation{
				ID:          operationID1,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        models.OperationTypeIncoming,
				Amount:      "100",
				Description: "test_get_1",
			},
			args: service.GetOperationFilter{
				ID: operationID1,
			},
			expected: &models.Operation{
				ID:          operationID1,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        models.OperationTypeIncoming,
				Amount:      "100",
				Description: "test_get_1",
			},
		},
		{
			desc: "found operation by type",
			preconditions: &models.Operation{
				ID:          operationID2,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        models.OperationTypeSpending,
				Amount:      "100",
				Description: "test_get_2",
			},
			args: service.GetOperationFilter{
				Type: models.OperationTypeSpending,
			},
			expected: &models.Operation{
				ID:          operationID2,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        models.OperationTypeSpending,
				Amount:      "100",
				Description: "test_get_2",
			},
		},
		{
			desc: "found operation by createdAtFrom and createdAtTo",
			preconditions: &models.Operation{
				ID:          operationID3,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        models.OperationTypeTransfer,
				Amount:      "100",
				Description: "test_get_3",
				CreatedAt:   now.Add(-3 * time.Hour),
			},
			args: service.GetOperationFilter{
				CreateAtFrom: now.Add(-4 * time.Hour),
				CreateAtTo:   now.Add(-1 * time.Hour),
			},
			expected: &models.Operation{
				ID:          operationID3,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        models.OperationTypeTransfer,
				Amount:      "100",
				Description: "test_get_3",
				CreatedAt:   now.Add(-3 * time.Hour),
			},
		},
		{
			desc: "found operation by balances ids filter",
			preconditions: &models.Operation{
				ID:          operationID4,
				CategoryID:  categoryID,
				BalanceID:   balanceID2,
				Type:        models.OperationTypeTransfer,
				Amount:      "100",
				Description: "test_get_4",
			},
			args: service.GetOperationFilter{
				BalanceIDs: []string{balanceID2},
			},
			expected: &models.Operation{
				ID:          operationID4,
				CategoryID:  categoryID,
				BalanceID:   balanceID2,
				Type:        models.OperationTypeTransfer,
				Amount:      "100",
				Description: "test_get_4",
			},
		},
		{
			desc: "found operation by amount filter",
			preconditions: &models.Operation{
				ID:          operationID5,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        models.OperationTypeTransfer,
				Amount:      "50",
				Description: "test_get_5",
			},
			args: service.GetOperationFilter{
				Amount: "50",
			},
			expected: &models.Operation{
				ID:          operationID5,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        models.OperationTypeTransfer,
				Amount:      "50",
				Description: "test_get_5",
			},
		},
		{
			desc: "operation not found",
			args: service.GetOperationFilter{
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
				err := operationStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err := operationStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := operationStore.Get(ctx, tc.args)
			assert.NoError(t, err)

			if tc.expected == nil {
				assert.Nil(t, actual)
				return
			}

			assert.NotNil(t, actual)
			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.Equal(t, tc.expected.BalanceID, actual.BalanceID)
			assert.Equal(t, tc.expected.CategoryID, actual.CategoryID)
			assert.Equal(t, tc.expected.Type, actual.Type)
			assert.Equal(t, tc.expected.Amount, actual.Amount)
			assert.Equal(t, tc.expected.Description, actual.Description)
		})
	}
}

func TestOperation_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "operation_list")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	operationStore := store.NewOperation(testCaseDB)

	userID := uuid.NewString()
	balanceID1, balanceID2, balanceID3,
		balanceID4, balanceID5, balanceID6 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString()
	categoryID := uuid.NewString()
	operationID1, operationID2,
		operationID3, operationID4, operationID5,
		operationID6, operationID7, operationID8,
		operationID9, operationID10, operationID11, operationID12,
		operationID13, operationID14, operationID15, operationID16,
		operationID17, operationID18, operationID19, operationID20 := uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()

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

	for _, balanceID := range [...]string{balanceID1, balanceID2, balanceID3, balanceID4, balanceID5, balanceID6} {
		err = balanceStore.Create(ctx, &models.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currency.ID,
		})
		require.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &models.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, balanceID := range [...]string{balanceID1, balanceID2, balanceID3, balanceID4, balanceID5, balanceID6} {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions []models.Operation
		args          service.ListOperationsFilter
		expected      []models.Operation
	}{
		{
			desc: "received all operations by only balance id",
			preconditions: []models.Operation{
				{
					ID:         operationID1,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       models.OperationTypeIncoming,
				},
				{
					ID:         operationID2,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       models.OperationTypeIncoming,
				},
			},
			args: service.ListOperationsFilter{
				BalanceID: balanceID1,
			},
			expected: []models.Operation{
				{
					ID:         operationID1,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       models.OperationTypeIncoming,
				},
				{
					ID:         operationID2,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       models.OperationTypeIncoming,
				},
			},
		},
		{
			desc: "received all operations by day as a creation period",
			preconditions: []models.Operation{
				{
					ID:         operationID3,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-23 * time.Hour),
				},
				{
					ID:         operationID4,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
				{
					ID:         operationID5,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID2,
				CreationPeriod: models.CreationPeriodDay,
			},
			expected: []models.Operation{
				{
					ID:         operationID4,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
				{
					ID:         operationID3,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-23 * time.Hour),
				},
			},
		},
		{
			desc: "received all operations by week as a creation period",
			preconditions: []models.Operation{
				{
					ID:         operationID6,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-168 * time.Hour),
				},
				{
					ID:         operationID7,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-100 * time.Hour),
				},
				{
					ID:         operationID8,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID3,
				CreationPeriod: models.CreationPeriodWeek,
			},
			expected: []models.Operation{
				{
					ID:         operationID7,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-100 * time.Hour),
				},
				{
					ID:         operationID8,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
		},
		{
			desc: "received all operations by month as a creation period",
			preconditions: []models.Operation{
				{
					ID:         operationID9,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-730 * time.Hour),
				},
				{
					ID:         operationID10,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-200 * time.Hour),
				},
				{
					ID:         operationID11,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-300 * time.Hour),
				},
				{
					ID:         operationID12,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID4,
				CreationPeriod: models.CreationPeriodMonth,
			},
			expected: []models.Operation{
				{
					ID:         operationID10,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-200 * time.Hour),
				},
				{
					ID:         operationID11,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-300 * time.Hour),
				},
			},
		},
		{
			desc: "received all operations by year as a creation period",
			preconditions: []models.Operation{
				{
					ID:         operationID13,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-8760 * time.Hour),
				},
				{
					ID:         operationID14,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-3500 * time.Hour),
				},
				{
					ID:         operationID15,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-1000 * time.Hour),
				},
				{
					ID:         operationID16,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID5,
				CreationPeriod: models.CreationPeriodYear,
			},
			expected: []models.Operation{
				{
					ID:         operationID14,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-3500 * time.Hour),
				},
				{
					ID:         operationID15,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-1000 * time.Hour),
				},
			},
		},
		{
			desc: "received all operations where time less than args time",
			preconditions: []models.Operation{
				{
					ID:         operationID17,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
				{
					ID:         operationID18,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-24 * time.Hour),
				},
				{
					ID:         operationID19,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-1 * time.Hour),
				},
				{
					ID:         operationID20,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:         balanceID6,
				CreatedAtLessThan: time.Now().Add(-10 * time.Hour),
			},
			expected: []models.Operation{
				{
					ID:         operationID17,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
				{
					ID:         operationID18,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-24 * time.Hour),
				},
			},
		},
		{
			desc: "negative: operations not found",
			args: service.ListOperationsFilter{
				BalanceID:      uuid.NewString(),
				CreationPeriod: models.CreationPeriodYear,
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				for _, o := range tc.preconditions {
					err := operationStore.Create(ctx, &o)
					require.NoError(t, err)
				}
			}

			t.Cleanup(func() {
				_, err := testCaseDB.DB.Exec("DELETE FROM operations WHERE balance_id = $1;", tc.args.BalanceID)
				assert.NoError(t, err)
			})

			actual, err := operationStore.List(ctx, tc.args)
			assert.NoError(t, err)

			// NOTE: Sort both slices to get right order when compare.
			sort.Slice(actual, func(i, j int) bool {
				return actual[i].CreatedAt.Unix() < actual[j].CreatedAt.Unix()
			})
			sort.Slice(tc.expected, func(i, j int) bool {
				return tc.expected[i].CreatedAt.Unix() < tc.expected[j].CreatedAt.Unix()
			})

			for i := 0; i < len(tc.expected); i++ {
				assert.Equal(t, tc.expected[i].ID, actual[i].ID)
				assert.Equal(t, tc.expected[i].CategoryID, actual[i].CategoryID)
				assert.Equal(t, tc.expected[i].BalanceID, actual[i].BalanceID)
			}
		})
	}
}

func TestOperation_Count(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "operation_count")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	operationStore := store.NewOperation(testCaseDB)

	userID := uuid.NewString()
	balanceID1, balanceID2, balanceID3,
		balanceID4, balanceID5, balanceID6 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString()
	categoryID := uuid.NewString()
	operationID1, operationID2,
		operationID3, operationID4, operationID5,
		operationID6, operationID7, operationID8,
		operationID9, operationID10, operationID11, operationID12,
		operationID13, operationID14, operationID15, operationID16,
		operationID17, operationID18, operationID19, operationID20 := uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()

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

	for _, balanceID := range [...]string{balanceID1, balanceID2, balanceID3, balanceID4, balanceID5, balanceID6} {
		err = balanceStore.Create(ctx, &models.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currency.ID,
		})
		require.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &models.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, balanceID := range [...]string{balanceID1, balanceID2, balanceID3, balanceID4, balanceID5, balanceID6} {
			err = balanceStore.Delete(ctx, balanceID)
			require.NoError(t, err)
		}
		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions []models.Operation
		args          service.ListOperationsFilter
		expected      int
	}{
		{
			desc: "received all operations by only balance id",
			preconditions: []models.Operation{
				{
					ID:         operationID1,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       models.OperationTypeIncoming,
				},
				{
					ID:         operationID2,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       models.OperationTypeIncoming,
				},
			},
			args: service.ListOperationsFilter{
				BalanceID: balanceID1,
			},
			expected: 2,
		},
		{
			desc: "received all operations by day as a creation period",
			preconditions: []models.Operation{
				{
					ID:         operationID3,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-23 * time.Hour),
				},
				{
					ID:         operationID4,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
				{
					ID:         operationID5,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID2,
				CreationPeriod: models.CreationPeriodDay,
			},
			expected: 2,
		},
		{
			desc: "received all operations by week as a creation period",
			preconditions: []models.Operation{
				{
					ID:         operationID6,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-168 * time.Hour),
				},
				{
					ID:         operationID7,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-100 * time.Hour),
				},
				{
					ID:         operationID8,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID3,
				CreationPeriod: models.CreationPeriodWeek,
			},
			expected: 2,
		},
		{
			desc: "received all operations by month as a creation period",
			preconditions: []models.Operation{
				{
					ID:         operationID9,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-730 * time.Hour),
				},
				{
					ID:         operationID10,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-200 * time.Hour),
				},
				{
					ID:         operationID11,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-300 * time.Hour),
				},
				{
					ID:         operationID12,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(1 * time.Second),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID4,
				CreationPeriod: models.CreationPeriodMonth,
			},
			expected: 2,
		},
		{
			desc: "received all operations by year as a creation period",
			preconditions: []models.Operation{
				{
					ID:         operationID13,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-8760 * time.Hour),
				},
				{
					ID:         operationID14,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-3500 * time.Hour),
				},
				{
					ID:         operationID15,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-1000 * time.Hour),
				},
				{
					ID:         operationID16,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID5,
				CreationPeriod: models.CreationPeriodYear,
			},
			expected: 3,
		},
		{
			desc: "received all operations where time less than args time",
			preconditions: []models.Operation{
				{
					ID:         operationID17,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
				{
					ID:         operationID18,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-24 * time.Hour),
				},
				{
					ID:         operationID19,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-1 * time.Hour),
				},
				{
					ID:         operationID20,
					CategoryID: categoryID,
					BalanceID:  balanceID6,
					Type:       models.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:         balanceID6,
				CreatedAtLessThan: time.Now().Add(-10 * time.Hour),
			},
			expected: 2,
		},
		{
			desc: "negative: operations not found",
			args: service.ListOperationsFilter{
				BalanceID:      uuid.NewString(),
				CreationPeriod: models.CreationPeriodYear,
			},
			expected: 0,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				for _, o := range tc.preconditions {
					err := operationStore.Create(ctx, &o)
					require.NoError(t, err)
				}
			}

			t.Cleanup(func() {
				_, err := testCaseDB.DB.Exec("DELETE FROM operations WHERE balance_id = $1;", tc.args.BalanceID)
				assert.NoError(t, err)
			})

			actual, err := operationStore.Count(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestOperation_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "operation_delete")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	operationStore := store.NewOperation(testCaseDB)

	userID := uuid.NewString()
	balanceID := uuid.NewString()
	categoryID := uuid.NewString()
	operationID := uuid.NewString()
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
	testCases := []struct {
		desc          string
		preconditions *models.Operation
		args          string
	}{
		{
			desc: "operation deleted",
			preconditions: &models.Operation{
				ID:         operationID,
				CategoryID: categoryID,
				BalanceID:  balanceID,
				Type:       models.OperationTypeIncoming,
			},
			args: operationID,
		},
		{
			desc: "negatie: operation not deleted because of not existed id",
			preconditions: &models.Operation{
				ID:         uuid.NewString(),
				CategoryID: categoryID,
				BalanceID:  balanceID,
				Type:       models.OperationTypeIncoming,
			},
			args: uuid.NewString(),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := operationStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err := operationStore.Delete(ctx, tc.preconditions.ID)
				assert.NoError(t, err)
			})

			err := operationStore.Delete(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := operationStore.Get(ctx, service.GetOperationFilter{ID: tc.preconditions.ID})
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
