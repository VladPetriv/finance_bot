package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCurrency_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, "currency_create_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.DB.Drop(ctx)
		assert.NoError(t, err)
	})

	currencyStore := store.NewCurrency(db)

	testCases := [...]struct {
		desc                string
		preconditions       *models.Currency
		args                *models.Currency
		createIsNotExpected bool
	}{
		{
			desc: "positive: currency created",
			args: &models.Currency{
				ID:     uuid.NewString(),
				Name:   "US Dollar",
				Code:   "US",
				Symbol: "$",
			},
		},
		{
			desc: "positive: currency with args symbol already exists, new currency won't be created",
			preconditions: &models.Currency{
				ID:     uuid.NewString(),
				Name:   "US Dollar",
				Code:   "US",
				Symbol: "$",
			},
			args: &models.Currency{
				ID:     uuid.NewString(),
				Name:   "CAD Dollar",
				Code:   "US",
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
				err = currencyStore.CreateIfNotExists(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				_, err := currencyStore.DB.Collection("Currencies").DeleteOne(ctx, bson.M{"_id": tc.args.ID})
				assert.NoError(t, err)

				if tc.preconditions != nil {
					_, err := currencyStore.DB.Collection("Currencies").DeleteOne(ctx, bson.M{"_id": tc.preconditions.ID})
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
			err = db.DB.Collection("Currencies").FindOne(ctx, bson.M{"_id": currencyToCompareID}).Decode(&createdCurrency)
			assert.NoError(t, err)
			assert.Equal(t, currencyToCompare, createdCurrency)
		})
	}
}

func TestCurrency_Count(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, "currency_count_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.DB.Drop(ctx)
		assert.NoError(t, err)
	})

	currencyStore := store.NewCurrency(db)

	testCases := [...]struct {
		desc          string
		preconditions []models.Currency
		expected      int
	}{
		{
			desc:     "positive: count currencies when collection is empty",
			expected: 0,
		},
		{
			desc: "positive: count currencies when collection has items",
			preconditions: []models.Currency{
				{
					ID:     uuid.NewString(),
					Name:   "US Dollar",
					Code:   "USD",
					Symbol: "$",
				},
				{
					ID:     uuid.NewString(),
					Name:   "Euro",
					Code:   "EUR",
					Symbol: "€",
				},
			},
			expected: 2,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if len(tc.preconditions) > 0 {
				for _, currency := range tc.preconditions {
					err := currencyStore.CreateIfNotExists(ctx, &currency)
					assert.NoError(t, err)
				}
			}

			t.Cleanup(func() {
				if len(tc.preconditions) > 0 {
					for _, currency := range tc.preconditions {
						_, err := currencyStore.DB.Collection("Currencies").DeleteOne(ctx, bson.M{"_id": currency.ID})
						assert.NoError(t, err)
					}
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
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, "currency_list_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.DB.Drop(ctx)
		assert.NoError(t, err)
	})

	currencyStore := store.NewCurrency(db)

	testCases := [...]struct {
		desc          string
		preconditions []models.Currency
		expected      []models.Currency
	}{
		{
			desc:     "positive: empty list when no currencies",
			expected: []models.Currency{},
		},
		{
			desc: "positive: list all currencies",
			preconditions: []models.Currency{
				{
					ID:     uuid.NewString(),
					Name:   "US Dollar",
					Symbol: "$",
					Code:   "USD",
				},
				{
					ID:     uuid.NewString(),
					Name:   "Euro",
					Symbol: "€",
					Code:   "EUR",
				},
			},
			expected: []models.Currency{
				{
					ID:     uuid.NewString(),
					Name:   "US Dollar",
					Symbol: "$",
					Code:   "USD",
				},
				{
					ID:     uuid.NewString(),
					Name:   "Euro",
					Symbol: "€",
					Code:   "EUR",
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if len(tc.preconditions) > 0 {
				for _, currency := range tc.preconditions {
					err := currencyStore.CreateIfNotExists(ctx, &currency)
					assert.NoError(t, err)
				}
			}

			t.Cleanup(func() {
				for _, currency := range tc.preconditions {
					_, err := currencyStore.DB.Collection("Currencies").DeleteOne(ctx, bson.M{"_id": currency.ID})
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
			}
		})
	}
}

func TestCurrency_Exists(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, "currency_exists_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.DB.Drop(ctx)
		assert.NoError(t, err)
	})

	currencyStore := store.NewCurrency(db)

	currencyID := uuid.NewString()

	testCases := [...]struct {
		desc          string
		preconditions *models.Currency
		args          service.ExistsCurrencyFilter
		expected      bool
	}{
		{
			desc: "negative: currency by id exists",
			preconditions: &models.Currency{
				ID:     currencyID,
				Name:   "US Dollar",
				Symbol: "$",
				Code:   "USD",
			},
			expected: true,
		},
		{
			desc: "negative: currency by id doesn't exists",
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
					_, err := currencyStore.DB.Collection("Currencies").DeleteOne(ctx, bson.M{"_id": tc.preconditions.ID})
					assert.NoError(t, err)
				}
			})

			actual, err := currencyStore.Exists(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
