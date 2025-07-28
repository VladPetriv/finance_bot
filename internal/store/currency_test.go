package store_test

import (
	"context"
	"fmt"
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
		preconditions func() []models.Currency
		args          service.ListCurrenciesFilter
		expected      []models.Currency
	}{
		{
			desc: "list all currencies",
			preconditions: func() []models.Currency {
				return []models.Currency{
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
				}
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
			desc: "list currencies with pagination: total:10, page:1, limit:5",
			preconditions: func() []models.Currency {
				currencies := make([]models.Currency, 0, 10)
				for i := range 10 {
					currencies = append(currencies, models.Currency{
						ID:     uuid.NewString(),
						Name:   fmt.Sprintf("Currency %d", i),
						Symbol: fmt.Sprintf("$%d", i),
						Code:   fmt.Sprintf("C%d", i),
					})
				}
				return currencies
			},
			args: service.ListCurrenciesFilter{
				Pagination: &service.Pagination{
					Page:  1,
					Limit: 5,
				},
			},
			expected: []models.Currency{
				{
					Name:   "Currency 0",
					Symbol: "$0",
					Code:   "C0",
				},
				{
					Name:   "Currency 1",
					Symbol: "$1",
					Code:   "C1",
				},
				{
					Name:   "Currency 2",
					Symbol: "$2",
					Code:   "C2",
				},
				{
					Name:   "Currency 3",
					Symbol: "$3",
					Code:   "C3",
				},
				{
					Name:   "Currency 4",
					Symbol: "$4",
					Code:   "C4",
				},
			},
		},
		{
			desc: "list currencies with pagination: total:10, page 2, limit 5",
			preconditions: func() []models.Currency {
				currencies := make([]models.Currency, 0, 10)
				for i := range 10 {
					currencies = append(currencies, models.Currency{
						ID:     uuid.NewString(),
						Name:   fmt.Sprintf("tc2_Currency %d", i),
						Symbol: fmt.Sprintf("tc2_$%d", i),
						Code:   fmt.Sprintf("tc2_C%d", i),
					})
				}
				return currencies
			},
			args: service.ListCurrenciesFilter{
				Pagination: &service.Pagination{
					Page:  2,
					Limit: 5,
				},
			},
			expected: []models.Currency{
				{
					Name:   "tc2_Currency 5",
					Symbol: "tc2_$5",
					Code:   "tc2_C5",
				},
				{
					Name:   "tc2_Currency 6",
					Symbol: "tc2_$6",
					Code:   "tc2_C6",
				},
				{
					Name:   "tc2_Currency 7",
					Symbol: "tc2_$7",
					Code:   "tc2_C7",
				},
				{
					Name:   "tc2_Currency 8",
					Symbol: "tc2_$8",
					Code:   "tc2_C8",
				},
				{
					Name:   "tc2_Currency 9",
					Symbol: "tc2_$9",
					Code:   "tc2_C9",
				},
			},
		},
		{
			desc: "positive: empty list when no currencies",
			preconditions: func() []models.Currency {
				return []models.Currency{}
			},
			expected: []models.Currency{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			preconditions := tc.preconditions()

			for _, currency := range preconditions {
				err := currencyStore.CreateIfNotExists(ctx, &currency)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				for _, currency := range preconditions {
					err := deleteCurrencyByID(testCaseDB.DB, currency.ID)
					assert.NoError(t, err)
				}
			})

			currencies, err := currencyStore.List(ctx, tc.args)
			assert.NoError(t, err)

			if len(preconditions) == 0 {
				assert.Empty(t, currencies)
				return
			}

			for i, currency := range currencies {
				assert.Equal(t, tc.expected[i].Name, currency.Name)
				assert.Equal(t, tc.expected[i].Symbol, currency.Symbol)
				assert.Equal(t, tc.expected[i].Code, currency.Code)
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
