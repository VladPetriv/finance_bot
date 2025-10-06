package service

import (
	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/money"
)

type calculationActionType string

const (
	calculationActionTypeCreate calculationActionType = "create"
	calculationActionTypeUpdate calculationActionType = "update"
)

type calculationOptions struct {
	// Used for simple balance operations
	balance *money.Money

	// Used for transfer operations
	balanceFrom *money.Money
	balanceTo   *money.Money

	// Used
	transferAmountIn  *money.Money
	transferAmountOut *money.Money

	operationAmount money.Money

	updatedOperationAmount money.Money

	exchangeRate *money.Money
}

func calculateBalanceAmountBasedOnOperationType(actionType calculationActionType, operationType model.OperationType, opts calculationOptions) {
	// Check if required amounts are present
	switch operationType {
	case model.OperationTypeIncoming, model.OperationTypeSpending:
		if opts.balance == nil {
			return
		}
	case model.OperationTypeTransfer, model.OperationTypeTransferIn, model.OperationTypeTransferOut:
		if opts.balanceFrom == nil || opts.balanceTo == nil {
			return
		}

		if actionType == calculationActionTypeUpdate && (opts.transferAmountIn == nil || opts.transferAmountOut == nil) {
			return
		}
	}

	switch actionType {
	case calculationActionTypeCreate:
		switch operationType {
		case model.OperationTypeIncoming:
			calculateIncomingOperation(opts.balance, opts.operationAmount)
		case model.OperationTypeSpending:
			calculateSpendingOperation(opts.balance, opts.operationAmount)
		case model.OperationTypeTransfer, model.OperationTypeTransferIn, model.OperationTypeTransferOut:
			calculateTranferOperation(calculateTransferOperationOptions{
				balanceFrom:     opts.balanceFrom,
				balanceTo:       opts.balanceTo,
				operationAmount: opts.operationAmount,
				exchangeRate:    opts.exchangeRate,
			})
		}
	case calculationActionTypeUpdate:
		switch operationType {
		case model.OperationTypeIncoming:
			calculateUpdatedIncomingOperation(opts.balance, opts.operationAmount, opts.updatedOperationAmount)
		case model.OperationTypeSpending:
			calculateUpdatedSpendingOperation(opts.balance, opts.operationAmount, opts.updatedOperationAmount)
		case model.OperationTypeTransfer, model.OperationTypeTransferIn, model.OperationTypeTransferOut:
			calculateUpdatedTranferOperation(calculateTransferOperationOptions{
				operationType:          operationType,
				balanceFrom:            opts.balanceFrom,
				balanceTo:              opts.balanceTo,
				transferAmountIn:       opts.transferAmountIn,
				transferAmountOut:      opts.transferAmountOut,
				updatedOperationAmount: opts.updatedOperationAmount,
				exchangeRate:           opts.exchangeRate,
			})

		}
	}
}

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

	// Used for update action only
	transferAmountIn  *money.Money
	transferAmountOut *money.Money

	operationAmount        money.Money
	updatedOperationAmount money.Money

	exchangeRate *money.Money
}

func calculateTranferOperation(opts calculateTransferOperationOptions) {
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

	// Handle transfer amount update with exchange rate
	case opts.exchangeRate != nil:
		opts.balanceFrom.Inc(*opts.transferAmountOut)
		opts.balanceTo.Sub(*opts.transferAmountIn)

		switch opts.operationType {
		case model.OperationTypeTransferIn:
			updatedTransferOutAmount, _ := money.NewFromString(opts.updatedOperationAmount.StringFixed())
			updatedTransferOutAmount.Div(*opts.exchangeRate)
			opts.balanceFrom.Sub(updatedTransferOutAmount)
			opts.transferAmountOut = &updatedTransferOutAmount

			opts.balanceTo.Inc(opts.updatedOperationAmount)
			opts.transferAmountIn = &opts.updatedOperationAmount
		case model.OperationTypeTransferOut:
			updatedTransferInAmount, _ := money.NewFromString(opts.updatedOperationAmount.StringFixed())
			updatedTransferInAmount.Mul(*opts.exchangeRate)
			opts.balanceTo.Inc(updatedTransferInAmount)
			opts.transferAmountIn = &updatedTransferInAmount

			opts.balanceFrom.Sub(opts.updatedOperationAmount)
			opts.transferAmountOut = &opts.updatedOperationAmount
		}
	}
}
