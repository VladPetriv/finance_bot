package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/models"
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

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)

	currencyStore := store.NewCurrency(db)

	testCases := [...]struct {
		desc                 string
		preconditions        *models.Currency
		input                *models.Currency
		expectDuplicateError bool
	}{
		{
			desc: "positive: currency created",
			input: &models.Currency{
				ID:     uuid.NewString(),
				Name:   "US Dollar",
				Symbol: "$",
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = currencyStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				_, err := currencyStore.DB.Collection("Currencies").DeleteOne(ctx, bson.M{"_id": tc.input.ID})
				assert.NoError(t, err)
			})

			err := currencyStore.Create(ctx, tc.input)
			assert.NoError(t, err)
		})
	}
}

func TestCurrency_Count(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo
	cfg := config.Get()

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
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
					err = currencyStore.Create(ctx, &currency)
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

	db, err := database.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	require.NoError(t, err)
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
				},
				{
					ID:     uuid.NewString(),
					Name:   "Euro",
					Symbol: "€",
				},
			},
			expected: []models.Currency{
				{
					ID:     uuid.NewString(),
					Name:   "US Dollar",
					Symbol: "$",
				},
				{
					ID:     uuid.NewString(),
					Name:   "Euro",
					Symbol: "€",
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
					err = currencyStore.Create(ctx, &currency)
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
