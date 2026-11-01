package money

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestMoney_LessThan(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		desc        string
		left, right Money
		expected    bool
	}{
		{
			desc:     "Should return false for greater values",
			left:     NewFromInt(20),
			right:    NewFromInt(10),
			expected: false,
		},
		{
			desc:     "Should return false for equal values",
			left:     NewFromInt(10),
			right:    NewFromInt(10),
			expected: false,
		},
		{
			desc:     "Should return true for smaller values",
			left:     NewFromInt(10),
			right:    NewFromInt(20),
			expected: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.left.LessThan(tc.right))
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

func TestMoney_MinorUnits(t *testing.T) {
	t.Parallel()

	type precondition struct {
		amount string
	}

	type expected struct {
		minorUnits int64
	}

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should convert whole amount to minor units",
			precondition: &precondition{
				amount: "10",
			},
			expected: &expected{
				minorUnits: 1000,
			},
		},
		{
			desc: "it should convert amount with two decimals to minor units",
			precondition: &precondition{
				amount: "12.34",
			},
			expected: &expected{
				minorUnits: 1234,
			},
		},
		{
			desc: "it should convert zero to zero minor units",
			precondition: &precondition{
				amount: "0",
			},
			expected: &expected{
				minorUnits: 0,
			},
		},
		{
			desc: "it should convert negative amount to negative minor units",
			precondition: &precondition{
				amount: "-12.34",
			},
			expected: &expected{
				minorUnits: -1234,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			amount, err := NewFromString(tc.precondition.amount)
			require.NoError(t, err)

			actual, err := amount.MinorUnits()
			require.NoError(t, err)
			assert.Equal(t, tc.expected.minorUnits, actual)

			roundTripped := NewFromMinorUnits(actual)
			assert.True(t, amount.Equal(roundTripped))
		})
	}
}

func TestMoney_MinorUnits_Rounding(t *testing.T) {
	t.Parallel()

	type precondition struct {
		amount string
	}

	type expected struct {
		minorUnits int64
	}

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should round half away from zero for positive amount",
			precondition: &precondition{
				amount: "12.345",
			},
			expected: &expected{
				minorUnits: 1235,
			},
		},
		{
			desc: "it should round half away from zero for negative amount",
			precondition: &precondition{
				amount: "-12.345",
			},
			expected: &expected{
				minorUnits: -1235,
			},
		},
		{
			desc: "it should round down below half",
			precondition: &precondition{
				amount: "9.173",
			},
			expected: &expected{
				minorUnits: 917,
			},
		},
		{
			desc: "it should round up above half",
			precondition: &precondition{
				amount: "9.176",
			},
			expected: &expected{
				minorUnits: 918,
			},
		},
		{
			desc: "it should round tiny positive amount to zero",
			precondition: &precondition{
				amount: "0.004",
			},
			expected: &expected{
				minorUnits: 0,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			amount, err := NewFromString(tc.precondition.amount)
			require.NoError(t, err)

			actual, err := amount.MinorUnits()
			require.NoError(t, err)
			assert.Equal(t, tc.expected.minorUnits, actual)
		})
	}
}

func TestMoney_MinorUnits_Overflow(t *testing.T) {
	t.Parallel()

	type precondition struct {
		amount string
	}

	type expected struct {
		minorUnits int64
		err        error
	}

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should convert the largest representable amount",
			precondition: &precondition{
				amount: "92233720368547758.07",
			},
			expected: &expected{
				minorUnits: math.MaxInt64,
			},
		},
		{
			desc: "it should convert the smallest representable amount",
			precondition: &precondition{
				amount: "-92233720368547758.08",
			},
			expected: &expected{
				minorUnits: math.MinInt64,
			},
		},
		{
			desc: "it should return error when positive amount does not fit",
			precondition: &precondition{
				amount: "92233720368547758.08",
			},
			expected: &expected{
				err: errMinorUnitsOverflow,
			},
		},
		{
			desc: "it should return error when negative amount does not fit",
			precondition: &precondition{
				amount: "-92233720368547758.09",
			},
			expected: &expected{
				err: errMinorUnitsOverflow,
			},
		},
		{
			desc: "it should return error when rounding pushes amount over the limit",
			precondition: &precondition{
				amount: "92233720368547758.075",
			},
			expected: &expected{
				err: errMinorUnitsOverflow,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			amount, err := NewFromString(tc.precondition.amount)
			require.NoError(t, err)

			actual, err := amount.MinorUnits()
			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.minorUnits, actual)
		})
	}
}

func TestMoney_ValueAndScan(t *testing.T) {
	t.Parallel()

	value, err := NewFromMinorUnits(1234).Value()
	require.NoError(t, err)
	assert.Equal(t, int64(1234), value)

	var scanned Money
	require.NoError(t, scanned.Scan(int64(-505)))
	assert.True(t, scanned.Equal(NewFromMinorUnits(-505)))

	assert.Error(t, scanned.Scan("12.34"))
	assert.Error(t, scanned.Scan(nil))
}
