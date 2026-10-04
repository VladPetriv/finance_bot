package model_test

import (
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestAutomaticReportPeriod_CalculateNextExecutionDate(t *testing.T) {
	t.Parallel()

	type precondition struct {
		period model.AutomaticReportPeriod
		from   time.Time
		now    time.Time
	}

	type expected struct {
		nextExecutionDate time.Time
	}

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should add one day for daily period",
			precondition: &precondition{
				period: model.AutomaticReportPeriodDaily,
				from:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				now:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: &expected{
				nextExecutionDate: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should add seven days for weekly period",
			precondition: &precondition{
				period: model.AutomaticReportPeriodWeekly,
				from:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				now:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: &expected{
				nextExecutionDate: time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should add one month for monthly period",
			precondition: &precondition{
				period: model.AutomaticReportPeriodMonthly,
				from:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				now:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: &expected{
				nextExecutionDate: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should add three months for quarterly period",
			precondition: &precondition{
				period: model.AutomaticReportPeriodQuarterly,
				from:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				now:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: &expected{
				nextExecutionDate: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should add one year for yearly period",
			precondition: &precondition{
				period: model.AutomaticReportPeriodYearly,
				from:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				now:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: &expected{
				nextExecutionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should skip periods that are already in the past",
			precondition: &precondition{
				period: model.AutomaticReportPeriodDaily,
				from:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				now:    time.Date(2025, 1, 5, 12, 0, 0, 0, time.UTC),
			},
			expected: &expected{
				nextExecutionDate: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual := tc.precondition.period.CalculateNextExecutionDate(tc.precondition.from, tc.precondition.now)

			assert.Equal(t, tc.expected.nextExecutionDate, actual)
			assert.Equal(t, tc.precondition.from, tc.precondition.period.SubtractFrom(tc.precondition.period.AddTo(tc.precondition.from)))
		})
	}
}

func TestParseAutomaticReportPeriod(t *testing.T) {
	t.Parallel()

	type precondition struct {
		raw string
	}

	type expected struct {
		period model.AutomaticReportPeriod
		isErr  bool
	}

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should parse weekly period",
			precondition: &precondition{
				raw: "weekly",
			},
			expected: &expected{
				period: model.AutomaticReportPeriodWeekly,
			},
		},
		{
			desc: "it should return error for unknown period",
			precondition: &precondition{
				raw: "hourly",
			},
			expected: &expected{
				isErr: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := model.ParseAutomaticReportPeriod(tc.precondition.raw)

			assert.Equal(t, tc.expected.isErr, err != nil)
			assert.Equal(t, tc.expected.period, actual)
		})
	}
}

func TestAutomaticReportPeriod_CalculateFirstExecutionDate(t *testing.T) {
	t.Parallel()

	type precondition struct {
		period model.AutomaticReportPeriod
		now    time.Time
	}

	type expected struct {
		executionDate time.Time
		reportStart   time.Time
	}

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should schedule end of day report at the next midnight",
			precondition: &precondition{
				period: model.AutomaticReportPeriodEndOfDay,
				now:    time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			},
			expected: &expected{
				executionDate: time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC),
				reportStart:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should schedule end of week report at the next Monday midnight",
			precondition: &precondition{
				period: model.AutomaticReportPeriodEndOfWeek,
				now:    time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			},
			expected: &expected{
				executionDate: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
				reportStart:   time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should schedule end of week report a week later when today is Monday",
			precondition: &precondition{
				period: model.AutomaticReportPeriodEndOfWeek,
				now:    time.Date(2025, 1, 13, 9, 0, 0, 0, time.UTC),
			},
			expected: &expected{
				executionDate: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
				reportStart:   time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should schedule end of month report at the first day of the next month",
			precondition: &precondition{
				period: model.AutomaticReportPeriodEndOfMonth,
				now:    time.Date(2025, 12, 15, 14, 30, 0, 0, time.UTC),
			},
			expected: &expected{
				executionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				reportStart:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should schedule end of quarter report at the first day of the next quarter",
			precondition: &precondition{
				period: model.AutomaticReportPeriodEndOfQuarter,
				now:    time.Date(2025, 5, 15, 14, 30, 0, 0, time.UTC),
			},
			expected: &expected{
				executionDate: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
				reportStart:   time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should schedule end of quarter report at the next year start for the last quarter",
			precondition: &precondition{
				period: model.AutomaticReportPeriodEndOfQuarter,
				now:    time.Date(2025, 11, 15, 14, 30, 0, 0, time.UTC),
			},
			expected: &expected{
				executionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				reportStart:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should schedule end of year report at the first day of the next year",
			precondition: &precondition{
				period: model.AutomaticReportPeriodEndOfYear,
				now:    time.Date(2025, 5, 15, 14, 30, 0, 0, time.UTC),
			},
			expected: &expected{
				executionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				reportStart:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should schedule rolling weekly report one week after today",
			precondition: &precondition{
				period: model.AutomaticReportPeriodWeekly,
				now:    time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			},
			expected: &expected{
				executionDate: time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC),
				reportStart:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual := tc.precondition.period.CalculateFirstExecutionDate(tc.precondition.now)

			assert.Equal(t, tc.expected.executionDate, actual)
			assert.Equal(t, tc.expected.reportStart, tc.precondition.period.SubtractFrom(actual))
		})
	}
}
