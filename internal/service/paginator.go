package service

import (
	"fmt"
	"math"
	"slices"

	"github.com/VladPetriv/finance_bot/internal/model"
)

const firstPage = 1

func isPaginationNeeded(text string) bool {
	return slices.Contains([]string{model.BotPreviousCommand, model.BotNextCommand}, text)
}

func calculateNextPage(text string, metadata map[string]any) int {
	currentPage := metadata[pageMetadataKey].(float64)
	switch {
	case text == model.BotPreviousCommand && currentPage > firstPage:
		currentPage--
	case text == model.BotNextCommand:
		currentPage++
	}

	return int(currentPage)
}

type inlineKeyboardPaginatorOptions struct {
	totalCount     int
	maxPerKeyboard int
	maxPerRow      int
	currentPage    int
}

type getValuesFunc[T identifiable] func() ([]T, error)

func paginateInlineKeyboard[T identifiable](opts inlineKeyboardPaginatorOptions, getValues getValuesFunc[T]) ([]InlineKeyboardRow, error) {
	values, err := getValues()
	if err != nil {
		return nil, fmt.Errorf("get values for inline keyboard pagination: %w", err)
	}

	keyboard := make([]InlineKeyboardRow, 0, opts.maxPerKeyboard)

	var currentRow InlineKeyboardRow
	for index, value := range values {
		currentRow.Buttons = append(currentRow.Buttons, InlineKeyboardButton{
			Text: value.GetName(),
			Data: value.GetID(),
		})

		if len(currentRow.Buttons) == opts.maxPerRow || index == len(values)-1 {
			keyboard = append(keyboard, currentRow)
			currentRow = InlineKeyboardRow{}
		}
	}

	if opts.totalCount == 1 {
		return keyboard, nil
	}

	maxPage := calculateMaxPage(opts.totalCount, opts.maxPerKeyboard)
	keyboard = handlePaginationInlineKeyboardButtons(keyboard, opts.currentPage, maxPage)

	return keyboard, nil
}

type getMessageText func() (string, error)

func paginateTextUsingInlineKeybaord(opts inlineKeyboardPaginatorOptions, getMessage getMessageText) (string, []InlineKeyboardRow, error) {
	message, err := getMessage()
	if err != nil {
		return "", nil, fmt.Errorf("get message: %w", err)
	}

	keyboard := make([]InlineKeyboardRow, 0, opts.maxPerKeyboard)
	if opts.totalCount == 1 {
		return message, keyboard, nil
	}

	maxPage := calculateMaxPage(opts.totalCount, opts.maxPerKeyboard)
	keyboard = handlePaginationInlineKeyboardButtons(keyboard, opts.currentPage, maxPage)

	return message, keyboard, nil
}

func calculateMaxPage(totalCount, maxPerKeyboard int) int {
	return int(math.Ceil(float64(totalCount) / float64(maxPerKeyboard)))
}

func handlePaginationInlineKeyboardButtons(keyboard []InlineKeyboardRow, currentPage, maxPage int) []InlineKeyboardRow {
	switch {
	case currentPage == maxPage: // Max page do not display Next button
		keyboard = append(keyboard, InlineKeyboardRow{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotPreviousCommand,
				},
			},
		})
	case currentPage > firstPage: // Current page is not the first page and not last page, display both buttons
		keyboard = append(keyboard, InlineKeyboardRow{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotPreviousCommand,
				},
				{
					Text: model.BotNextCommand,
				},
			},
		})
	case currentPage == firstPage: // First page, display Next button only
		keyboard = append(keyboard, InlineKeyboardRow{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotNextCommand,
				},
			},
		})
	}

	return keyboard
}
