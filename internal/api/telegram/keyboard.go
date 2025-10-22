package telegram

import (
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func (t *telegramMessenger) createKeyboard(rows []service.KeyboardRow) *telego.ReplyKeyboardMarkup {
	var convertedRows [][]telego.KeyboardButton

	for _, r := range rows {
		var buttons []telego.KeyboardButton

		for _, b := range r.Buttons {
			buttons = append(buttons, telegoutil.KeyboardButton(b))
		}

		convertedRows = append(convertedRows, buttons)
	}

	keyboard := telegoutil.Keyboard(convertedRows...).WithResizeKeyboard()

	return keyboard
}

const maxButtonsPerMessage = 100

func (t *telegramMessenger) createInlineKeyboard(rows []service.InlineKeyboardRow) []*telego.InlineKeyboardMarkup {
	convertedRows := make([][]telego.InlineKeyboardButton, 0)

	var totalButtonsCount int

	for _, r := range rows {
		var buttons []telego.InlineKeyboardButton

		for _, b := range r.Buttons {
			totalButtonsCount++

			inlineKeyboardButton := telegoutil.
				InlineKeyboardButton(b.Text).
				WithCallbackData(b.Text)

			if b.Data != "" {
				inlineKeyboardButton = inlineKeyboardButton.WithCallbackData(b.Data)
			}

			buttons = append(buttons, inlineKeyboardButton)
		}

		convertedRows = append(convertedRows, buttons)
	}

	if totalButtonsCount <= maxButtonsPerMessage {
		return []*telego.InlineKeyboardMarkup{telegoutil.InlineKeyboard(convertedRows...)}
	}

	return splitInlineKeyboardRows(convertedRows, maxButtonsPerMessage)
}

func splitInlineKeyboardRows(convertedRows [][]telego.InlineKeyboardButton, maxButtonsPerMessage int) []*telego.InlineKeyboardMarkup {
	var (
		buttonsCount        int
		lastProcessedRowIdx int
	)

	result := make([]*telego.InlineKeyboardMarkup, 0, 2)

	for rowIdx, row := range convertedRows {
		for btnIdx := range row {
			if buttonsCount == maxButtonsPerMessage {
				splitIndex := rowIdx

				// Ensure the split doesn't occur in the middle of a row
				if btnIdx > 0 {
					splitIndex--
				}

				result = append(result, telegoutil.InlineKeyboard(convertedRows[lastProcessedRowIdx:splitIndex]...))

				// Reset counters
				buttonsCount = 0
				lastProcessedRowIdx = splitIndex
			}

			buttonsCount++
		}
	}

	// Append the remaining buttons
	if lastProcessedRowIdx < len(convertedRows) {
		result = append(result, telegoutil.InlineKeyboard(convertedRows[lastProcessedRowIdx:]...))
	}

	return result
}
