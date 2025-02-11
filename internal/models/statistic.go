package models

import (
	"fmt"
	"sort"

	"github.com/VladPetriv/finance_bot/pkg/money"
)

// OperationsStatistics represents aggregated statistics for financial operations,
// including counts and totals for different operation types
type OperationsStatistics struct {
	IncomingCount    int
	SpendingCount    int
	TransferInCount  int
	TransferOutCount int

	IncomingTotal    money.Money
	SpendingTotal    money.Money
	TransferInTotal  money.Money
	TransferOutTotal money.Money

	OperationsByType map[OperationType][]Operation
}

// CalculateOperationsStatistics processes a slice of operations and returns aggregated statistics
// for different operation types (incoming, spending, transfers)
func CalculateOperationsStatistics(operations []Operation) (*OperationsStatistics, error) {
	stats := &OperationsStatistics{
		IncomingTotal:    money.Zero,
		SpendingTotal:    money.Zero,
		TransferInTotal:  money.Zero,
		TransferOutTotal: money.Zero,
		OperationsByType: make(map[OperationType][]Operation),
	}

	for _, operation := range operations {
		amount, err := money.NewFromString(operation.Amount)
		if err != nil {
			return nil, fmt.Errorf("invalid operation amount: %w", err)
		}

		switch operation.Type {
		case OperationTypeIncoming:
			stats.IncomingCount++
			stats.IncomingTotal.Inc(amount)

		case OperationTypeSpending:
			stats.SpendingCount++
			stats.SpendingTotal.Inc(amount)

		case OperationTypeTransferIn:
			stats.TransferInCount++
			stats.TransferInTotal.Inc(amount)
		case OperationTypeTransferOut:
			stats.TransferOutCount++
			stats.TransferOutTotal.Inc(amount)
		}

		stats.OperationsByType[operation.Type] = append(stats.OperationsByType[operation.Type], operation)
	}

	return stats, nil
}

// CategoryStatistics represents statistics for a single spending category,
// including total amount and percentage of total spending
type CategoryStatistics struct {
	Title      string
	Amount     money.Money
	Percentage money.Money
}

// CalculateCategoryStatistics calculates spending statistics by category, including
// the amount and percentage of total spending for each category
func CalculateCategoryStatistics(totalAmount money.Money, operations []Operation, categories []Category) ([]CategoryStatistics, error) {
	categoryStats := make(map[string]*CategoryStatistics)

	for _, category := range categories {
		categoryStats[category.Title] = &CategoryStatistics{
			Title:      category.Title,
			Amount:     money.Zero,
			Percentage: money.Zero,
		}
	}

	for _, operation := range operations {
		for _, category := range categories {
			if operation.CategoryID == category.ID {
				amount, err := money.NewFromString(operation.Amount)
				if err != nil {
					return nil, fmt.Errorf("invalid operation amount: %w", err)
				}

				stats := categoryStats[category.Title]
				stats.Amount.Inc(amount)
			}
		}
	}

	result := make([]CategoryStatistics, 0, len(categoryStats))
	for _, stats := range categoryStats {
		if stats.Amount == money.Zero {
			continue
		}

		hundred := money.NewFromInt(100)
		percentageAmount := stats.Amount
		percentageAmount.Mul(hundred)
		percentageAmount.Div(totalAmount)

		stats.Percentage = percentageAmount

		result = append(result, *stats)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Percentage.GreaterThan(result[j].Percentage)
	})

	return result, nil
}
