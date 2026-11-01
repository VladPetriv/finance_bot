package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	amount300 = money.NewFromInt(300)
	amount400 = money.NewFromInt(400)
)

func TestBalance_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_create")
	balanceStore := store.NewBalance(testCaseDB)
	userStore := store.NewUser(testCaseDB)
	currencyStore := store.NewCurrency(testCaseDB)

	balanceID, userID := uuid.NewString(), uuid.NewString()
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

	t.Cleanup(func() {
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc                 string
		preconditions        *model.Balance
		args                 *model.Balance
		expectDuplicateError bool
	}{
		{
			desc: "balance created",
			args: &model.Balance{
				ID:         uuid.NewString(),
				UserID:     userID,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
		},
		{
			desc: "duplicate key error because balance already exists",
			preconditions: &model.Balance{
				ID:         balanceID,
				UserID:     userID,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: &model.Balance{
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
			assert.True(t, tc.args.Amount.Equal(actual.Amount))
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
	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	for _, userID := range [...]string{userID1, userID2, userID3} {
		err := userStore.Create(ctx, &model.User{
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
		preconditions *model.Balance
		args          service.GetBalanceFilter
		expected      *model.Balance
	}{
		{
			desc: "balance received by user id",
			preconditions: &model.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: service.GetBalanceFilter{
				UserID: userID1,
			},
			expected: &model.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
		},
		{
			desc: "balance received by id with currency preload",
			preconditions: &model.Balance{
				ID:         balanceID2,
				UserID:     userID2,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: service.GetBalanceFilter{
				BalanceID:       balanceID2,
				PreloadCurrency: true,
			},
			expected: &model.Balance{
				ID:         balanceID2,
				UserID:     userID2,
				CurrencyID: currency.ID,
				Amount:     amount300,
				Currency:   *currency,
			},
		},
		{
			desc: "balance received by name",
			preconditions: &model.Balance{
				ID:         balanceID3,
				Name:       "test_x3",
				UserID:     userID3,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: service.GetBalanceFilter{
				Name: "test_x3",
			},
			expected: &model.Balance{
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

			if tc.expected == nil {
				assert.Nil(t, actual)
				return
			}
			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.Equal(t, tc.expected.UserID, actual.UserID)
			assert.Equal(t, tc.expected.CurrencyID, actual.CurrencyID)
			assert.Equal(t, tc.expected.Name, actual.Name)
			assert.True(t, tc.expected.Amount.Equal(actual.Amount))

			if tc.args.PreloadCurrency {
				assert.Equal(t, tc.expected.Currency.ID, actual.Currency.ID)
				assert.Equal(t, tc.expected.Currency.Name, actual.Currency.Name)
				assert.Equal(t, tc.expected.Currency.Code, actual.Currency.Code)
				assert.Equal(t, tc.expected.Currency.Symbol, actual.Currency.Symbol)
			}
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
	currency := &model.Currency{
		ID:   uuid.NewString(),
		Code: "USD",
	}

	err := currencyStore.CreateIfNotExists(ctx, currency)
	require.NoError(t, err)

	for _, userID := range [...]string{userID1, userID2} {
		err := userStore.Create(ctx, &model.User{
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
		preconditions *model.Balance
		args          *model.Balance
		expected      *model.Balance
	}{
		{
			desc: "balance updated",
			preconditions: &model.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: &model.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount400,
			},
			expected: &model.Balance{
				ID:         balanceID1,
				UserID:     userID1,
				CurrencyID: currency.ID,
				Amount:     amount400,
			},
		},
		{
			desc: "balance not updated because of not existed id",
			preconditions: &model.Balance{
				ID:         balanceID2,
				UserID:     userID2,
				CurrencyID: currency.ID,
				Amount:     amount300,
			},
			args: &model.Balance{
				ID:         uuid.NewString(),
				UserID:     userID2,
				CurrencyID: currency.ID,
				Amount:     amount400,
			},
			expected: &model.Balance{
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
			assert.Equal(t, tc.expected.ID, actual.ID)
			assert.Equal(t, tc.expected.UserID, actual.UserID)
			assert.Equal(t, tc.expected.CurrencyID, actual.CurrencyID)
			assert.Equal(t, tc.expected.Name, actual.Name)
			assert.True(t, tc.expected.Amount.Equal(actual.Amount))
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

	t.Cleanup(func() {
		err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
		require.NoError(t, err)
		err = deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions *model.Balance
		args          string
	}{
		{
			desc: "balance deleted",
			preconditions: &model.Balance{
				ID:         balanceID,
				UserID:     userID,
				CurrencyID: currency.ID,
			},
			args: balanceID,
		},
		{
			desc: "balance not deleted because of not existed id",
			preconditions: &model.Balance{
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

func TestBalance_AmountMinorFallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_amount_minor_fallback")
	balanceStore := store.NewBalance(testCaseDB)

	userID := uuid.NewString()
	err := store.NewUser(testCaseDB).Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	type precondition struct {
		legacyAmount *string
		amountMinor  *int64
	}

	type expected struct {
		amount money.Money
	}

	legacy := "12.34"
	legacyOther := "99.99"
	minor := int64(5600)

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should return legacy amount converted to minor units when amount_minor is NULL",
			precondition: &precondition{
				legacyAmount: &legacy,
			},
			expected: &expected{
				amount: money.NewFromMinorUnits(1234),
			},
		},
		{
			desc: "it should return amount_minor when legacy amount is NULL",
			precondition: &precondition{
				amountMinor: &minor,
			},
			expected: &expected{
				amount: money.NewFromMinorUnits(5600),
			},
		},
		{
			desc: "it should prefer amount_minor when both columns are set",
			precondition: &precondition{
				legacyAmount: &legacyOther,
				amountMinor:  &minor,
			},
			expected: &expected{
				amount: money.NewFromMinorUnits(5600),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			balanceID := uuid.NewString()
			_, err := testCaseDB.DB.ExecContext(
				ctx,
				"INSERT INTO balances (id, user_id, currency_id, name, amount, amount_minor) VALUES ($1, $4, '', '', $2, $3);",
				balanceID, tc.precondition.legacyAmount, tc.precondition.amountMinor, userID,
			)
			require.NoError(t, err)

			actual, err := balanceStore.Get(ctx, service.GetBalanceFilter{BalanceID: balanceID})
			require.NoError(t, err)
			require.NotNil(t, actual)
			assert.True(t, tc.expected.amount.Equal(actual.Amount))
		})
	}
}

func TestBalance_WritesOnlyAmountMinor(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "balance_writes_only_amount_minor")
	balanceStore := store.NewBalance(testCaseDB)

	userID := uuid.NewString()
	err := store.NewUser(testCaseDB).Create(ctx, &model.User{
		ID:       userID,
		Username: "test" + userID,
	})
	require.NoError(t, err)

	balanceID := uuid.NewString()
	err = balanceStore.Create(ctx, &model.Balance{
		ID:     balanceID,
		UserID: userID,
		Amount: money.NewFromMinorUnits(1050),
	})
	require.NoError(t, err)

	type row struct {
		Legacy *string `db:"amount"`
		Minor  *int64  `db:"amount_minor"`
	}

	var created row
	err = testCaseDB.DB.GetContext(ctx, &created, "SELECT amount, amount_minor FROM balances WHERE id = $1;", balanceID)
	require.NoError(t, err)
	assert.Nil(t, created.Legacy)
	require.NotNil(t, created.Minor)
	assert.Equal(t, int64(1050), *created.Minor)

	err = balanceStore.Update(ctx, &model.Balance{
		ID:     balanceID,
		UserID: userID,
		Amount: money.NewFromMinorUnits(2000),
	})
	require.NoError(t, err)

	var updated row
	err = testCaseDB.DB.GetContext(ctx, &updated, "SELECT amount, amount_minor FROM balances WHERE id = $1;", balanceID)
	require.NoError(t, err)
	assert.Nil(t, updated.Legacy)
	require.NotNil(t, updated.Minor)
	assert.Equal(t, int64(2000), *updated.Minor)
}
