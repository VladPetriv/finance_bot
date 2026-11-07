package service

import (
	"context"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOperationsHistoryKeyboard(t *testing.T) {
	t.Parallel()

	type precondition struct {
		balances   []model.Balance
		operations []model.Operation
	}

	type expected struct {
		message string
	}

	createdAt := time.Date(2026, 10, 5, 14, 30, 0, 0, time.UTC)
	usd := model.Currency{
		Symbol: "$",
	}
	eur := model.Currency{
		Symbol: "€",
	}

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should render balance and period lines and operation blocks without balance line for a single balance",
			precondition: &precondition{
				balances: []model.Balance{
					{
						ID:       "balance-1",
						Name:     "Cash",
						Amount:   money.NewFromInt(100),
						Currency: usd,
					},
				},
				operations: []model.Operation{
					{
						ID:          "operation-1",
						CategoryID:  "category-1",
						BalanceID:   "balance-1",
						Type:        model.OperationTypeSpending,
						Amount:      money.NewFromInt(10),
						Description: "coffee",
						CreatedAt:   createdAt,
					},
					{
						ID:           "operation-2",
						CategoryID:   "category-1",
						BalanceID:    "balance-1",
						Type:         model.OperationTypeTransferIn,
						Amount:       money.NewFromFloat(20.5),
						Description:  "transfer",
						ExchangeRate: "1.2",
						CreatedAt:    createdAt,
					},
				},
			},
			expected: &expected{
				message: "💰 *Balance:* 100.00$\n" +
					"📅 *Period:* day\n\n" +
					"━━━━━━━━━━━━━━━\n" +
					"📌 *Operation:* " + operationTypeLine(model.OperationTypeSpending) + "\n" +
					"📝 *Description:* coffee\n" +
					"📂 *Category:* Food\n" +
					"💰 *Amount:* 10.00$\n" +
					"🕐 *Date:* Mon Oct  5 14:30:00 2026\n" +
					"━━━━━━━━━━━━━━━\n" +
					"📌 *Operation:* " + operationTypeLine(model.OperationTypeTransferIn) + "\n" +
					"📝 *Description:* transfer\n" +
					"📂 *Category:* Food\n" +
					"💰 *Amount:* 20.50$\n" +
					"💱 *Exchange Rate:* 1.2\n" +
					"🕐 *Date:* Mon Oct  5 14:30:00 2026\n" +
					"━━━━━━━━━━━━━━━",
			},
		},
		{
			desc: "it should render every balance line and a balance line per operation for several balances",
			precondition: &precondition{
				balances: []model.Balance{
					{
						ID:       "balance-1",
						Name:     "Cash",
						Amount:   money.NewFromInt(100),
						Currency: usd,
					},
					{
						ID:       "balance-2",
						Name:     "Card",
						Amount:   money.NewFromInt(50),
						Currency: eur,
					},
				},
				operations: []model.Operation{
					{
						ID:          "operation-1",
						CategoryID:  "category-1",
						BalanceID:   "balance-2",
						Type:        model.OperationTypeIncoming,
						Amount:      money.NewFromInt(7),
						Description: "salary",
						CreatedAt:   createdAt,
					},
				},
			},
			expected: &expected{
				message: "💰 *Cash:* 100.00$\n" +
					"💰 *Card:* 50.00€\n" +
					"📅 *Period:* day\n\n" +
					"━━━━━━━━━━━━━━━\n" +
					"📌 *Operation:* " + operationTypeLine(model.OperationTypeIncoming) + "\n" +
					"📝 *Description:* salary\n" +
					"📂 *Category:* Food\n" +
					"💰 *Amount:* 7.00€\n" +
					"🏦 *Balance:* Card\n" +
					"🕐 *Date:* Mon Oct  5 14:30:00 2026\n" +
					"━━━━━━━━━━━━━━━",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			h := handlerService{
				logger: logger.NewDummy(),
				stores: Stores{
					Operation: fakeOperationStore{
						operations: tc.precondition.operations,
					},
					Category: fakeCategoryStore{},
				},
			}

			message, keyboard, err := h.getOperationsHistoryKeyboard(t.Context(), getOperationsHistoryKeyboardOptions{
				balances:       tc.precondition.balances,
				creationPeriod: model.CreationPeriodDay,
				page:           firstPage,
				location:       time.UTC,
			})
			require.NoError(t, err)

			assert.Equal(t, tc.expected.message, message)
			assert.Empty(t, keyboard)
		})
	}
}

func operationTypeLine(t model.OperationType) string {
	emoji, label := model.GetOperationTypeLabel(t)

	return emoji + " " + label
}

type fakeOperationStore struct {
	OperationStore

	operations []model.Operation
}

func (s fakeOperationStore) Count(_ context.Context, _ ListOperationsFilter) (int, error) {
	return len(s.operations), nil
}

func (s fakeOperationStore) List(_ context.Context, _ ListOperationsFilter) ([]model.Operation, error) {
	return append([]model.Operation(nil), s.operations...), nil
}

type fakeCategoryStore struct {
	CategoryStore
}

func (fakeCategoryStore) Get(_ context.Context, filter GetCategoryFilter) (*model.Category, error) {
	return &model.Category{
		ID:    filter.ID,
		Title: "Food",
	}, nil
}
