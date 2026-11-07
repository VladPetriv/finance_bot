package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
)

const operationPickerMissingCategoryTitle = "—"

type buildOperationPickerOptions struct {
	header     string
	operations []model.OperationBlockData
	totalCount int
	page       int
}

func buildOperationPicker(opts buildOperationPickerOptions) (string, []InlineKeyboardRow) {
	var message strings.Builder
	message.WriteString(opts.header + "\n\n")

	numberButtons := make([]InlineKeyboardButton, 0, len(opts.operations))
	for i, block := range opts.operations {
		number := strconv.Itoa(i + 1)
		message.WriteString(fmt.Sprintf("*%s.*\n", number))
		message.WriteString(model.BuildOperationBlock(block))
		numberButtons = append(numberButtons, InlineKeyboardButton{
			Text: number,
			Data: block.Operation.ID,
		})
	}
	message.WriteString(model.OperationBlockSeparator)

	keyboard := []InlineKeyboardRow{
		{
			Buttons: numberButtons,
		},
	}
	if opts.totalCount > 1 {
		keyboard = handlePaginationInlineKeyboardButtons(keyboard, opts.page, calculateMaxPage(opts.totalCount, operationsPerKeyboard))
	}

	return message.String(), keyboard
}

type getOperationPickerOptions struct {
	header   string
	balance  model.Balance
	page     int
	location *time.Location
}

func (h handlerService) getOperationPicker(ctx context.Context, opts getOperationPickerOptions) (string, []InlineKeyboardRow, error) {
	logger := h.logger.With().Str("name", "handlerService.getOperationPicker").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationsCount, err := h.stores.Operation.Count(ctx, ListOperationsFilter{
		BalanceID: opts.balance.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("count operations")
		return "", nil, fmt.Errorf("count operations: %w", err)
	}
	if operationsCount == 0 {
		logger.Info().Any("balanceID", opts.balance.ID).Msg("operations not found")
		return "", nil, ErrOperationsNotFound
	}

	operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
		BalanceID:            opts.balance.ID,
		OrderByCreatedAtDesc: true,
		Pagination: &Pagination{
			Limit: operationsPerKeyboard,
			Page:  opts.page,
		},
		Location: opts.location,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list operations from store")
		return "", nil, fmt.Errorf("list operations from store: %w", err)
	}
	if len(operations) == 0 {
		logger.Info().Msg("operations not found")
		return "", nil, ErrOperationsNotFound
	}
	convertOperationsToLocation(operations, opts.location)

	blocks := make([]model.OperationBlockData, 0, len(operations))
	for _, operation := range operations {
		category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
			ID: operation.CategoryID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get category from store")
			return "", nil, fmt.Errorf("get category from store: %w", err)
		}

		categoryTitle := operationPickerMissingCategoryTitle
		if category != nil {
			categoryTitle = category.Title
		}

		blocks = append(blocks, model.OperationBlockData{
			Operation:      operation,
			CategoryTitle:  categoryTitle,
			CurrencySymbol: opts.balance.GetCurrency().Symbol,
		})
	}

	message, keyboard := buildOperationPicker(buildOperationPickerOptions{
		header:     opts.header,
		operations: blocks,
		totalCount: operationsCount,
		page:       opts.page,
	})

	return message, keyboard, nil
}
