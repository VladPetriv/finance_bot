package errs_test

import (
	"fmt"
	"testing"

	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/stretchr/testify/assert"
)

func Test_IsExpected(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name     string
		args     error
		expected bool
	}{
		{
			name:     "should return true, since the error was custom",
			args:     errs.New("custom error"),
			expected: true,
		},
		{
			name:     "should return false, since the error wasn't custom",
			args:     fmt.Errorf("not custom error"),
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := errs.IsExpected(tc.args)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
