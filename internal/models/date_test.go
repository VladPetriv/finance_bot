package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		tc := tc
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
		tc := tc
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

	testCases := [...]struct {
		desc        string
		month       Month
		currentTime time.Time
		expected    struct {
			start time.Time
			end   time.Time
		}
	}{
		{
			desc:        "positive: get time range for current month",
			month:       MonthMarch,
			currentTime: fixedTime,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC),
			},
		},
		{
			desc:        "positive: get time range for different month",
			month:       MonthJanuary,
			currentTime: fixedTime,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			},
		},
		{
			desc:        "positive: get time range for December with year transition",
			month:       MonthDecember,
			currentTime: fixedTime,
			expected: struct {
				start time.Time
				end   time.Time
			}{
				start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			start, end := tc.month.GetTimeRange(tc.currentTime)
			assert.Equal(t, tc.expected.start, start)
			assert.Equal(t, tc.expected.end, end)
		})
	}
}

func TestCreationPeriod_CalculateTimeRange(t *testing.T) {
	t.Parallel()

	now := time.Now()

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
				start: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local),
				end:   now,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			start, end := tc.period.CalculateTimeRange()

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
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual := GetCreationPeriodFromText(tc.text)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
