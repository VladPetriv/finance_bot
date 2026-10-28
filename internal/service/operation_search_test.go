package service

import (
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBuildSearchOperationsFilter(t *testing.T) {
	t.Parallel()

	type precondition struct {
		metadata model.Metadata
	}

	type expected struct {
		filter ListOperationsFilter
	}

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should map every metadata value to the filter and make the end date inclusive",
			precondition: &precondition{
				metadata: model.Metadata{
					model.SearchQueryMetadataKey:         "coffee",
					model.SearchOperationTypeMetadataKey: "transfer",
					model.SearchCategoryIDsMetadataKey:   "category-1,category-2",
					model.SearchAmountFromMetadataKey:    "10",
					model.SearchAmountToMetadataKey:      "20.5",
					model.SearchDateFromMetadataKey:      "2026-10-01",
					model.SearchDateToMetadataKey:        "2026-10-31",
				},
			},
			expected: &expected{
				filter: ListOperationsFilter{
					BalanceIDs:  []string{"balance-1", "balance-2"},
					Location:    time.UTC,
					Description: "coffee",
					Types:       []model.OperationType{model.OperationTypeTransferIn, model.OperationTypeTransferOut},
					CategoryIDs: []string{"category-1", "category-2"},
					AmountFrom:  "10",
					AmountTo:    "20.5",
					BetweenFilter: &BetweenFilter{
						From: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			desc: "it should leave the filter empty when nothing is selected",
			precondition: &precondition{
				metadata: model.Metadata{},
			},
			expected: &expected{
				filter: ListOperationsFilter{
					BalanceIDs: []string{"balance-1", "balance-2"},
					Location:   time.UTC,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual := buildSearchOperationsFilter(tc.precondition.metadata, []string{"balance-1", "balance-2"}, time.UTC)

			assert.Equal(t, tc.expected.filter, actual)
		})
	}
}

func TestGetSearchCalendarKeyboard(t *testing.T) {
	t.Parallel()

	keyboard := getSearchCalendarKeyboard(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC))

	assert.Len(t, keyboard, 7)
	for _, row := range keyboard[1:] {
		assert.Len(t, row.Buttons, 7)
	}
	assert.Equal(t, "cal:2026-09", keyboard[0].Buttons[0].Data)
	assert.Equal(t, "cal:2026-11", keyboard[0].Buttons[2].Data)
	assert.Equal(t, searchCalendarIgnoreData, keyboard[2].Buttons[0].Data)
	assert.Equal(t, "2026-10-01", keyboard[2].Buttons[3].Data)
	assert.Equal(t, "2026-10-31", keyboard[6].Buttons[5].Data)
}
