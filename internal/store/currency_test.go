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
)

func TestCurrency_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "currency_create")
	currencyStore := store.NewCurrency(testCaseDB)

	testCases := [...]struct {
		desc                string
		preconditions       *models.Currency
		args                *models.Currency
		createIsNotExpected bool
	}{
		{
			desc: "created currency",
			args: &models.Currency{
				ID:     uuid.NewString(),
				Name:   "US Dollar",
				Code:   "US_test_create_1",
				Symbol: "$",
			},
		},
		{
			desc: "currency with args symbol already exists, new currency won't be created",
			preconditions: &models.Currency{
				ID:     uuid.NewString(),
				Name:   "US Dollar",
				Code:   "US_test_create_2",
				Symbol: "$",
			},
			args: &models.Currency{
				ID:     uuid.NewString(),
				Name:   "CAD Dollar",
				Code:   "US_test_create_2",
				Symbol: "$",
			},
			createIsNotExpected: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := currencyStore.CreateIfNotExists(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				err := deleteCurrencyByID(testCaseDB.DB, tc.args.ID)
				assert.NoError(t, err)

				if tc.preconditions != nil {
					err := deleteCurrencyByID(testCaseDB.DB, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			err := currencyStore.CreateIfNotExists(ctx, tc.args)
			assert.NoError(t, err)

			var (
				currencyToCompareID string
				currencyToCompare   models.Currency
			)
			switch tc.createIsNotExpected {
			case true:
				currencyToCompareID = tc.preconditions.ID
				currencyToCompare = *tc.preconditions
			case false:
				currencyToCompareID = tc.args.ID
				currencyToCompare = *tc.args
			}

			var createdCurrency models.Currency
			err = testCaseDB.DB.Get(&createdCurrency, "SELECT * FROM currencies WHERE id=$1;", currencyToCompareID)
			assert.NoError(t, err)
			assert.Equal(t, currencyToCompare, createdCurrency)
		})
	}
}

func TestCurrency_Count(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "currency_count")
	currencyStore := store.NewCurrency(testCaseDB)

	testCases := [...]struct {
		desc          string
		preconditions []models.Currency
		expected      int
	}{
		{
			desc: "count currencies when table has 2 items",
			preconditions: []models.Currency{
				{
					ID:     uuid.NewString(),
					Name:   "US Dollar",
					Code:   "USD_test_count_2",
					Symbol: "$",
				},
				{
					ID:     uuid.NewString(),
					Name:   "Euro",
					Code:   "EUR_test_count_2",
					Symbol: "€",
				},
			},
			expected: 2,
		},
		{
			desc:     "count currencies when table is empty",
			expected: 0,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			for _, currency := range tc.preconditions {
				err := currencyStore.CreateIfNotExists(ctx, &currency)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				for _, currency := range tc.preconditions {
					err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
					assert.NoError(t, err)
				}
			})

			count, err := currencyStore.Count(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, count)
		})
	}
}

func TestCurrency_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "currency_list")
	currencyStore := store.NewCurrency(testCaseDB)

	testCases := [...]struct {
		desc          string
		preconditions []models.Currency
		expected      []models.Currency
	}{
		{
			desc: "list all currencies",
			preconditions: []models.Currency{
				{
					ID:     uuid.NewString(),
					Name:   "US Dollar",
					Symbol: "$",
					Code:   "USD_test_list_1",
				},
				{
					ID:     uuid.NewString(),
					Name:   "Euro",
					Symbol: "€",
					Code:   "EUR_test_list_1",
				},
			},
			expected: []models.Currency{
				{
					ID:     uuid.NewString(),
					Name:   "US Dollar",
					Symbol: "$",
					Code:   "USD_test_list_1",
				},
				{
					ID:     uuid.NewString(),
					Name:   "Euro",
					Symbol: "€",
					Code:   "EUR_test_list_1",
				},
			},
		},
		{
			desc:     "positive: empty list when no currencies",
			expected: []models.Currency{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			for _, currency := range tc.preconditions {
				err := currencyStore.CreateIfNotExists(ctx, &currency)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				for _, currency := range tc.preconditions {
					err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
					assert.NoError(t, err)
				}
			})

			currencies, err := currencyStore.List(ctx)
			assert.NoError(t, err)

			if len(tc.preconditions) == 0 {
				assert.Empty(t, currencies)
				return
			}

			for i, currency := range currencies {
				assert.Equal(t, tc.preconditions[i].Name, currency.Name)
				assert.Equal(t, tc.preconditions[i].Symbol, currency.Symbol)
				assert.Equal(t, tc.preconditions[i].Code, currency.Code)
			}
		})
	}
}

func TestCurrency_Exists(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	testCaseDB := createTestDB(t, "currency_exists")
	currencyStore := store.NewCurrency(testCaseDB)

	currencyID := uuid.NewString()

	testCases := [...]struct {
		desc          string
		preconditions *models.Currency
		args          service.ExistsCurrencyFilter
		expected      bool
	}{
		{
			desc: "currency by id exists",
			preconditions: &models.Currency{
				ID:     currencyID,
				Name:   "US Dollar",
				Symbol: "$",
				Code:   "USD_test_exists_1",
			},
			args: service.ExistsCurrencyFilter{
				ID: currencyID,
			},
			expected: true,
		},
		{
			desc: "currency by id doesn't exists",
			args: service.ExistsCurrencyFilter{
				ID: uuid.NewString(),
			},
			expected: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := currencyStore.CreateIfNotExists(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err := deleteCurrencyByID(testCaseDB.DB, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			actual, err := currencyStore.Exists(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func deleteCurrencyByID(db *sqlx.DB, currencyID string) error {
	_, err := db.Exec("DELETE FROM currencies WHERE id = $1;", currencyID)
	return err
}
