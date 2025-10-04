package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoney_MathOperation(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc            string
		initialAmount   Money
		amountOperation func(initialAmount *Money)
		expected        Money
	}{
		{
			desc:          "Should increase initial amount by 10",
			initialAmount: NewFromInt(10),
			amountOperation: func(initialAmount *Money) {
				initialAmount.Inc(NewFromInt(10))
			},
			expected: NewFromInt(20),
		},
		{
			desc:          "Should decrease initial amount by 10",
			initialAmount: NewFromInt(10),
			amountOperation: func(initialAmount *Money) {
				initialAmount.Sub(NewFromInt(10))
			},
			expected: NewFromInt(0),
		},
		{
			desc:          "Should multiply initial amount by 2",
			initialAmount: NewFromInt(10),
			amountOperation: func(initialAmount *Money) {
				initialAmount.Mul(NewFromInt(2))
			},
			expected: NewFromInt(20),
		},
		{
			desc:          "Should divide initial amount by 2",
			initialAmount: NewFromInt(10),
			amountOperation: func(initialAmount *Money) {
				initialAmount.Div(NewFromInt(2))
			},
			expected: NewFromInt(5),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			tc.amountOperation(&tc.initialAmount)
			assert.True(t, tc.expected.Equal(tc.initialAmount))
		})
	}
}

func TestMoney_Equal(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc        string
		left, right Money
		expected    bool
	}{
		{
			desc:     "Should return true for equal values",
			left:     NewFromInt(10),
			right:    NewFromInt(10),
			expected: true,
		},
		{
			desc:     "Should return false for different values",
			left:     NewFromInt(10),
			right:    NewFromInt(20),
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.left.Equal(tc.right))
		})
	}
}

func TestMoney_GreaterThan(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc        string
		left, right Money
		expected    bool
	}{
		{
			desc:     "Should return true for greater values",
			left:     NewFromInt(20),
			right:    NewFromInt(10),
			expected: true,
		},
		{
			desc:     "Should return false for equal values",
			left:     NewFromInt(10),
			right:    NewFromInt(10),
			expected: false,
		},
		{
			desc:     "Should return false for smaller values",
			left:     NewFromInt(10),
			right:    NewFromInt(20),
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.left.GreaterThan(tc.right))
		})
	}
}

func TestMoney_StringFixed(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc          string
		initialAmount Money
		expected      string
	}{
		{
			desc:          "Should return string representation of float with 2 places after digit",
			initialAmount: NewFromFloat(10.12),
			expected:      "10.12",
		},
		{
			desc:          "Should return string representation of float with 5 places after digit with limitation to 2 decimal places",
			initialAmount: NewFromFloat(100.12345),
			expected:      "100.12",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.initialAmount.StringFixed())
		})
	}
}

func TestMoney_String(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc          string
		initialAmount Money
		expected      string
	}{
		{
			desc:          "Should return string representation of float with 2 places after digit",
			initialAmount: NewFromFloat(10.12),
			expected:      "10.12",
		},
		{
			desc:          "Should return string representation of float with 5 places after digit without any limitation",
			initialAmount: NewFromFloat(100.12345),
			expected:      "100.12345",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.initialAmount.String())
		})
	}
}
