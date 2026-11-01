package store_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/VladPetriv/finance_bot/pkg/typecast"
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
	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	err = balanceStore.Create(ctx, &model.Balance{
		ID:         balanceID,
		UserID:     userID,
		CurrencyID: currency.ID,
	})
	assert.NoError(t, err)

	err = categoryStore.Create(ctx, &model.Category{
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
		preconditions        *model.Operation
		args                 *model.Operation
		expectDuplicateError bool
	}{
		{
			desc: "operation created",
			args: &model.Operation{
				ID:                operationID1,
				CategoryID:        categoryID,
				BalanceID:         balanceID,
				Type:              model.OperationTypeIncoming,
				Amount:            money.NewFromInt(100),
				ParentOperationID: "123",
				ExchangeRate:      "10.00",
				Description:       "test_create_1",
			},
		},
		{
			desc: "operation not created because already exist",
			preconditions: &model.Operation{
				ID:          operationID2,
				CategoryID:  categoryID,
				BalanceID:   balanceID,
				Type:        model.OperationTypeIncoming,
				Amount:      money.NewFromInt(100),
				Description: "test_create_2",
			},
			args: &model.Operation{
				ID:          operationID2,
				CategoryID:  categoryID,
				BalanceID:   balanceID,
				Type:        model.OperationTypeIncoming,
				Amount:      money.NewFromInt(100),
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
			assert.Equal(t, tc.args.ParentOperationID, actual.ParentOperationID)
			assert.Equal(t, tc.args.Type, actual.Type)
			assert.True(t, tc.args.Amount.Equal(actual.Amount))
			assert.Equal(t, tc.args.ExchangeRate, actual.ExchangeRate)
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
	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range [...]string{balanceID1, balanceID2} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currency.ID,
		})
		require.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &model.Category{
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
		preconditions *model.Operation
		args          service.GetOperationFilter
		expected      *model.Operation
	}{
		{
			desc: "found operation by id",
			preconditions: &model.Operation{
				ID:                operationID1,
				CategoryID:        categoryID,
				BalanceID:         balanceID1,
				ParentOperationID: "123",
				Type:              model.OperationTypeIncoming,
				Amount:            money.NewFromInt(100),
				ExchangeRate:      "1.0",
				Description:       "test_get_1",
			},
			args: service.GetOperationFilter{
				ID: operationID1,
			},
			expected: &model.Operation{
				ID:                operationID1,
				CategoryID:        categoryID,
				BalanceID:         balanceID1,
				ParentOperationID: "123",
				Type:              model.OperationTypeIncoming,
				Amount:            money.NewFromInt(100),
				ExchangeRate:      "1.0",
				Description:       "test_get_1",
			},
		},
		{
			desc: "found operation by type",
			preconditions: &model.Operation{
				ID:          operationID2,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        model.OperationTypeSpending,
				Amount:      money.NewFromInt(100),
				Description: "test_get_2",
			},
			args: service.GetOperationFilter{
				Type: model.OperationTypeSpending,
			},
			expected: &model.Operation{
				ID:          operationID2,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        model.OperationTypeSpending,
				Amount:      money.NewFromInt(100),
				Description: "test_get_2",
			},
		},
		{
			desc: "found operation by createdAtFrom and createdAtTo",
			preconditions: &model.Operation{
				ID:          operationID3,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        model.OperationTypeTransfer,
				Amount:      money.NewFromInt(100),
				Description: "test_get_3",
				CreatedAt:   now.Add(-3 * time.Hour),
			},
			args: service.GetOperationFilter{
				CreateAtFrom: now.Add(-4 * time.Hour),
				CreateAtTo:   now.Add(-1 * time.Hour),
			},
			expected: &model.Operation{
				ID:          operationID3,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        model.OperationTypeTransfer,
				Amount:      money.NewFromInt(100),
				Description: "test_get_3",
				CreatedAt:   now.Add(-3 * time.Hour),
			},
		},
		{
			desc: "found operation by balances ids filter",
			preconditions: &model.Operation{
				ID:          operationID4,
				CategoryID:  categoryID,
				BalanceID:   balanceID2,
				Type:        model.OperationTypeTransfer,
				Amount:      money.NewFromInt(100),
				Description: "test_get_4",
			},
			args: service.GetOperationFilter{
				BalanceIDs: []string{balanceID2},
			},
			expected: &model.Operation{
				ID:          operationID4,
				CategoryID:  categoryID,
				BalanceID:   balanceID2,
				Type:        model.OperationTypeTransfer,
				Amount:      money.NewFromInt(100),
				Description: "test_get_4",
			},
		},
		{
			desc: "found operation by amount filter",
			preconditions: &model.Operation{
				ID:          operationID5,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        model.OperationTypeTransfer,
				Amount:      money.NewFromInt(50),
				Description: "test_get_5",
			},
			args: service.GetOperationFilter{
				Amount: typecast.ToPtr(money.NewFromInt(50)),
			},
			expected: &model.Operation{
				ID:          operationID5,
				CategoryID:  categoryID,
				BalanceID:   balanceID1,
				Type:        model.OperationTypeTransfer,
				Amount:      money.NewFromInt(50),
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
			assert.Equal(t, tc.expected.ParentOperationID, actual.ParentOperationID)
			assert.Equal(t, tc.expected.Type, actual.Type)
			assert.True(t, tc.expected.Amount.Equal(actual.Amount))
			assert.Equal(t, tc.expected.ExchangeRate, actual.ExchangeRate)
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
		balanceID4, balanceID5, balanceID6,
		balanceID7, balanceID8 := uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString()

	categoryID := uuid.NewString()
	operationID1, operationID2,
		operationID3, operationID4, operationID5,
		operationID6, operationID7, operationID8,
		operationID9, operationID10, operationID11, operationID12,
		operationID13, operationID14, operationID15, operationID16,
		operationID17, operationID18, operationID19, operationID20,
		operationID21, operationID22, operationID23, operationID24 := uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()

	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range [...]string{
		balanceID1, balanceID2, balanceID3,
		balanceID4, balanceID5, balanceID6,
		balanceID7, balanceID8,
	} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currency.ID,
		})
		require.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &model.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, balanceID := range [...]string{
			balanceID1, balanceID2, balanceID3,
			balanceID4, balanceID5, balanceID6,
			balanceID7, balanceID8,
		} {
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
		preconditions []model.Operation
		args          service.ListOperationsFilter
		expected      []model.Operation
	}{
		{
			desc: "received all operations by only balance id",
			preconditions: []model.Operation{
				{
					ID:         operationID1,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID2,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
				},
			},
			args: service.ListOperationsFilter{
				BalanceID: balanceID1,
			},
			expected: []model.Operation{
				{
					ID:         operationID1,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID2,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
				},
			},
		},
		{
			desc: "received all operations by day as a creation period",
			preconditions: []model.Operation{
				{
					ID:         operationID3,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-23 * time.Hour),
				},
				{
					ID:         operationID4,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
				{
					ID:         operationID5,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID2,
				CreationPeriod: model.CreationPeriodDay,
			},
			expected: []model.Operation{
				{
					ID:         operationID4,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
				{
					ID:         operationID3,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-23 * time.Hour),
				},
			},
		},
		{
			desc: "received all operations by week as a creation period",
			preconditions: []model.Operation{
				{
					ID:         operationID6,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-168 * time.Hour),
				},
				{
					ID:         operationID7,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-100 * time.Hour),
				},
				{
					ID:         operationID8,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID3,
				CreationPeriod: model.CreationPeriodWeek,
			},
			expected: []model.Operation{
				{
					ID:         operationID7,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-100 * time.Hour),
				},
				{
					ID:         operationID8,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
		},
		{
			desc: "received all operations by month as a creation period",
			preconditions: []model.Operation{
				{
					ID:         operationID9,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-1000 * time.Hour),
				},
				{
					ID:         operationID10,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-200 * time.Hour),
				},
				{
					ID:         operationID11,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-300 * time.Hour),
				},
				{
					ID:         operationID12,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID4,
				CreationPeriod: model.CreationPeriodMonth,
			},
			expected: []model.Operation{
				{
					ID:         operationID10,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-200 * time.Hour),
				},
				{
					ID:         operationID11,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-300 * time.Hour),
				},
			},
		},
		{
			desc: "received all operations by year as a creation period",
			preconditions: []model.Operation{
				{
					ID:         operationID13,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-8760 * time.Hour),
				},
				{
					ID:         operationID14,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-3500 * time.Hour),
				},
				{
					ID:         operationID15,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-1000 * time.Hour),
				},
				{
					ID:         operationID16,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID5,
				CreationPeriod: model.CreationPeriodYear,
			},
			expected: []model.Operation{
				{
					ID:         operationID14,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-3500 * time.Hour),
				},
				{
					ID:         operationID15,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-1000 * time.Hour),
				},
			},
		},
		{
			desc: "received all operations on page 1 with limit of 2, total amount 4",
			preconditions: []model.Operation{
				{
					ID:         operationID17,
					CategoryID: categoryID,
					BalanceID:  balanceID7,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID18,
					CategoryID: categoryID,
					BalanceID:  balanceID7,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID19,
					CategoryID: categoryID,
					BalanceID:  balanceID7,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID20,
					CategoryID: categoryID,
					BalanceID:  balanceID7,
					Type:       model.OperationTypeIncoming,
				},
			},
			args: service.ListOperationsFilter{
				BalanceID: balanceID7,
				Pagination: &service.Pagination{
					Page:  1,
					Limit: 2,
				},
			},
			expected: []model.Operation{
				{
					ID:         operationID17,
					CategoryID: categoryID,
					BalanceID:  balanceID7,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID18,
					CategoryID: categoryID,
					BalanceID:  balanceID7,
					Type:       model.OperationTypeIncoming,
				},
			},
		},
		{
			desc: "received all operations on page 2 with limit of 2, total amount 4",
			preconditions: []model.Operation{
				{
					ID:         operationID21,
					CategoryID: categoryID,
					BalanceID:  balanceID8,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID22,
					CategoryID: categoryID,
					BalanceID:  balanceID8,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID23,
					CategoryID: categoryID,
					BalanceID:  balanceID8,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID24,
					CategoryID: categoryID,
					BalanceID:  balanceID8,
					Type:       model.OperationTypeIncoming,
				},
			},
			args: service.ListOperationsFilter{
				BalanceID: balanceID8,
				Pagination: &service.Pagination{
					Page:  2,
					Limit: 2,
				},
			},
			expected: []model.Operation{
				{
					ID:         operationID23,
					CategoryID: categoryID,
					BalanceID:  balanceID8,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID24,
					CategoryID: categoryID,
					BalanceID:  balanceID8,
					Type:       model.OperationTypeIncoming,
				},
			},
		},
		{
			desc: "negative: operations not found",
			args: service.ListOperationsFilter{
				BalanceID:      uuid.NewString(),
				CreationPeriod: model.CreationPeriodYear,
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

func TestOperation_ListOperationYears(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "operation_list_operation_years")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	operationStore := store.NewOperation(testCaseDB)

	userID := uuid.NewString()
	balanceID1, balanceID2, balanceID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()
	categoryID := uuid.NewString()
	operationID1, operationID2,
		operationID3 := uuid.NewString(), uuid.NewString(),
		uuid.NewString()

	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range [...]string{
		balanceID1, balanceID2, balanceID3,
	} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currency.ID,
		})
		require.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &model.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, balanceID := range [...]string{
			balanceID1, balanceID2, balanceID3,
		} {
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

	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)

	testCases := [...]struct {
		desc          string
		preconditions []model.Operation
		args          service.ListOperationYearsFilter
		expected      []model.Year
	}{
		{
			desc: "received all years of operations by balance id filter",
			preconditions: []model.Operation{
				{
					ID:         operationID1,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Date(2025, time.March, 1, 10, 0, 0, 0, time.UTC),
				},
				{
					ID:         operationID2,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Date(2023, time.March, 1, 10, 0, 0, 0, time.UTC),
				},
				{
					ID:         operationID3,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Date(2024, time.March, 1, 10, 0, 0, 0, time.UTC),
				},
			},
			args: service.ListOperationYearsFilter{
				BalanceID: balanceID1,
			},
			expected: []model.Year{2023, 2024, 2025},
		},
		{
			desc: "operation near the end of the year is counted in the UTC year when no location is provided",
			preconditions: []model.Operation{
				{
					ID:         uuid.NewString(),
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Date(2024, time.December, 31, 23, 30, 0, 0, time.UTC),
				},
			},
			args: service.ListOperationYearsFilter{
				BalanceID: balanceID2,
			},
			expected: []model.Year{2024},
		},
		{
			desc: "operation near the end of the year is counted in the next year for a location ahead of UTC",
			preconditions: []model.Operation{
				{
					ID:         uuid.NewString(),
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Date(2024, time.December, 31, 23, 30, 0, 0, time.UTC),
				},
			},
			args: service.ListOperationYearsFilter{
				BalanceID: balanceID3,
				Location:  auckland,
			},
			expected: []model.Year{2025},
		},
		{
			desc: "negative: operations not found",
			args: service.ListOperationYearsFilter{
				BalanceID: uuid.NewString(),
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			for _, o := range tc.preconditions {
				err := operationStore.Create(ctx, &o)
				require.NoError(t, err)
			}

			t.Cleanup(func() {
				_, err := testCaseDB.DB.Exec("DELETE FROM operations WHERE balance_id = $1;", tc.args.BalanceID)
				assert.NoError(t, err)
			})

			actual, err := operationStore.ListOperationYears(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestOperation_ListOperationMonths(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "operation_list_operation_months")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	operationStore := store.NewOperation(testCaseDB)

	userID := uuid.NewString()
	balanceID1, balanceID2 := uuid.NewString(), uuid.NewString()
	categoryID := uuid.NewString()
	operationID1, operationID2,
		operationID3 := uuid.NewString(), uuid.NewString(),
		uuid.NewString()

	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range [...]string{
		balanceID1, balanceID2,
	} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currency.ID,
		})
		require.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &model.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, balanceID := range [...]string{
			balanceID1, balanceID2,
		} {
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
		preconditions []model.Operation
		args          service.ListOperationMonthsFilter
		expected      []model.Month
	}{
		{
			desc: "received all montsh of operations by balance id filter",
			preconditions: []model.Operation{
				{
					ID:         operationID1,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Date(2024, time.December, 1, 10, 0, 0, 0, time.UTC),
				},
				{
					ID:         operationID2,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
				},
				{
					ID:         operationID3,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Date(2024, time.June, 1, 10, 0, 0, 0, time.UTC),
				},
			},
			args: service.ListOperationMonthsFilter{
				BalanceID: balanceID1,
				Year:      2024,
			},
			expected: []model.Month{model.MonthJanuary, model.MonthJune, model.MonthDecember},
		},
		{
			desc: "negative: operations not found",
			args: service.ListOperationMonthsFilter{
				BalanceID: uuid.NewString(),
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			for _, o := range tc.preconditions {
				err := operationStore.Create(ctx, &o)
				require.NoError(t, err)
			}

			t.Cleanup(func() {
				_, err := testCaseDB.DB.Exec("DELETE FROM operations WHERE balance_id = $1;", tc.args.BalanceID)
				assert.NoError(t, err)
			})

			actual, err := operationStore.ListOperationMonths(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
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
		operationID13, operationID14, operationID15, operationID16 := uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()

	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	for _, balanceID := range [...]string{balanceID1, balanceID2, balanceID3, balanceID4, balanceID5, balanceID6} {
		err = balanceStore.Create(ctx, &model.Balance{
			ID:         balanceID,
			UserID:     userID,
			CurrencyID: currency.ID,
		})
		require.NoError(t, err)
	}

	err = categoryStore.Create(ctx, &model.Category{
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
		preconditions []model.Operation
		args          service.ListOperationsFilter
		expected      int
	}{
		{
			desc: "received all operations by only balance id",
			preconditions: []model.Operation{
				{
					ID:         operationID1,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
				},
				{
					ID:         operationID2,
					CategoryID: categoryID,
					BalanceID:  balanceID1,
					Type:       model.OperationTypeIncoming,
				},
			},
			args: service.ListOperationsFilter{
				BalanceID: balanceID1,
			},
			expected: 2,
		},
		{
			desc: "received all operations by day as a creation period",
			preconditions: []model.Operation{
				{
					ID:         operationID3,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-23 * time.Hour),
				},
				{
					ID:         operationID4,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
				{
					ID:         operationID5,
					CategoryID: categoryID,
					BalanceID:  balanceID2,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID2,
				CreationPeriod: model.CreationPeriodDay,
			},
			expected: 2,
		},
		{
			desc: "received all operations by week as a creation period",
			preconditions: []model.Operation{
				{
					ID:         operationID6,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-168 * time.Hour),
				},
				{
					ID:         operationID7,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-100 * time.Hour),
				},
				{
					ID:         operationID8,
					CategoryID: categoryID,
					BalanceID:  balanceID3,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-48 * time.Hour),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID3,
				CreationPeriod: model.CreationPeriodWeek,
			},
			expected: 2,
		},
		{
			desc: "received all operations by month as a creation period",
			preconditions: []model.Operation{
				{
					ID:         operationID9,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-730 * time.Hour),
				},
				{
					ID:         operationID10,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-200 * time.Hour),
				},
				{
					ID:         operationID11,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-300 * time.Hour),
				},
				{
					ID:         operationID12,
					CategoryID: categoryID,
					BalanceID:  balanceID4,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(1 * time.Hour),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID4,
				CreationPeriod: model.CreationPeriodMonth,
			},
			expected: 2,
		},
		{
			desc: "received all operations by year as a creation period",
			preconditions: []model.Operation{
				{
					ID:         operationID13,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-8760 * time.Hour),
				},
				{
					ID:         operationID14,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-3500 * time.Hour),
				},
				{
					ID:         operationID15,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now().Add(-1000 * time.Hour),
				},
				{
					ID:         operationID16,
					CategoryID: categoryID,
					BalanceID:  balanceID5,
					Type:       model.OperationTypeIncoming,
					CreatedAt:  time.Now(),
				},
			},
			args: service.ListOperationsFilter{
				BalanceID:      balanceID5,
				CreationPeriod: model.CreationPeriodYear,
			},
			expected: 3,
		},
		{
			desc: "negative: operations not found",
			args: service.ListOperationsFilter{
				BalanceID:      uuid.NewString(),
				CreationPeriod: model.CreationPeriodYear,
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
	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	err = balanceStore.Create(ctx, &model.Balance{
		ID:         balanceID,
		UserID:     userID,
		CurrencyID: currency.ID,
	})
	assert.NoError(t, err)

	err = categoryStore.Create(ctx, &model.Category{
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
		preconditions *model.Operation
		args          string
	}{
		{
			desc: "operation deleted",
			preconditions: &model.Operation{
				ID:         operationID,
				CategoryID: categoryID,
				BalanceID:  balanceID,
				Type:       model.OperationTypeIncoming,
			},
			args: operationID,
		},
		{
			desc: "negatie: operation not deleted because of not existed id",
			preconditions: &model.Operation{
				ID:         uuid.NewString(),
				CategoryID: categoryID,
				BalanceID:  balanceID,
				Type:       model.OperationTypeIncoming,
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

func TestOperation_List_BetweenFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	testCaseDB := createTestDB(t, "operation_list_between_filter")
	currencyStore := store.NewCurrency(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	balanceStore := store.NewBalance(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)
	operationStore := store.NewOperation(testCaseDB)

	userID, balanceID, categoryID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	insideOperationID, boundaryOperationID, outsideOperationID := uuid.NewString(), uuid.NewString(), uuid.NewString()

	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}
	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	err = userStore.Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	err = balanceStore.Create(ctx, &model.Balance{
		ID:         balanceID,
		UserID:     userID,
		CurrencyID: currency.ID,
	})
	require.NoError(t, err)

	err = categoryStore.Create(ctx, &model.Category{
		ID:     categoryID,
		UserID: userID,
		Title:  "test_category",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		for _, operationID := range [...]string{insideOperationID, boundaryOperationID, outsideOperationID} {
			err := operationStore.Delete(ctx, operationID)
			require.NoError(t, err)
		}
		err := balanceStore.Delete(ctx, balanceID)
		require.NoError(t, err)
		err = categoryStore.Delete(ctx, categoryID)
		require.NoError(t, err)
		err = deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	for operationID, createdAt := range map[string]time.Time{
		insideOperationID:   from.Add(time.Hour),
		boundaryOperationID: to,
		outsideOperationID:  from.Add(-time.Hour),
	} {
		err = operationStore.Create(ctx, &model.Operation{
			ID:         operationID,
			BalanceID:  balanceID,
			CategoryID: categoryID,
			Type:       model.OperationTypeSpending,
			Amount:     money.NewFromInt(10),
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		})
		require.NoError(t, err)
	}

	testCases := [...]struct {
		desc     string
		filter   service.ListOperationsFilter
		expected []string
	}{
		{
			desc: "it should return only operations in [from, to) range",
			filter: service.ListOperationsFilter{
				BalanceID: balanceID,
				BetweenFilter: &service.BetweenFilter{
					From: from,
					To:   to,
				},
			},
			expected: []string{insideOperationID},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := operationStore.List(ctx, tc.filter)
			require.NoError(t, err)

			actualIDs := make([]string, 0, len(actual))
			for _, operation := range actual {
				actualIDs = append(actualIDs, operation.ID)
			}
			assert.ElementsMatch(t, tc.expected, actualIDs)
		})
	}
}

func TestOperation_StoresAmountMinor(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "operation_writes_only_amount_minor")
	operationStore := store.NewOperation(testCaseDB)

	balanceID := insertTestBalance(t, testCaseDB)
	operationID := uuid.NewString()

	err := operationStore.Create(ctx, &model.Operation{
		ID:        operationID,
		BalanceID: balanceID,
		Type:      model.OperationTypeIncoming,
		Amount:    money.NewFromMinorUnits(1050),
	})
	require.NoError(t, err)

	type row struct {
		Minor *int64 `db:"amount_minor"`
	}

	var created row
	err = testCaseDB.DB.GetContext(ctx, &created, "SELECT amount_minor FROM operations WHERE id = $1;", operationID)
	require.NoError(t, err)
	require.NotNil(t, created.Minor)
	assert.Equal(t, int64(1050), *created.Minor)

	err = operationStore.Update(ctx, operationID, &model.Operation{
		BalanceID: balanceID,
		Type:      model.OperationTypeIncoming,
		Amount:    money.NewFromMinorUnits(2000),
	})
	require.NoError(t, err)

	var updated row
	err = testCaseDB.DB.GetContext(ctx, &updated, "SELECT amount_minor FROM operations WHERE id = $1;", operationID)
	require.NoError(t, err)
	require.NotNil(t, updated.Minor)
	assert.Equal(t, int64(2000), *updated.Minor)
}

func TestOperation_SearchByAmount(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "operation_search_by_amount")
	operationStore := store.NewOperation(testCaseDB)

	balanceID := insertTestBalance(t, testCaseDB)

	ten := insertTestOperation(t, testCaseDB, balanceID, 1000)
	twenty := insertTestOperation(t, testCaseDB, balanceID, 2050)
	thirty := insertTestOperation(t, testCaseDB, balanceID, 3000)

	type precondition struct {
		filter service.ListOperationsFilter
	}

	type expected struct {
		ids []string
	}

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should find rows by exact amount via from and to",
			precondition: &precondition{
				filter: service.ListOperationsFilter{
					AmountFrom: typecast.ToPtr(money.NewFromInt(10)),
					AmountTo:   typecast.ToPtr(money.NewFromInt(10)),
				},
			},
			expected: &expected{
				ids: []string{ten},
			},
		},
		{
			desc: "it should filter by from",
			precondition: &precondition{
				filter: service.ListOperationsFilter{
					AmountFrom: typecast.ToPtr(money.NewFromInt(20)),
				},
			},
			expected: &expected{
				ids: []string{twenty, thirty},
			},
		},
		{
			desc: "it should filter by to",
			precondition: &precondition{
				filter: service.ListOperationsFilter{
					AmountTo: typecast.ToPtr(money.NewFromInt(15)),
				},
			},
			expected: &expected{
				ids: []string{ten},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			filter := tc.precondition.filter
			filter.BalanceID = balanceID

			listed, err := operationStore.List(ctx, filter)
			require.NoError(t, err)

			actualIDs := make([]string, 0, len(listed))
			for _, operation := range listed {
				actualIDs = append(actualIDs, operation.ID)
			}
			assert.ElementsMatch(t, tc.expected.ids, actualIDs)

			count, err := operationStore.Count(ctx, filter)
			require.NoError(t, err)
			assert.Equal(t, len(tc.expected.ids), count)
		})
	}

	t.Run("it should get an operation by exact amount", func(t *testing.T) {
		t.Parallel()

		got, err := operationStore.Get(ctx, service.GetOperationFilter{
			Amount:     typecast.ToPtr(money.NewFromInt(30)),
			BalanceIDs: []string{balanceID},
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, thirty, got.ID)
	})
}

func insertTestBalance(t *testing.T, db *database.PostgreSQL) string {
	t.Helper()

	userID := uuid.NewString()
	err := store.NewUser(db).Create(context.Background(), &model.User{ //nolint: forbidigo
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	balanceID := uuid.NewString()
	_, err = db.DB.Exec("INSERT INTO balances (id, user_id, currency_id, name, amount_minor) VALUES ($1, $2, '', '', 0);", balanceID, userID)
	require.NoError(t, err)

	return balanceID
}

func insertTestOperation(t *testing.T, db *database.PostgreSQL, balanceID string, amountMinor int64) string {
	t.Helper()

	operationID := uuid.NewString()
	_, err := db.DB.Exec(
		"INSERT INTO operations (id, category_id, balance_id, parent_operation_id, type, description, exchange_rate, amount_minor) VALUES ($1, '', $2, '', 'incoming', '', '', $3);",
		operationID, balanceID, amountMinor,
	)
	require.NoError(t, err)

	return operationID
}
