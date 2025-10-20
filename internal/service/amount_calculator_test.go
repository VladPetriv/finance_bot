package service

import (
	"testing"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/VladPetriv/finance_bot/pkg/typecast"
	"github.com/stretchr/testify/assert"
)

func TestCalculateIncomingOperation(t *testing.T) {
	t.Parallel()

	balance := money.NewFromInt(100)
	income := money.NewFromInt(50)

	calculateIncomingOperation(&balance, income)

	assert.Equal(t, "150.00", balance.StringFixed())
}

func TestCalculateUpdatedIncomingOperation(t *testing.T) {
	t.Parallel()

	balance := money.NewFromInt(150) // balance after initial 50 was added to 100
	initialAmount := money.NewFromInt(50)
	updateAmount := money.NewFromInt(80)

	calculateUpdatedIncomingOperation(&balance, initialAmount, updateAmount)

	assert.Equal(t, "180.00", balance.StringFixed())
}

func TestCalculateSpendingOperation(t *testing.T) {
	t.Parallel()

	balance := money.NewFromInt(100)
	spending := money.NewFromInt(30)

	calculateSpendingOperation(&balance, spending)

	assert.Equal(t, "70.00", balance.StringFixed())
}

func TestCalculateUpdatedSpendingOperation(t *testing.T) {
	t.Parallel()

	balance := money.NewFromInt(70) // balance after initial 30 was subtracted from 100
	initialAmount := money.NewFromInt(30)
	updateAmount := money.NewFromInt(45)

	calculateUpdatedSpendingOperation(&balance, initialAmount, updateAmount)

	assert.Equal(t, "55.00", balance.StringFixed())
}

func TestCalculateTransferOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc         string
		opts         calculateTransferOperationOptions
		expectedFrom string
		expectedTo   string
	}{
		{
			desc: "simple transfer without exchange rate",
			opts: calculateTransferOperationOptions{
				balanceFrom:     typecast.ToPtr(money.NewFromInt(200)),
				balanceTo:       typecast.ToPtr(money.NewFromInt(50)),
				operationAmount: money.NewFromInt(75),
			},
			expectedFrom: "125.00",
			expectedTo:   "125.00",
		},
		{
			desc: "transfer with exchange rate",
			opts: calculateTransferOperationOptions{
				balanceFrom:     typecast.ToPtr(money.NewFromInt(1000)),
				balanceTo:       typecast.ToPtr(money.NewFromInt(100)),
				operationAmount: money.NewFromInt(100),
				exchangeRate:    typecast.ToPtr(money.NewFromFloat(1.5)),
			},
			expectedFrom: "900.00",
			expectedTo:   "250.00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			calculateTransferOperation(tc.opts)

			assert.Equal(t, tc.expectedFrom, tc.opts.balanceFrom.StringFixed())
			assert.Equal(t, tc.expectedTo, tc.opts.balanceTo.StringFixed())
		})
	}
}

func TestCalculateUpdatedTransferOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc                      string
		opts                      calculateTransferOperationOptions
		expectedBalanceFrom       string
		expectedBalanceTo         string
		expectedTransferAmountIn  string
		expectedTransferAmountOut string
	}{
		{
			desc: "update transfer without exchange rate",
			opts: calculateTransferOperationOptions{
				balanceFrom:            typecast.ToPtr(money.NewFromInt(125)),
				balanceTo:              typecast.ToPtr(money.NewFromInt(125)),
				transferAmountOut:      typecast.ToPtr(money.NewFromInt(75)),
				transferAmountIn:       typecast.ToPtr(money.NewFromInt(75)),
				updatedOperationAmount: money.NewFromInt(100),
			},
			expectedBalanceFrom:       "100.00",
			expectedBalanceTo:         "150.00",
			expectedTransferAmountIn:  "100.00",
			expectedTransferAmountOut: "100.00",
		},
		{
			desc: "update transfer_in with exchange rate",
			opts: calculateTransferOperationOptions{
				operationType:          model.OperationTypeTransferIn,
				balanceFrom:            typecast.ToPtr(money.NewFromInt(900)),
				balanceTo:              typecast.ToPtr(money.NewFromInt(250)),
				transferAmountOut:      typecast.ToPtr(money.NewFromInt(100)),
				transferAmountIn:       typecast.ToPtr(money.NewFromInt(150)),
				updatedOperationAmount: money.NewFromInt(120),
				exchangeRate:           typecast.ToPtr(money.NewFromFloat(1.5)),
			},
			expectedBalanceFrom:       "920.00",
			expectedBalanceTo:         "220.00",
			expectedTransferAmountIn:  "120.00",
			expectedTransferAmountOut: "80.00",
		},
		{
			desc: "update transfer_out with exchange rate",
			opts: calculateTransferOperationOptions{
				operationType:          model.OperationTypeTransferOut,
				balanceFrom:            typecast.ToPtr(money.NewFromInt(900)),
				balanceTo:              typecast.ToPtr(money.NewFromInt(250)),
				transferAmountOut:      typecast.ToPtr(money.NewFromInt(100)),
				transferAmountIn:       typecast.ToPtr(money.NewFromInt(150)),
				updatedOperationAmount: money.NewFromInt(80),
				exchangeRate:           typecast.ToPtr(money.NewFromFloat(1.5)),
			},
			expectedBalanceFrom:       "920.00",
			expectedBalanceTo:         "220.00",
			expectedTransferAmountIn:  "120.00",
			expectedTransferAmountOut: "80.00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			calculateUpdatedTranferOperation(tc.opts)

			assert.Equal(t, tc.expectedBalanceFrom, tc.opts.balanceFrom.StringFixed())
			assert.Equal(t, tc.expectedBalanceTo, tc.opts.balanceTo.StringFixed())

			if tc.expectedTransferAmountIn != "" {
				assert.Equal(t, tc.expectedTransferAmountIn, tc.opts.transferAmountIn.StringFixed())
			}
			if tc.expectedTransferAmountOut != "" {
				assert.Equal(t, tc.expectedTransferAmountOut, tc.opts.transferAmountOut.StringFixed())
			}
		})
	}
}
