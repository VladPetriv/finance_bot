package service

import (
	"fmt"
	"math"
	"slices"

	"github.com/VladPetriv/finance_bot/internal/models"
)

const firstPage = 1

func isPaginationNeeded(text string) bool {
	return slices.Contains([]string{models.BotPreviousCommand, models.BotNextCommand}, text)
}

func calculateNextPage(text string, metadata map[string]any) int {
	currentPage := metadata[pageMetadataKey].(float64)
	switch {
	case text == models.BotPreviousCommand && currentPage > firstPage:
		currentPage--
	case text == models.BotNextCommand:
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

	maxPage := int(math.Ceil(float64(opts.totalCount) / float64(opts.maxPerKeyboard)))
	switch {
	case opts.currentPage == maxPage: // Max page do not display Next button
		keyboard = append(keyboard, InlineKeyboardRow{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotPreviousCommand,
				},
			},
		})
	case opts.currentPage > firstPage: // Current page is not the first page and not last page, display both buttons
		keyboard = append(keyboard, InlineKeyboardRow{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotPreviousCommand,
				},
				{
					Text: models.BotNextCommand,
				},
			},
		})
	case opts.currentPage == firstPage: // First page, display Next button only
		keyboard = append(keyboard, InlineKeyboardRow{
			Buttons: []InlineKeyboardButton{
				{
					Text: models.BotNextCommand,
				},
			},
		})
	}

	return keyboard, nil
}
