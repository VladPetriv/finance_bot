package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/stretchr/testify/assert"
)

func Test_calculateOperationsStatistics(t *testing.T) {
	t.Parallel()

	amount200, _ := money.NewFromString("200.00")
	amount150, _ := money.NewFromString("150.00")
	amount100, _ := money.NewFromString("100.00")
	amount50, _ := money.NewFromString("50.00")

	type expected struct {
		stats *operationsStatistics
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
				stats: &operationsStatistics{
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
				stats: &operationsStatistics{
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

			actual, err := calculateOperationsStatistics(tc.args)
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

func Test_calculateCategoryStatistics(t *testing.T) {
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
		stats []categoryStatistics
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
				stats: []categoryStatistics{
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
				stats: []categoryStatistics{},
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

			actual, err := calculateCategoryStatistics(tc.args.totalAmount, tc.args.operations, tc.args.categories)
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

func TestStatisticsMessageBuilder_Build(t *testing.T) {
	t.Parallel()

	type args struct {
		balance    *Balance
		operations []Operation
		categories []Category
	}

	type expected struct {
		message string
		err     bool
	}

	testCases := [...]struct {
		desc     string
		args     args
		expected expected
	}{
		{
			desc: "positive: builds statistics message with all operation types",
			args: args{
				balance: &Balance{
					Name:     "Main Balance",
					Amount:   "1000.00",
					Currency: &Currency{Symbol: "$"},
				},
				operations: []Operation{
					{Type: OperationTypeIncoming, Amount: "100.00", CategoryID: "1"},
					{Type: OperationTypeSpending, Amount: "50.00", CategoryID: "2"},
					{Type: OperationTypeTransferIn, Amount: "75.00"},
					{Type: OperationTypeTransferOut, Amount: "25.00"},
				},
				categories: []Category{
					{ID: "1", Title: "Salary"},
					{ID: "2", Title: "Food"},
				},
			},
			expected: expected{
				message: fmt.Sprintf(`📊 Balance Statistics: *Main Balance*
💰 Current Balance: `+"`1000.00$`"+`

📅 Period: _%s - %s_

📈 Summary:
📥 Incoming Operations: `+"`100.00$`"+` *(1)*
			- Salary: `+"`100.00$`"+` *(100.00%%)*

💸 Spending Operations: `+"`50.00$`"+` *(1)*
			- Food: `+"`50.00$`"+` *(100.00%%)*

🔄 Transfers Operations *(2)*:
		 ➡️ In: `+"`75.00$`"+` *(1)*
			⬅️ Out: `+"`25.00$`"+` *(1)*
	`, time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateFormat), time.Now().Format(dateFormat)),
			},
		},
		{
			desc: "positive: builds statistics message only incoming operation type",
			args: args{
				balance: &Balance{
					Name:     "Main Balance",
					Amount:   "1000.00",
					Currency: &Currency{Symbol: "$"},
				},
				operations: []Operation{
					{Type: OperationTypeIncoming, Amount: "100.00", CategoryID: "1"},
				},
				categories: []Category{
					{ID: "1", Title: "Salary"},
					{ID: "2", Title: "Food"},
				},
			},
			expected: expected{
				message: fmt.Sprintf(`📊 Balance Statistics: *Main Balance*
💰 Current Balance: `+"`1000.00$`"+`

📅 Period: _%s - %s_

📈 Summary:
📥 Incoming Operations: `+"`100.00$`"+` *(1)*
			- Salary: `+"`100.00$`"+` *(100.00%%)*

💸 Spending Operations: `+"`0.00$`"+` *(0)*
🔄 Transfers Operations *(0)*:
		 ➡️ In: `+"`0.00$`"+` *(0)*
			⬅️ Out: `+"`0.00$`"+` *(0)*
	`, time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateFormat), time.Now().Format(dateFormat)),
			},
		},
		{
			desc: "positive: builds statistics message only spending operation type",
			args: args{
				balance: &Balance{
					Name:     "Main Balance",
					Amount:   "1000.00",
					Currency: &Currency{Symbol: "$"},
				},
				operations: []Operation{
					{Type: OperationTypeSpending, Amount: "50.00", CategoryID: "2"},
				},
				categories: []Category{
					{ID: "1", Title: "Salary"},
					{ID: "2", Title: "Food"},
				},
			},
			expected: expected{
				message: fmt.Sprintf(`📊 Balance Statistics: *Main Balance*
💰 Current Balance: `+"`1000.00$`"+`

📅 Period: _%s - %s_

📈 Summary:
📥 Incoming Operations: `+"`0.00$`"+` *(0)*
💸 Spending Operations: `+"`50.00$`"+` *(1)*
			- Food: `+"`50.00$`"+` *(100.00%%)*

🔄 Transfers Operations *(0)*:
		 ➡️ In: `+"`0.00$`"+` *(0)*
			⬅️ Out: `+"`0.00$`"+` *(0)*
	`, time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateFormat), time.Now().Format(dateFormat)),
			},
		},
		{
			desc: "positive: builds statistics message only transfer in operation type",
			args: args{
				balance: &Balance{
					Name:     "Main Balance",
					Amount:   "1000.00",
					Currency: &Currency{Symbol: "$"},
				},
				operations: []Operation{
					{Type: OperationTypeTransferIn, Amount: "75.00"},
				},
			},
			expected: expected{
				message: fmt.Sprintf(`📊 Balance Statistics: *Main Balance*
💰 Current Balance: `+"`1000.00$`"+`

📅 Period: _%s - %s_

📈 Summary:
📥 Incoming Operations: `+"`0.00$`"+` *(0)*
💸 Spending Operations: `+"`0.00$`"+` *(0)*
🔄 Transfers Operations *(1)*:
		 ➡️ In: `+"`75.00$`"+` *(1)*
			⬅️ Out: `+"`0.00$`"+` *(0)*
	`, time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateFormat), time.Now().Format(dateFormat)),
			},
		},
		{
			desc: "positive: builds statistics message only transfer out operation type",
			args: args{
				balance: &Balance{
					Name:     "Main Balance",
					Amount:   "1000.00",
					Currency: &Currency{Symbol: "$"},
				},
				operations: []Operation{
					{Type: OperationTypeTransferOut, Amount: "75.00"},
				},
			},
			expected: expected{
				message: fmt.Sprintf(`📊 Balance Statistics: *Main Balance*
💰 Current Balance: `+"`1000.00$`"+`

📅 Period: _%s - %s_

📈 Summary:
📥 Incoming Operations: `+"`0.00$`"+` *(0)*
💸 Spending Operations: `+"`0.00$`"+` *(0)*
🔄 Transfers Operations *(1)*:
		 ➡️ In: `+"`0.00$`"+` *(0)*
			⬅️ Out: `+"`75.00$`"+` *(1)*
	`, time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateFormat), time.Now().Format(dateFormat)),
			},
		},
		{
			desc: "positive: builds statistics message with no operations",
			args: args{
				balance: &Balance{
					Name:     "Empty Balance",
					Amount:   "0.00",
					Currency: &Currency{Symbol: "$"},
				},
				operations: []Operation{},
				categories: []Category{},
			},
			expected: expected{
				message: fmt.Sprintf(`📊 Balance Statistics: *Empty Balance*
💰 Current Balance: `+"`0.00$`"+`

📅 Period: _%s - %s_

📈 Summary:
📥 Incoming Operations: `+"`0.00$`"+` *(0)*
💸 Spending Operations: `+"`0.00$`"+` *(0)*
🔄 Transfers Operations *(0)*:
		 ➡️ In: `+"`0.00$`"+` *(0)*
			⬅️ Out: `+"`0.00$`"+` *(0)*
	`, time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateFormat), time.Now().Format(dateFormat)),
			},
		},
		{
			desc: "negative: invalid operation amount",
			args: args{
				balance: &Balance{
					Name:     "Main Balance",
					Amount:   "1000.00",
					Currency: &Currency{Symbol: "$"},
				},
				operations: []Operation{
					{Type: OperationTypeIncoming, Amount: "invalid"},
				},
				categories: []Category{},
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

			builder := NewStatisticsMessageBuilder(tc.args.balance, tc.args.operations, tc.args.categories)
			message, err := builder.Build(convertToMonth(int(time.Now().Month())))

			if tc.expected.err {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expected.message, message)
		})
	}
}

func convertToMonth(monthIndex int) Month {
	switch monthIndex {
	case 1:
		return MonthJanuary
	case 2:
		return MonthFebruary
	case 3:
		return MonthMarch
	case 4:
		return MonthApril
	case 5:
		return MonthMay
	case 6:
		return MonthJune
	case 7:
		return MonthJuly
	case 8:
		return MonthAugust
	case 9:
		return MonthSeptember
	case 10:
		return MonthOctober
	case 11:
		return MonthNovember
	case 12:
		return MonthDecember
	default:
		return MonthJanuary
	}
}
