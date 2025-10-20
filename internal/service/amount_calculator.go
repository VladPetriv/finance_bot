package service

import (
	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/money"
)

func calculateIncomingOperation(balanceAmount *money.Money, incomeAmount money.Money) {
	balanceAmount.Inc(incomeAmount)
}

func calculateUpdatedIncomingOperation(balanceAmount *money.Money, initialAmount money.Money, updateAmount money.Money) {
	balanceAmount.Sub(initialAmount)
	balanceAmount.Inc(updateAmount)
}

func calculateSpendingOperation(balanceAmount *money.Money, spendingAmount money.Money) {
	balanceAmount.Sub(spendingAmount)
}

func calculateUpdatedSpendingOperation(balanceAmount *money.Money, initialAmount money.Money, updateAmount money.Money) {
	balanceAmount.Inc(initialAmount)
	balanceAmount.Sub(updateAmount)
}

type calculateTransferOperationOptions struct {
	operationType model.OperationType

	balanceFrom *money.Money
	balanceTo   *money.Money

	operationAmount money.Money
	exchangeRate    *money.Money

	// Used for update action only
	transferAmountIn       *money.Money
	transferAmountOut      *money.Money
	updatedOperationAmount money.Money
}

func calculateTransferOperation(opts calculateTransferOperationOptions) {
	switch {
	// Handle simply balance transfer
	case opts.exchangeRate == nil:
		opts.balanceFrom.Sub(opts.operationAmount)
		opts.balanceTo.Inc(opts.operationAmount)

	// Handle transfer with exchange rate
	case opts.exchangeRate != nil:
		opts.balanceFrom.Sub(opts.operationAmount)
		opts.operationAmount.Mul(*opts.exchangeRate)
		opts.balanceTo.Inc(opts.operationAmount)
	}
}

func calculateUpdatedTranferOperation(opts calculateTransferOperationOptions) {
	switch {
	// Handle simply balance transfer
	case opts.exchangeRate == nil:
		opts.balanceFrom.Inc(*opts.transferAmountOut)
		opts.balanceTo.Sub(*opts.transferAmountIn)

		opts.balanceFrom.Sub(opts.updatedOperationAmount)
		opts.balanceTo.Inc(opts.updatedOperationAmount)

		opts.transferAmountIn.Set(opts.updatedOperationAmount)
		opts.transferAmountOut.Set(opts.updatedOperationAmount)

	// Handle transfer amount update with exchange rate
	case opts.exchangeRate != nil:
		opts.balanceFrom.Inc(*opts.transferAmountOut)
		opts.balanceTo.Sub(*opts.transferAmountIn)

		switch opts.operationType {
		case model.OperationTypeTransferIn:
			updatedTransferOutAmount, _ := money.NewFromString(opts.updatedOperationAmount.StringFixed())
			updatedTransferOutAmount.Div(*opts.exchangeRate)
			opts.balanceFrom.Sub(updatedTransferOutAmount)
			opts.transferAmountOut.Set(updatedTransferOutAmount)

			opts.balanceTo.Inc(opts.updatedOperationAmount)
			opts.transferAmountIn.Set(opts.updatedOperationAmount)
		case model.OperationTypeTransferOut:
			updatedTransferInAmount, _ := money.NewFromString(opts.updatedOperationAmount.StringFixed())
			updatedTransferInAmount.Mul(*opts.exchangeRate)
			opts.balanceTo.Inc(updatedTransferInAmount)
			opts.transferAmountIn.Set(updatedTransferInAmount)

			opts.balanceFrom.Sub(opts.updatedOperationAmount)
			opts.transferAmountOut.Set(opts.updatedOperationAmount)
		}
	}
}
