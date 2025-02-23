package telegram

import (
	"testing"

	"github.com/mymmrac/telego"
	"github.com/stretchr/testify/assert"
)

func TestSplitInlineKeyboardRows(t *testing.T) {

	type args struct {
		rows       [][]telego.InlineKeyboardButton
		maxButtons int
	}

	testCases := [...]struct {
		desc     string
		args     args
		expected int
	}{
		{
			desc: "No split needed (fewer buttons than limit)",
			args: args{
				rows:       [][]telego.InlineKeyboardButton{{createButton("A"), createButton("B")}},
				maxButtons: 2,
			},
			expected: 1,
		},
		{
			desc: "Exact limit (no split needed)",
			args: args{
				rows:       [][]telego.InlineKeyboardButton{{createButton("A"), createButton("B"), createButton("C")}},
				maxButtons: 3,
			},
			expected: 1,
		},
		{
			desc: "Split into two messages",
			args: args{
				rows: [][]telego.InlineKeyboardButton{
					{createButton("A"), createButton("B")},
					{createButton("C"), createButton("D")},
					{createButton("E")},
				},
				maxButtons: 3,
			},
			expected: 2,
		},
		{
			desc: "Split into multiple messages",
			args: args{
				rows: [][]telego.InlineKeyboardButton{
					{createButton("A"), createButton("B")},
					{createButton("C"), createButton("D")},
					{createButton("E"), createButton("F")},
					{createButton("G"), createButton("H")},
				},
				maxButtons: 2,
			},
			expected: 4,
		},
		{
			desc: "Single button per row, requiring multiple splits",
			args: args{
				rows: [][]telego.InlineKeyboardButton{
					{createButton("A")}, {createButton("B")}, {createButton("C")},
					{createButton("D")}, {createButton("E")}, {createButton("F")},
				},
				maxButtons: 2,
			},
			expected: 3,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			actual := splitInlineKeyboardRows(tc.args.rows, tc.args.maxButtons)
			assert.Equal(t, tc.expected, len(actual), "unexpected number of messages")
		})
	}
}

func createButton(text string) telego.InlineKeyboardButton {
	return telego.InlineKeyboardButton{Text: text}
}
