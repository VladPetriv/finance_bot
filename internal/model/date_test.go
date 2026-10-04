package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonth_GetName(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc     string
		month    Month
		expected string
	}{
		{
			desc:     "positive: get January name",
			month:    MonthJanuary,
			expected: "January",
		},
		{
			desc:     "positive: get June name",
			month:    MonthJune,
			expected: "June",
		},
		{
			desc:     "positive: get December name",
			month:    MonthDecember,
			expected: "December",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual := tc.month.GetName()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestMonth_GetIndex(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc     string
		month    Month
		expected int
	}{
		{
			desc:     "positive: get January index",
			month:    MonthJanuary,
			expected: 1,
		},
		{
			desc:     "positive: get June index",
			month:    MonthJune,
			expected: 6,
		},
		{
			desc:     "positive: get December index",
			month:    MonthDecember,
			expected: 12,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual := tc.month.GetIndex()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestMonth_GetTimeRange(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)

	testCases := []struct {
		desc        string
		month       Month
		currentTime time.Time
		year        int
		expected    struct {
			start time.Time
			end   time.Time
		}
	}{
		{
			desc:        "current month same year -> end is now",
			month:       MonthMarch,
			currentTime: fixedTime,
			year:        0,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
				end:   fixedTime,
			},
		},
		{
			desc:        "same month but different year -> full month",
			month:       MonthMarch,
			currentTime: fixedTime,
			year:        2023,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc:        "different month same year",
			month:       MonthJanuary,
			currentTime: fixedTime,
			year:        0,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc:        "december -> january transition",
			month:       MonthDecember,
			currentTime: fixedTime,
			year:        2024,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc:        "leap year february",
			month:       MonthFebruary,
			currentTime: fixedTime,
			year:        2024, // leap year
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			desc:        "non-leap year february",
			month:       MonthFebruary,
			currentTime: fixedTime,
			year:        2023,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			start, end := tc.month.GetTimeRange(tc.currentTime, tc.year, time.UTC)
			assert.Equal(t, tc.expected.start, start)
			assert.Equal(t, tc.expected.end, end)
		})
	}
}

func TestCreationPeriod_CalculateTimeRange(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	testCases := [...]struct {
		desc     string
		period   CreationPeriod
		expected struct {
			start time.Time
			end   time.Time
		}
	}{
		{
			desc:   "positive: calculate day range",
			period: CreationPeriodDay,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: now.Add(-24 * time.Hour),
				end:   now,
			},
		},
		{
			desc:   "positive: calculate week range",
			period: CreationPeriodWeek,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: now.Add(-7 * 24 * time.Hour),
				end:   now,
			},
		},
		{
			desc:   "positive: calculate current month range",
			period: CreationPeriodCurrentMonth,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
				end:   now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			start, end := tc.period.CalculateTimeRange(time.UTC)

			// Using truncate to ignore small time differences during test execution
			assert.Equal(t, tc.expected.start.Truncate(time.Second), start.Truncate(time.Second))
			assert.Equal(t, tc.expected.end.Truncate(time.Second), end.Truncate(time.Second))
		})
	}
}

func TestGetCreationPeriodFromText(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc     string
		text     string
		expected CreationPeriod
	}{
		{
			desc:     "positive: get day period",
			text:     "day",
			expected: CreationPeriodDay,
		},
		{
			desc:     "positive: get week period",
			text:     "week",
			expected: CreationPeriodWeek,
		},
		{
			desc:     "positive: get month period",
			text:     "month",
			expected: CreationPeriodMonth,
		},
		{
			desc:     "positive: get year period",
			text:     "year",
			expected: CreationPeriodYear,
		},
		{
			desc:     "negative: invalid period",
			text:     "invalid",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual := GetCreationPeriodFromText(tc.text)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestMonth_GetTimeRange_WithLocation(t *testing.T) {
	t.Parallel()

	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)

	losAngeles, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	type args struct {
		month       Month
		currentTime time.Time
		year        int
		location    *time.Location
	}

	type expected struct {
		start time.Time
		end   time.Time
	}

	testCases := [...]struct {
		desc     string
		args     *args
		expected *expected
	}{
		{
			desc: "it should start the month at local midnight which is the previous UTC day for a zone ahead of UTC",
			args: &args{
				month:       MonthMarch,
				currentTime: time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC),
				year:        2024,
				location:    auckland,
			},
			expected: &expected{
				start: time.Date(2024, 2, 29, 11, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 3, 31, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should start the month at local midnight which is the same UTC day later for a zone behind UTC",
			args: &args{
				month:       MonthMarch,
				currentTime: time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC),
				year:        2024,
				location:    losAngeles,
			},
			expected: &expected{
				start: time.Date(2024, 3, 1, 8, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 4, 1, 7, 0, 0, 0, time.UTC),
			},
		},
		{
			desc: "it should treat the current month by local date and cut the range at now",
			args: &args{
				month:       MonthJuly,
				currentTime: time.Date(2024, 6, 30, 13, 0, 0, 0, time.UTC),
				year:        0,
				location:    auckland,
			},
			expected: &expected{
				start: time.Date(2024, 6, 30, 12, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 6, 30, 13, 0, 0, 0, time.UTC),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			start, end := tc.args.month.GetTimeRange(tc.args.currentTime, tc.args.year, tc.args.location)

			assert.Equal(t, tc.expected.start, start)
			assert.Equal(t, tc.expected.end, end)
		})
	}
}
