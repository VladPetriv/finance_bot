package service

import (
	"testing"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/VladPetriv/finance_bot/pkg/typecast"
	"github.com/stretchr/testify/assert"
)

func TestCalculateBalanceAmountBasedOnOperationType(t *testing.T) {
	t.Parallel()

	type args struct {
		actionType    calculationActionType
		operationType model.OperationType
		opts          calculationOptions
	}

	type expected struct {
		balanceFrom string
		balanceTo   string
		balance     string
	}

	testCases := [...]struct {
		desc     string
		args     args
		expected expected
	}{
		{
			desc: "should increase balance with create action and incoming operation",
			args: args{
				actionType:    calculationActionTypeCreate,
				operationType: model.OperationTypeIncoming,
				opts: calculationOptions{
					balance:         typecast.ToPtr(money.NewFromInt(100)),
					operationAmount: money.NewFromInt(50),
				},
			},
			expected: expected{
				balance: "150.00",
			},
		},
		{
			desc: "should decrease balance with create action and spending operation",
			args: args{
				actionType:    calculationActionTypeCreate,
				operationType: model.OperationTypeSpending,
				opts: calculationOptions{
					balance:         typecast.ToPtr(money.NewFromInt(100)),
					operationAmount: money.NewFromInt(30),
				},
			},
			expected: expected{
				balance: "70.00",
			},
		},
		{
			desc: "should transfer from one balance to another with create action and transfer operation",
			args: args{
				actionType:    calculationActionTypeCreate,
				operationType: model.OperationTypeTransfer,
				opts: calculationOptions{
					balanceFrom:     typecast.ToPtr(money.NewFromInt(200)),
					balanceTo:       typecast.ToPtr(money.NewFromInt(50)),
					operationAmount: money.NewFromInt(75),
				},
			},
			expected: expected{
				balanceFrom: "125.00",
				balanceTo:   "125.00",
			},
		},
		{
			desc: "should transfer from one balance to another including exchange rate with create action and transfer operation",
			args: args{
				actionType:    calculationActionTypeCreate,
				operationType: model.OperationTypeTransfer,
				opts: calculationOptions{
					balanceFrom:     typecast.ToPtr(money.NewFromInt(1000)),
					balanceTo:       typecast.ToPtr(money.NewFromInt(100)),
					operationAmount: money.NewFromInt(100),
					exchangeRate:    typecast.ToPtr(money.NewFromFloat(1.5)),
				},
			},
			expected: expected{
				balanceFrom: "900.00",
				balanceTo:   "250.00",
			},
		},
		{
			desc: "should update balance with updating incoming operation amount",
			args: args{
				actionType:    calculationActionTypeUpdate,
				operationType: model.OperationTypeIncoming,
				opts: calculationOptions{
					balance:                typecast.ToPtr(money.NewFromInt(150)),
					operationAmount:        money.NewFromInt(50),
					updatedOperationAmount: money.NewFromInt(80),
				},
			},
			expected: expected{
				balance: "180.00",
			},
		},
		{
			desc: "should update balance with updating spending operation amount",
			args: args{
				actionType:    calculationActionTypeUpdate,
				operationType: model.OperationTypeSpending,
				opts: calculationOptions{
					balance:                typecast.ToPtr(money.NewFromInt(70)),
					operationAmount:        money.NewFromInt(30),
					updatedOperationAmount: money.NewFromInt(45),
				},
			},
			expected: expected{
				balance: "55.00",
			},
		},
		{
			desc: "should update balance's with updating transfers amounts without exchange rate",
			args: args{
				actionType:    calculationActionTypeUpdate,
				operationType: model.OperationTypeTransfer,
				opts: calculationOptions{
					balanceFrom:            typecast.ToPtr(money.NewFromInt(125)),
					balanceTo:              typecast.ToPtr(money.NewFromInt(125)),
					transferAmountOut:      typecast.ToPtr(money.NewFromInt(75)),
					transferAmountIn:       typecast.ToPtr(money.NewFromInt(75)),
					updatedOperationAmount: money.NewFromInt(100),
				},
			},
			expected: expected{
				balanceFrom: "100.00",
				balanceTo:   "150.00",
			},
		},
		{
			desc: "should update balance's thought transfer_in operation with exchange rate",
			args: args{
				actionType:    calculationActionTypeUpdate,
				operationType: model.OperationTypeTransferIn,
				opts: calculationOptions{
					balanceFrom:            typecast.ToPtr(money.NewFromInt(900)),
					balanceTo:              typecast.ToPtr(money.NewFromInt(250)),
					transferAmountOut:      typecast.ToPtr(money.NewFromInt(100)),
					transferAmountIn:       typecast.ToPtr(money.NewFromInt(150)),
					updatedOperationAmount: money.NewFromInt(120),
					exchangeRate:           typecast.ToPtr(money.NewFromFloat(1.5)),
				},
			},
			expected: expected{
				balanceFrom: "920.00",
				balanceTo:   "220.00",
			},
		},
		{
			desc: "should update balance's thought transfer_out operation with exchange rate",
			args: args{
				actionType:    calculationActionTypeUpdate,
				operationType: model.OperationTypeTransferOut,
				opts: calculationOptions{
					balanceFrom:            typecast.ToPtr(money.NewFromInt(900)),
					balanceTo:              typecast.ToPtr(money.NewFromInt(250)),
					transferAmountOut:      typecast.ToPtr(money.NewFromInt(100)),
					transferAmountIn:       typecast.ToPtr(money.NewFromInt(150)),
					updatedOperationAmount: money.NewFromInt(80),
					exchangeRate:           typecast.ToPtr(money.NewFromFloat(1.5)),
				},
			},
			expected: expected{
				balanceFrom: "920.00",
				balanceTo:   "220.00",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			calculateBalanceAmountBasedOnOperationType(tc.args.actionType, tc.args.operationType, tc.args.opts)

			if tc.args.opts.balance != nil && tc.expected.balance != "" {
				assert.Equal(t, tc.expected.balance, tc.args.opts.balance.StringFixed())
			}

			if tc.args.opts.balanceFrom != nil && tc.expected.balanceFrom != "" {
				assert.Equal(t, tc.expected.balanceFrom, tc.args.opts.balanceFrom.StringFixed())
			}
			if tc.args.opts.balanceTo != nil && tc.expected.balanceTo != "" {
				assert.Equal(t, tc.expected.balanceTo, tc.args.opts.balanceTo.StringFixed())
			}
		})
	}
}
