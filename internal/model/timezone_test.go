package model_test

import (
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTimezonesByRegion(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc   string
		region string
	}{
		{
			desc:   "it should return only valid IANA names for Europe",
			region: "Europe",
		},
		{
			desc:   "it should return only valid IANA names for America",
			region: "America",
		},
		{
			desc:   "it should return only valid IANA names for Asia",
			region: "Asia",
		},
		{
			desc:   "it should return only valid IANA names for Africa",
			region: "Africa",
		},
		{
			desc:   "it should return only valid IANA names for Pacific",
			region: "Pacific",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			timezones := model.GetTimezonesByRegion(tc.region)
			require.NotEmpty(t, timezones)

			for _, timezone := range timezones {
				_, err := time.LoadLocation(timezone.GetID())
				assert.NoError(t, err, timezone)
			}
		})
	}
}

func TestTimezoneRegions(t *testing.T) {
	t.Parallel()

	for _, region := range model.TimezoneRegions {
		assert.NotEmpty(t, model.GetTimezonesByRegion(region), region)
	}

	assert.NotContains(t, model.GetTimezonesByRegion("Europe"), model.Timezone("Europe/Moscow"))
	assert.Contains(t, model.GetTimezonesByRegion("Europe"), model.Timezone("Europe/Kyiv"))
}

func TestTimezone_GetName(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc     string
		timezone model.Timezone
		expected string
	}{
		{
			desc:     "it should return the city of a two-part name",
			timezone: "Europe/Kyiv",
			expected: "Kyiv",
		},
		{
			desc:     "it should replace underscores with spaces",
			timezone: "America/New_York",
			expected: "New York",
		},
		{
			desc:     "it should keep nested part of a three-part name",
			timezone: "America/Argentina/Buenos_Aires",
			expected: "Argentina/Buenos Aires",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.timezone.GetName())
		})
	}
}
