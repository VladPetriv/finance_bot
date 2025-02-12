package models

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/VladPetriv/finance_bot/pkg/money"
)

// StatisticsMessageBuilder is responsible for building a formatted message containing
// financial statistics including balance, operations, and category breakdowns.
type StatisticsMessageBuilder struct {
	balance    *Balance
	operations []Operation
	categories []Category

	buffer strings.Builder
}

// NewStatisticsMessageBuilder creates a new instance of StatisticsMessageBuilder with the provided
// balance, operations, and categories data
func NewStatisticsMessageBuilder(balance *Balance, operations []Operation, categories []Category) *StatisticsMessageBuilder {
	return &StatisticsMessageBuilder{
		balance:    balance,
		operations: operations,
		categories: categories,
	}
}

// Build generates a formatted message string containing financial statistics.
// It includes balance information, period details, and breakdowns of operations by type and category.
func (b *StatisticsMessageBuilder) Build() (string, error) {
	stats, err := calculateOperationsStatistics(b.operations)
	if err != nil {
		return "", fmt.Errorf("error calculating statistics: %w", err)
	}

	return b.
		addHeader().
		addPeriod().
		addOperationsAndCategoriesStatistics(stats).buffer.String(), nil
}

func (b *StatisticsMessageBuilder) addHeader() *StatisticsMessageBuilder {
	balanceAmount, _ := money.NewFromString(b.balance.Amount)

	template := `📊 Balance Statistics: *%s*
💰 Current Balance: %s

`
	b.buffer.WriteString(fmt.Sprintf(template, b.balance.Name, formatAmount(balanceAmount, b.balance.Currency)))

	return b
}

func (b *StatisticsMessageBuilder) addPeriod() *StatisticsMessageBuilder {
	b.buffer.WriteString(formatPeriod())
	b.buffer.WriteString(`

`)
	return b
}

func (b *StatisticsMessageBuilder) addOperationsAndCategoriesStatistics(stats *operationsStatistics) *StatisticsMessageBuilder {
	headerTemplate := `📈 Summary:
`
	b.buffer.WriteString(headerTemplate)

	incomingOperationTemplate := `📥 Incoming Operations: %s *(%d)*
`
	b.buffer.WriteString(
		fmt.Sprintf(
			incomingOperationTemplate,
			formatAmount(stats.IncomingTotal, b.balance.Currency),
			stats.IncomingCount,
		),
	)

	incomingCategoriesStatistics, _ := calculateCategoryStatistics(stats.IncomingTotal, stats.OperationsByType[OperationTypeIncoming], b.categories)
	b.buffer.WriteString(b.buildCategoriesStatisticsMessage(incomingCategoriesStatistics))

	spendingOperationTemplate := `💸 Spending Operations: %s *(%d)*
`
	b.buffer.WriteString(
		fmt.Sprintf(
			spendingOperationTemplate,
			formatAmount(stats.SpendingTotal, b.balance.Currency),
			stats.SpendingCount,
		),
	)

	spendingCategoriesStatistics, _ := calculateCategoryStatistics(stats.SpendingTotal, stats.OperationsByType[OperationTypeSpending], b.categories)
	b.buffer.WriteString(b.buildCategoriesStatisticsMessage(spendingCategoriesStatistics))

	transferOperationTemplate := `🔄 Transfers Operations *(%d)*:
		 ➡️ In: %s *(%d)*
			⬅️ Out: %s *(%d)*
	`
	totalTransfers := stats.TransferInCount + stats.TransferOutCount
	b.buffer.WriteString(
		fmt.Sprintf(
			transferOperationTemplate,
			totalTransfers,

			formatAmount(stats.TransferInTotal, b.balance.Currency),
			stats.TransferInCount,

			formatAmount(stats.TransferOutTotal, b.balance.Currency),
			stats.TransferOutCount,
		),
	)

	return b
}

func (b *StatisticsMessageBuilder) buildCategoriesStatisticsMessage(categoriesStat []categoryStatistics) string {
	builder := strings.Builder{}

	templateForFirstElement := `			- %s: %s *(%s%%)*`
	regularTemplate := `
			- %s: %s *(%s%%)*`

	for index, category := range categoriesStat {
		template := regularTemplate

		if index == 0 {
			template = templateForFirstElement
		}

		builder.WriteString(
			fmt.Sprintf(
				template,
				category.Title,
				formatAmount(category.Amount, b.balance.Currency),
				category.Percentage.StringFixed(),
			),
		)
	}

	if len(categoriesStat) > 0 {
		builder.WriteString(`

`)
	}

	return builder.String()
}

func formatAmount(amount money.Money, currency string) string {
	return formatInlineCode(amount.StringFixed() + currency)
}

func formatInlineCode(s string) string {
	return fmt.Sprintf("`%s`", s)
}

const dateFormat = "02 Jan 2006"

func formatPeriod() string {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	template := "📅 Period: _%s - %s_"
	return fmt.Sprintf(template, startOfMonth.Format(dateFormat), now.Format(dateFormat))
}

type operationsStatistics struct {
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

func calculateOperationsStatistics(operations []Operation) (*operationsStatistics, error) {
	stats := &operationsStatistics{
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

type categoryStatistics struct {
	Title      string
	Amount     money.Money
	Percentage money.Money
}

func calculateCategoryStatistics(totalAmount money.Money, operations []Operation, categories []Category) ([]categoryStatistics, error) {
	categoryStats := make(map[string]*categoryStatistics)

	for _, category := range categories {
		categoryStats[category.Title] = &categoryStatistics{
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

	result := make([]categoryStatistics, 0, len(categoryStats))
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
