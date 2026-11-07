package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/stretchr/testify/assert"
)

func TestBuildOperationPicker(t *testing.T) {
	t.Parallel()

	type precondition struct {
		opts buildOperationPickerOptions
	}

	type expected struct {
		message  string
		keyboard []InlineKeyboardRow
	}

	const updateHeader = "Choose operation to update:"

	testCases := [...]struct {
		desc         string
		precondition *precondition
		expected     *expected
	}{
		{
			desc: "it should render five numbered blocks and a number row with Next only on the first page",
			precondition: &precondition{
				opts: buildOperationPickerOptions{
					header:     updateHeader,
					operations: pickerBlocks(1, 5),
					totalCount: 12,
					page:       1,
				},
			},
			expected: &expected{
				message:  pickerMessage(updateHeader, pickerBlocks(1, 5)),
				keyboard: pickerKeyboard(pickerBlocks(1, 5), model.BotNextCommand),
			},
		},
		{
			desc: "it should render Previous and Next on a middle page and restart numbering at 1",
			precondition: &precondition{
				opts: buildOperationPickerOptions{
					header:     updateHeader,
					operations: pickerBlocks(6, 5),
					totalCount: 12,
					page:       2,
				},
			},
			expected: &expected{
				message:  pickerMessage(updateHeader, pickerBlocks(6, 5)),
				keyboard: pickerKeyboard(pickerBlocks(6, 5), model.BotPreviousCommand, model.BotNextCommand),
			},
		},
		{
			desc: "it should render fewer number buttons and Previous only on a partial last page",
			precondition: &precondition{
				opts: buildOperationPickerOptions{
					header:     updateHeader,
					operations: pickerBlocks(11, 2),
					totalCount: 12,
					page:       3,
				},
			},
			expected: &expected{
				message:  pickerMessage(updateHeader, pickerBlocks(11, 2)),
				keyboard: pickerKeyboard(pickerBlocks(11, 2), model.BotPreviousCommand),
			},
		},
		{
			desc: "it should render one number button and no pagination row for a single operation",
			precondition: &precondition{
				opts: buildOperationPickerOptions{
					header:     updateHeader,
					operations: pickerBlocks(1, 1),
					totalCount: 1,
					page:       1,
				},
			},
			expected: &expected{
				message:  pickerMessage(updateHeader, pickerBlocks(1, 1)),
				keyboard: pickerKeyboard(pickerBlocks(1, 1)),
			},
		},
		{
			desc: "it should render the exchange rate only for the operation that has one",
			precondition: &precondition{
				opts: buildOperationPickerOptions{
					header: updateHeader,
					operations: []model.OperationBlockData{
						pickerBlock(1, "rate-op", "1.25"),
						pickerBlock(2, "plain-op", ""),
					},
					totalCount: 2,
					page:       1,
				},
			},
			expected: &expected{
				message: pickerMessage(updateHeader, []model.OperationBlockData{
					pickerBlock(1, "rate-op", "1.25"),
					pickerBlock(2, "plain-op", ""),
				}),
				keyboard: pickerKeyboard([]model.OperationBlockData{
					pickerBlock(1, "rate-op", "1.25"),
					pickerBlock(2, "plain-op", ""),
				}),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			message, keyboard := buildOperationPicker(tc.precondition.opts)

			assert.Equal(t, tc.expected.message, message)
			assert.Equal(t, tc.expected.keyboard, keyboard)
		})
	}
}

func pickerBlocks(firstSeq, count int) []model.OperationBlockData {
	blocks := make([]model.OperationBlockData, 0, count)
	for i := 0; i < count; i++ {
		blocks = append(blocks, pickerBlock(firstSeq+i, fmt.Sprintf("op-%d", firstSeq+i), ""))
	}

	return blocks
}

func pickerBlock(seq int, id, exchangeRate string) model.OperationBlockData {
	return model.OperationBlockData{
		Operation: model.Operation{
			ID:           id,
			Type:         model.OperationTypeSpending,
			Amount:       money.NewFromInt(int64(seq)),
			Description:  fmt.Sprintf("description %d", seq),
			ExchangeRate: exchangeRate,
			CreatedAt:    time.Date(2026, 10, 5, 14, 30, 0, 0, time.UTC),
		},
		CategoryTitle:  "Food",
		CurrencySymbol: "$",
	}
}

func pickerMessage(header string, blocks []model.OperationBlockData) string {
	message := header + "\n\n"
	for i, block := range blocks {
		message += fmt.Sprintf("*%d.*\n", i+1) + model.BuildOperationBlock(block)
	}

	return message + model.OperationBlockSeparator
}

func pickerKeyboard(blocks []model.OperationBlockData, navigation ...string) []InlineKeyboardRow {
	numbers := make([]InlineKeyboardButton, 0, len(blocks))
	for i, block := range blocks {
		numbers = append(numbers, InlineKeyboardButton{
			Text: fmt.Sprintf("%d", i+1),
			Data: block.Operation.ID,
		})
	}

	keyboard := []InlineKeyboardRow{
		{
			Buttons: numbers,
		},
	}
	if len(navigation) == 0 {
		return keyboard
	}

	buttons := make([]InlineKeyboardButton, 0, len(navigation))
	for _, text := range navigation {
		buttons = append(buttons, InlineKeyboardButton{
			Text: text,
		})
	}

	return append(keyboard, InlineKeyboardRow{
		Buttons: buttons,
	})
}
