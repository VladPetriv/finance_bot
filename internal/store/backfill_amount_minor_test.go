package store_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/VladPetriv/finance_bot/internal/migrations"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackfillAmountMinor(t *testing.T) {
	t.Parallel()

	ctx := context.Background() //nolint: forbidigo

	type expected struct {
		amountMinor int64
	}

	testCases := [...]struct {
		desc     string
		legacy   string
		expected *expected
	}{
		{
			desc:   "it should convert a two decimals legacy amount",
			legacy: "12.34",
			expected: &expected{
				amountMinor: 1234,
			},
		},
		{
			desc:   "it should round a sub-cent positive amount half away from zero",
			legacy: "10.555",
			expected: &expected{
				amountMinor: 1056,
			},
		},
		{
			desc:   "it should round a sub-cent negative amount half away from zero",
			legacy: "-10.555",
			expected: &expected{
				amountMinor: -1056,
			},
		},
		{
			desc:   "it should convert zero",
			legacy: "0",
			expected: &expected{
				amountMinor: 0,
			},
		},
	}
	for i, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			testDB := createTestDBWithMigrations(
				t,
				fmt.Sprintf("backfill_amount_minor_%d", i),
				migrations.Migrations[:len(migrations.Migrations)-2],
			)

			seed := [...]string{
				"INSERT INTO users (id, username) VALUES ('u', 'u')",
				"INSERT INTO categories (id, user_id, title) VALUES ('c', 'u', 'c')",
				"INSERT INTO balances (id, user_id, currency_id, name, amount) VALUES ('b', 'u', '', '', $1)",
				"INSERT INTO operations (id, category_id, balance_id, type, description, amount) VALUES ('o', 'c', 'b', 'incoming', '', $1)",
				"INSERT INTO balance_subscriptions (id, balance_id, category_id, name, amount, period, start_at) VALUES ('s', 'b', 'c', '', $1, 'weekly', NOW())",
			}
			for _, query := range seed {
				args := []any{}
				if strings.Contains(query, "$1") {
					args = append(args, tc.legacy)
				}

				_, err := testDB.DB.ExecContext(ctx, query, args...)
				require.NoError(t, err)
			}

			expectedMoney := money.NewFromMinorUnits(tc.expected.amountMinor)

			for run := 0; run < 2; run++ {
				err := migrations.MigrateDB(migrations.MigrateDBOptions{
					DB:         testDB.DB,
					DBName:     fmt.Sprintf("backfill_amount_minor_%d", i),
					Migrations: migrations.Migrations,
				})
				require.NoError(t, err)

				for _, table := range []string{"balances", "operations", "balance_subscriptions"} {
					var actual int64
					err = testDB.DB.QueryRowContext(ctx, "SELECT amount_minor FROM "+table).Scan(&actual)
					require.NoError(t, err)
					assert.True(t, expectedMoney.Equal(money.NewFromMinorUnits(actual)), table)
				}
			}
		})
	}
}
