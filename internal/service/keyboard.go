package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
)

type identifiable interface {
	GetID() string
	GetName() string
}

func getRowKeyboardRows[T identifiable](data []T, elementLimitPerRow int, includeRowWithCancelButton bool) []KeyboardRow {
	keyboardRows := make([]KeyboardRow, 0)

	var currentRow KeyboardRow
	for i, entry := range data {
		currentRow.Buttons = append(currentRow.Buttons, entry.GetName())

		// When row is full or we're at the last data item, append row
		if len(currentRow.Buttons) == elementLimitPerRow || i == len(data)-1 {
			keyboardRows = append(keyboardRows, currentRow)
			currentRow = KeyboardRow{} // Reset current row
		}
	}

	if includeRowWithCancelButton {
		keyboardRows = append(keyboardRows, KeyboardRow{
			Buttons: []string{models.BotCancelCommand},
		})
	}

	return keyboardRows
}

const (
	operationsPerKeyboard    = 5
	operationsPerKeyboardRow = 1
)

type getOperationsKeyboardOptions struct {
	balanceID string
	page      int
}

func (h handlerService) getOperationsKeyboard(ctx context.Context, opts getOperationsKeyboardOptions) ([]InlineKeyboardRow, error) {
	logger := h.logger.With().Str("name", "handlerService.getOperationsKeyboard").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationsCount, err := h.stores.Operation.Count(ctx, ListOperationsFilter{
		BalanceID: opts.balanceID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("count operations")
		return nil, fmt.Errorf("count operations: %w", err)
	}
	if operationsCount == 0 {
		logger.Info().Any("balanceID", opts.balanceID).Msg("operations not found")
		return nil, ErrOperationsNotFound
	}

	keyboard, err := paginateInlineKeyboard(
		inlineKeyboardPaginatorOptions{
			totalCount:     operationsCount,
			maxPerKeyboard: operationsPerKeyboard,
			maxPerRow:      operationsPerKeyboardRow,
			currentPage:    opts.page,
		},
		func() ([]models.Operation, error) {
			operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
				BalanceID:           opts.balanceID,
				SortByCreatedAtDesc: true,
				Pagination: &Pagination{
					Limit: operationsPerKeyboard,
					Page:  opts.page,
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("list operations from store")
				return nil, fmt.Errorf("list operations from store: %w", err)
			}

			if len(operations) == 0 {
				logger.Info().Msg("operations not found")
				return nil, ErrOperationsNotFound
			}

			return operations, nil
		})
	if err != nil {
		logger.Error().Err(err).Msg("paginate operations")
		return nil, fmt.Errorf("paginate operations: %w", err)
	}

	return keyboard, nil
}

const (
	balanceSubscriptionsPerKeyboard    = 5
	balanceSubscriptionsPerKeyboardRow = 1
)

type getBalanceSubscriptionsKeyboardOptions struct {
	balanceID string
	page      int
}

func (h handlerService) getBalanceSubscriptionsKeyboard(ctx context.Context, opts getBalanceSubscriptionsKeyboardOptions) ([]InlineKeyboardRow, error) {
	logger := h.logger.With().Str("name", "handlerService.getBalanceSubscriptionsKeyboard").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceSubscriptionsCount, err := h.stores.BalanceSubscription.Count(ctx, ListBalanceSubscriptionFilter{
		BalanceID: opts.balanceID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("count balance subscriptions in store")
		return nil, fmt.Errorf("count balance subscriptions in store: %w", err)
	}
	if balanceSubscriptionsCount == 0 {
		logger.Info().Any("balanceID", opts.balanceID).Msg("balance subscriptions not found")
		return nil, ErrNoBalanceSubscriptionsFound
	}

	keyboard, err := paginateInlineKeyboard(
		inlineKeyboardPaginatorOptions{
			totalCount:     balanceSubscriptionsCount,
			maxPerKeyboard: balanceSubscriptionsPerKeyboard,
			maxPerRow:      balanceSubscriptionsPerKeyboardRow,
			currentPage:    opts.page,
		},
		func() ([]models.BalanceSubscription, error) {
			balanceSubscriptions, err := h.stores.BalanceSubscription.List(ctx, ListBalanceSubscriptionFilter{
				BalanceID:            opts.balanceID,
				OrderByCreatedAtDesc: true,
				Pagination: &Pagination{
					Limit: balanceSubscriptionsPerKeyboard,
					Page:  opts.page,
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("list balance subscriptions from store")
				return nil, fmt.Errorf("list balance subscriptions from store: %w", err)
			}
			if len(balanceSubscriptions) == 0 {
				logger.Info().Msg("balance subscriptions not found")
				return nil, ErrNoBalanceSubscriptionsFound
			}

			return balanceSubscriptions, nil
		})
	if err != nil {
		logger.Error().Err(err).Msg("paginate balance subscriptions")
		return nil, fmt.Errorf("paginate balance subscriptions: %w", err)
	}

	return keyboard, nil
}

const (
	currenciesPerKeyboard    = 10
	currenciesPerKeyboardRow = 3
)

func (h handlerService) getCurrenciesKeyboard(ctx context.Context, page int) ([]InlineKeyboardRow, error) {
	logger := h.logger.With().Str("name", "handlerService.getCurrenciesKeyboardForBalance").Logger()

	keyboard, err := paginateInlineKeyboard(
		inlineKeyboardPaginatorOptions{
			totalCount:     availableCurrenciesCount,
			maxPerKeyboard: currenciesPerKeyboard,
			maxPerRow:      currenciesPerKeyboardRow,
			currentPage:    page,
		},
		func() ([]models.Currency, error) {
			currencies, err := h.stores.Currency.List(ctx, ListCurrenciesFilter{
				Pagination: &Pagination{
					Limit: currenciesPerKeyboard,
					Page:  page,
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("list currencies from store")
				return nil, fmt.Errorf("list currencies from store: %w", err)
			}
			if len(currencies) == 0 {
				logger.Info().Msg("currencies not found")
				return nil, fmt.Errorf("no currencies found")
			}

			return currencies, nil
		})
	if err != nil {
		logger.Error().Err(err).Msg("paginate currencies")
		return nil, fmt.Errorf("paginate currencies: %w", err)
	}

	return keyboard, nil
}
