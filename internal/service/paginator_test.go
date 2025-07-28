package service

import (
	"testing"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_paginateInlineKeyboard(t *testing.T) {
	t.Parallel()

	type args struct {
		opts      inlineKeyboardPaginatorOptions
		getValues getValuesFunc[models.Currency]
	}

	testCases := [...]struct {
		desc     string
		args     args
		expected []InlineKeyboardRow
	}{
		{
			desc: "received inline keyboard on first page with total count=1",
			args: args{
				opts: inlineKeyboardPaginatorOptions{
					totalCount:     1,
					maxPerKeyboard: 1,
					maxPerRow:      1,
					currentPage:    firstPage,
				},
				getValues: func() ([]models.Currency, error) {
					return []models.Currency{
						{
							ID:     "1",
							Name:   "USD",
							Code:   "USD",
							Symbol: "$",
						},
					}, nil
				},
			},
			expected: []InlineKeyboardRow{
				{
					Buttons: []InlineKeyboardButton{
						{
							Text: "USD (USD)",
							Data: "1",
						},
					},
				},
			},
		},
		{
			desc: "received inline keyboard on max page",
			args: args{
				opts: inlineKeyboardPaginatorOptions{
					totalCount:     4,
					maxPerKeyboard: 2,
					maxPerRow:      1,
					currentPage:    2,
				},
				getValues: func() ([]models.Currency, error) {
					return []models.Currency{
						{
							ID:     "1",
							Name:   "USD",
							Code:   "USD",
							Symbol: "$",
						},
						{
							ID:     "2",
							Name:   "EUR",
							Code:   "EUR",
							Symbol: "€",
						},
					}, nil
				},
			},
			expected: []InlineKeyboardRow{
				{
					Buttons: []InlineKeyboardButton{
						{
							Text: "USD (USD)",
							Data: "1",
						},
					},
				},
				{
					Buttons: []InlineKeyboardButton{
						{
							Text: "EUR (EUR)",
							Data: "2",
						},
					},
				},
				{
					Buttons: []InlineKeyboardButton{
						{
							Text: models.BotPreviousCommand,
						},
					},
				},
			},
		},
		{
			desc: "received inline keyboard in the middle of pagination current page is 2 and total count 4",
			args: args{
				opts: inlineKeyboardPaginatorOptions{
					totalCount:     4,
					maxPerKeyboard: 1,
					maxPerRow:      1,
					currentPage:    2,
				},
				getValues: func() ([]models.Currency, error) {
					return []models.Currency{
						{
							ID:     "2",
							Name:   "EUR",
							Code:   "EUR",
							Symbol: "€",
						},
					}, nil
				},
			},
			expected: []InlineKeyboardRow{
				{
					Buttons: []InlineKeyboardButton{
						{
							Text: "EUR (EUR)",
							Data: "2",
						},
					},
				},
				{
					Buttons: []InlineKeyboardButton{
						{
							Text: models.BotPreviousCommand,
						},
						{
							Text: models.BotNextCommand,
						},
					},
				},
			},
		},
		{
			desc: "received inline keyboard on first page with total count = 2",
			args: args{
				opts: inlineKeyboardPaginatorOptions{
					totalCount:     2,
					maxPerKeyboard: 1,
					maxPerRow:      2,
					currentPage:    firstPage,
				},
				getValues: func() ([]models.Currency, error) {
					return []models.Currency{
						{
							ID:     "2",
							Name:   "EUR",
							Code:   "EUR",
							Symbol: "€",
						},
					}, nil
				},
			},
			expected: []InlineKeyboardRow{
				{
					Buttons: []InlineKeyboardButton{
						{
							Text: "EUR (EUR)",
							Data: "2",
						},
					},
				},
				{
					Buttons: []InlineKeyboardButton{
						{
							Text: models.BotNextCommand,
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := paginateInlineKeyboard(tc.args.opts, tc.args.getValues)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
