package models

import (
	"testing"

	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/stretchr/testify/assert"
)

func TestCalculateOperationsStatistics(t *testing.T) {
	t.Parallel()

	amount200, _ := money.NewFromString("200.00")
	amount150, _ := money.NewFromString("150.00")
	amount100, _ := money.NewFromString("100.00")
	amount50, _ := money.NewFromString("50.00")

	type expected struct {
		stats *OperationsStatistics
		err   bool
	}

	testCases := [...]struct {
		desc     string
		args     []Operation
		expected expected
	}{
		{
			desc: "positive: calculate statistics for all operation types",
			args: []Operation{
				{Type: OperationTypeIncoming, Amount: "100.00"},
				{Type: OperationTypeIncoming, Amount: "100.00"},
				{Type: OperationTypeSpending, Amount: "50.00"},
				{Type: OperationTypeSpending, Amount: "50.00"},
				{Type: OperationTypeTransferIn, Amount: "75.00"},
				{Type: OperationTypeTransferIn, Amount: "75.00"},
				{Type: OperationTypeTransferOut, Amount: "25.00"},
				{Type: OperationTypeTransferOut, Amount: "25.00"},
			},
			expected: expected{
				stats: &OperationsStatistics{
					IncomingCount:    2,
					SpendingCount:    2,
					TransferInCount:  2,
					TransferOutCount: 2,
					IncomingTotal:    amount200,
					SpendingTotal:    amount100,
					TransferInTotal:  amount150,
					TransferOutTotal: amount50,
				},
			},
		},
		{
			desc: "positive: calculate statistics with empty operations",
			args: []Operation{},
			expected: expected{
				stats: &OperationsStatistics{
					IncomingTotal:    money.Zero,
					SpendingTotal:    money.Zero,
					TransferInTotal:  money.Zero,
					TransferOutTotal: money.Zero,
					OperationsByType: make(map[OperationType][]Operation),
				},
			},
		},
		{
			desc: "negative: operation amount is invalid",
			args: []Operation{
				{Type: OperationTypeIncoming, Amount: "invalid"},
			},
			expected: expected{
				err: true,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := CalculateOperationsStatistics(tc.args)
			if tc.expected.err {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expected.stats.IncomingCount, actual.IncomingCount)
			assert.Equal(t, tc.expected.stats.SpendingCount, actual.SpendingCount)
			assert.Equal(t, tc.expected.stats.TransferInCount, actual.TransferInCount)
			assert.Equal(t, tc.expected.stats.TransferOutCount, actual.TransferOutCount)
			assert.Equal(t, tc.expected.stats.IncomingTotal, actual.IncomingTotal)
			assert.Equal(t, tc.expected.stats.SpendingTotal, actual.SpendingTotal)
			assert.Equal(t, tc.expected.stats.TransferInTotal, actual.TransferInTotal)
			assert.Equal(t, tc.expected.stats.TransferOutTotal, actual.TransferOutTotal)
		})
	}
}

func TestCalculateCategoryStatistics(t *testing.T) {
	t.Parallel()

	amount100, _ := money.NewFromString("100.00")
	amount70, _ := money.NewFromString("70.00")
	amount30, _ := money.NewFromString("30.00")

	type args struct {
		totalAmount money.Money
		operations  []Operation
		categories  []Category
	}

	type expected struct {
		stats []CategoryStatistics
		err   bool
	}

	testCases := [...]struct {
		desc     string
		args     args
		expected expected
	}{
		{
			desc: "positive: calculate category statistics",

			args: args{
				totalAmount: amount100,
				operations: []Operation{
					{CategoryID: "1", Amount: "50.00"},
					{CategoryID: "2", Amount: "30.00"},
					{CategoryID: "1", Amount: "20.00"},
				},
				categories: []Category{
					{ID: "1", Title: "Food"},
					{ID: "2", Title: "Transport"},
				},
			},
			expected: expected{
				stats: []CategoryStatistics{
					{
						Title:      "Food",
						Amount:     amount70,
						Percentage: amount70,
					},
					{
						Title:      "Transport",
						Amount:     amount30,
						Percentage: amount30,
					},
				},
			},
		},
		{
			desc: "positive: empty operations",
			args: args{
				totalAmount: amount100,
				operations:  []Operation{},
				categories: []Category{
					{ID: "1", Title: "Food"},
				},
			},

			expected: expected{
				stats: []CategoryStatistics{},
			},
		},
		{
			desc: "negative: invalid amount format",
			args: args{
				totalAmount: amount100,
				operations: []Operation{
					{CategoryID: "1", Amount: "invalid"},
				},
				categories: []Category{
					{ID: "1", Title: "Food"},
				},
			},
			expected: expected{
				err: true,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := CalculateCategoryStatistics(tc.args.totalAmount, tc.args.operations, tc.args.categories)
			if tc.expected.err {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(tc.expected.stats), len(actual))

			for i := range tc.expected.stats {
				assert.Equal(t, tc.expected.stats[i].Title, actual[i].Title)
				assert.Equal(t, tc.expected.stats[i].Amount.StringFixed(), actual[i].Amount.StringFixed())
				assert.Equal(t, tc.expected.stats[i].Percentage.StringFixed(), actual[i].Percentage.StringFixed())
			}
		})
	}
}
