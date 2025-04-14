package models_test

import (
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCalculateScheduledOperationBillingDates(t *testing.T) {
	t.Parallel()

	type args struct {
		period    models.SubscriptionPeriod
		startDate time.Time
		maxDates  int
	}

	testCases := [...]struct {
		desc     string
		args     args
		expected []time.Time
	}{
		{
			desc: "Should receive 1 weekly billing dates",
			args: args{
				period:    models.SubscriptionPeriodWeekly,
				startDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				maxDates:  1,
			},
			expected: []time.Time{
				time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "Should receive 2 weekly billing dates",
			args: args{
				period:    models.SubscriptionPeriodWeekly,
				startDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				maxDates:  2,
			},
			expected: []time.Time{
				time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2023, 1, 8, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "Should receive 1 monthly billing dates",
			args: args{
				period:    models.SubscriptionPeriodMonthly,
				startDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				maxDates:  1,
			},
			expected: []time.Time{
				time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "Should receive 2 monthly billing dates",
			args: args{
				period:    models.SubscriptionPeriodMonthly,
				startDate: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
				maxDates:  2,
			},
			expected: []time.Time{
				time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
				time.Date(2023, 2, 15, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "Should receive 1 yearly billing date",
			args: args{
				period:    models.SubscriptionPeriodYearly,
				startDate: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
				maxDates:  1,
			},
			expected: []time.Time{
				time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "Should receive 2 yearly billing date",
			args: args{
				period:    models.SubscriptionPeriodYearly,
				startDate: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
				maxDates:  2,
			},
			expected: []time.Time{
				time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual := models.CalculateScheduledOperationBillingDates(tc.args.period, tc.args.startDate, tc.args.maxDates)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
