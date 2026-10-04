package model_test

import (
	"testing"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestUserSettings_GetLocation(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc     string
		timezone string
		expected string
	}{
		{
			desc:     "it should return configured location",
			timezone: "Pacific/Auckland",
			expected: "Pacific/Auckland",
		},
		{
			desc:     "it should fall back to UTC when timezone is empty",
			timezone: "",
			expected: "UTC",
		},
		{
			desc:     "it should fall back to UTC when timezone is unknown",
			timezone: "Mars/Olympus",
			expected: "UTC",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			settings := &model.UserSettings{
				Timezone: tc.timezone,
			}

			assert.Equal(t, tc.expected, settings.GetLocation().String())
		})
	}
}
