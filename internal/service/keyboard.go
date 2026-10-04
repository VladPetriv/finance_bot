package service

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/errs"
)

type identifiable interface {
	GetID() string
	GetName() string
}

const (
	timezonesPerKeyboard    = 20
	timezonesPerKeyboardRow = 2
	timezoneRegionsPerRow   = 2
)

func buildTimezoneRegionsKeyboard() []InlineKeyboardRow {
	rows := make([]InlineKeyboardRow, 0, len(model.TimezoneRegions)/timezoneRegionsPerRow+2)
	for i := 0; i < len(model.TimezoneRegions); i += timezoneRegionsPerRow {
		buttons := make([]InlineKeyboardButton, 0, timezoneRegionsPerRow)
		for _, region := range model.TimezoneRegions[i:min(i+timezoneRegionsPerRow, len(model.TimezoneRegions))] {
			buttons = append(buttons, InlineKeyboardButton{
				Text: region,
			})
		}

		rows = append(rows, InlineKeyboardRow{
			Buttons: buttons,
		})
	}

	return append(rows, InlineKeyboardRow{
		Buttons: []InlineKeyboardButton{
			{
				Text: "UTC",
			},
		},
	})
}

func getTimezonesKeyboard(region string, page int) ([]InlineKeyboardRow, error) {
	timezones := model.GetTimezonesByRegion(region)

	maxPage := calculateMaxPage(len(timezones), timezonesPerKeyboard)
	page = max(firstPage, min(page, maxPage))

	start := (page - 1) * timezonesPerKeyboard
	end := min(start+timezonesPerKeyboard, len(timezones))

	return paginateInlineKeyboard(
		inlineKeyboardPaginatorOptions{
			totalCount:     len(timezones),
			maxPerKeyboard: timezonesPerKeyboard,
			maxPerRow:      timezonesPerKeyboardRow,
			currentPage:    page,
		},
		func() ([]model.Timezone, error) {
			return timezones[start:end], nil
		},
	)
}

func getInlineKeyboardRows[T identifiable](data []T, elementLimitPerRow int) []InlineKeyboardRow {
	inlineKeyboardRows := make([]InlineKeyboardRow, 0)

	var currentRow InlineKeyboardRow
	for i, entry := range data {
		currentRow.Buttons = append(currentRow.Buttons, InlineKeyboardButton{
			Text: entry.GetName(),
		})

		// When row is full or we're at the last data item, append row
		if len(currentRow.Buttons) == elementLimitPerRow || i == len(data)-1 {
			inlineKeyboardRows = append(inlineKeyboardRows, currentRow)
			currentRow = InlineKeyboardRow{} // Reset current row
		}
	}

	return inlineKeyboardRows
}

const (
	operationsPerKeyboard    = 5
	operationsPerKeyboardRow = 1
)

type getOperationsKeyboardOptions struct {
	balanceID string
	page      int
	location  *time.Location
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
		func() ([]model.Operation, error) {
			operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
				BalanceID:            opts.balanceID,
				OrderByCreatedAtDesc: true,
				Pagination: &Pagination{
					Limit: operationsPerKeyboard,
					Page:  opts.page,
				},
				Location: opts.location,
			})
			if err != nil {
				logger.Error().Err(err).Msg("list operations from store")
				return nil, fmt.Errorf("list operations from store: %w", err)
			}
			convertOperationsToLocation(operations, opts.location)
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

type getOperationsHistoryKeyboardOptions struct {
	balance        *model.Balance
	creationPeriod model.CreationPeriod
	page           int
	location       *time.Location
}

func (h handlerService) getOperationsHistoryKeyboard(ctx context.Context, opts getOperationsHistoryKeyboardOptions) (string, []InlineKeyboardRow, error) {
	logger := h.logger.With().Str("name", "handlerService.getOperationsHistoryKeyboard").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	operationsCount, err := h.stores.Operation.Count(ctx, ListOperationsFilter{
		BalanceID:      opts.balance.ID,
		CreationPeriod: opts.creationPeriod,
		Location:       opts.location,
	})
	if err != nil {
		logger.Error().Err(err).Msg("count operations")
		return "", nil, fmt.Errorf("count operations: %w", err)
	}
	if operationsCount == 0 {
		logger.Info().Any("balanceID", opts.balance.ID).Msg("operations not found")
		return "", nil, ErrOperationsNotFound
	}

	message, keyboard, err := paginateTextUsingInlineKeybaord(
		inlineKeyboardPaginatorOptions{
			totalCount:     operationsCount,
			maxPerKeyboard: operationsPerKeyboard,
			maxPerRow:      operationsPerKeyboardRow,
			currentPage:    opts.page,
		},
		func() (string, error) {
			operations, err := h.stores.Operation.List(ctx, ListOperationsFilter{
				BalanceID:            opts.balance.ID,
				CreationPeriod:       opts.creationPeriod,
				Location:             opts.location,
				OrderByCreatedAtDesc: true,
				Pagination: &Pagination{
					Limit: operationsPerKeyboard,
					Page:  opts.page,
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("list operations from store")
				return "", fmt.Errorf("list operations from store: %w", err)
			}
			if len(operations) == 0 {
				logger.Info().Msg("operations not found")
				return "", ErrOperationsNotFound
			}
			convertOperationsToLocation(operations, opts.location)

			outputMessage := fmt.Sprintf(
				"💰 *Balance:* %v%s\n📅 *Period:* %v\n\n",
				opts.balance.Amount, opts.balance.GetCurrency().Symbol, opts.creationPeriod,
			)

			separator := "━━━━━━━━━━━━━━━"
			for _, o := range operations {
				category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
					ID: o.CategoryID,
				})
				if err != nil {
					logger.Error().Err(err).Msg("get category from store")
					return "", fmt.Errorf("get category from store: %w", err)
				}
				if category == nil {
					logger.Error().Msg("category not found")
					continue
				}

				emoji, typeLabel := model.GetOperationTypeLabel(o.Type)
				outputMessage += fmt.Sprintf(
					"%s\n"+
						"📌 *Operation:* %s %s\n"+
						"📝 *Description:* %s\n"+
						"📂 *Category:* %s\n"+
						"💰 *Amount:* %s%s\n",
					separator,
					emoji,
					typeLabel,
					o.Description,
					category.Title,
					o.Amount,
					opts.balance.GetCurrency().Symbol,
				)

				if o.ExchangeRate != "" {
					outputMessage += fmt.Sprintf("💱 *Exchange Rate:* %s\n", o.ExchangeRate)
				}

				outputMessage += fmt.Sprintf("🕐 *Date:* %s\n", o.CreatedAt.Format(time.ANSIC))

			}
			outputMessage += separator

			return outputMessage, nil
		},
	)
	if err != nil {
		logger.Error().Err(err).Msg("paginate operations")
		return "", nil, fmt.Errorf("paginate operations: %w", err)
	}

	return message, keyboard, nil
}

const (
	balanceSubscriptionsPerKeyboard    = 5
	balanceSubscriptionsPerKeyboardRow = 1
)

type getBalanceSubscriptionsKeyboardOptions struct {
	userID    string
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
		func() ([]model.BalanceSubscription, error) {
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

func (h handlerService) getListBalanceSubscriptionsKeyboard(ctx context.Context, opts getBalanceSubscriptionsKeyboardOptions) (string, []InlineKeyboardRow, error) {
	logger := h.logger.With().Str("name", "handlerService.getOperationsHistoryKeyboard").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	balanceSubscriptionsCount, err := h.stores.BalanceSubscription.Count(ctx, ListBalanceSubscriptionFilter{
		BalanceID: opts.balanceID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("count balance subscriptions in store")
		return "", nil, fmt.Errorf("count balance subscriptions in store: %w", err)
	}
	if balanceSubscriptionsCount == 0 {
		logger.Info().Any("balanceID", opts.balanceID).Msg("balance subscriptions not found")
		return "", nil, ErrNoBalanceSubscriptionsFound
	}

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: opts.userID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list categories")
		return "", nil, fmt.Errorf("list categories: %w", err)
	}
	if len(categories) == 0 {
		return "", nil, ErrCategoriesNotFound
	}

	message, keyboard, err := paginateTextUsingInlineKeybaord(
		inlineKeyboardPaginatorOptions{
			totalCount:     balanceSubscriptionsCount,
			maxPerKeyboard: balanceSubscriptionsPerKeyboard,
			maxPerRow:      balanceSubscriptionsPerKeyboardRow,
			currentPage:    opts.page,
		},
		func() (string, error) {
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
				return "", fmt.Errorf("list balance subscriptions from store: %w", err)
			}
			if len(balanceSubscriptions) == 0 {
				logger.Info().Msg("balance subscriptions not found")
				return "", ErrNoBalanceSubscriptionsFound
			}

			var outputMessage string
			for _, subscription := range balanceSubscriptions {
				categoryTitle := "—"
				categoryIndex := slices.IndexFunc(categories, func(category model.Category) bool {
					return category.ID == subscription.CategoryID
				})
				if categoryIndex != -1 {
					categoryTitle = categories[categoryIndex].Title
				}

				var startDate string
				if !subscription.StartAt.IsZero() {
					startDate = subscription.StartAt.Format("02 Jan 2006")
				}

				outputMessage += fmt.Sprintf(
					"💰 *Title*: %s\n"+
						"📦 *Amount*: %s\n"+
						"⏰ *Frequency*: %s\n"+
						"📅 *Start Date*: %s\n"+
						"🏷️ Category: %s\n"+
						"──────────────\n",
					subscription.Name,
					subscription.Amount,
					subscription.Period,
					startDate,
					categoryTitle,
				)
			}

			return outputMessage, nil
		},
	)
	if err != nil {
		logger.Error().Err(err).Msg("paginate operations")
		return "", nil, fmt.Errorf("paginate operations: %w", err)
	}

	return message, keyboard, nil
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
		func() ([]model.Currency, error) {
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

const (
	automaticReportsPerKeyboard    = 5
	automaticReportsPerKeyboardRow = 1
)

type getAutomaticReportsKeyboardOptions struct {
	user *model.User
	page int
}

func (h handlerService) getAutomaticReportsKeyboard(ctx context.Context, opts getAutomaticReportsKeyboardOptions) ([]InlineKeyboardRow, error) {
	logger := h.logger.With().Str("name", "handlerService.getAutomaticReportsKeyboard").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	automaticReportsCount, err := h.stores.AutomaticReport.Count(ctx, ListAutomaticReportFilter{
		BalanceIDs: opts.user.GetBalancesIDs(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("count automatic reports in store")
		return nil, fmt.Errorf("count automatic reports in store: %w", err)
	}
	if automaticReportsCount == 0 {
		logger.Info().Msg("automatic reports not found")
		return nil, ErrNoAutomaticReportsFound
	}

	keyboard, err := paginateInlineKeyboard(
		inlineKeyboardPaginatorOptions{
			totalCount:     automaticReportsCount,
			maxPerKeyboard: automaticReportsPerKeyboard,
			maxPerRow:      automaticReportsPerKeyboardRow,
			currentPage:    opts.page,
		},
		func() ([]model.AutomaticReport, error) {
			automaticReports, err := h.stores.AutomaticReport.List(ctx, ListAutomaticReportFilter{
				BalanceIDs:           opts.user.GetBalancesIDs(),
				OrderByCreatedAtDesc: true,
				Pagination: &Pagination{
					Limit: automaticReportsPerKeyboard,
					Page:  opts.page,
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("list automatic reports from store")
				return nil, fmt.Errorf("list automatic reports from store: %w", err)
			}
			if len(automaticReports) == 0 {
				logger.Info().Msg("automatic reports not found")
				return nil, ErrNoAutomaticReportsFound
			}

			return automaticReports, nil
		})
	if err != nil {
		if errs.IsExpected(err) {
			return nil, err
		}
		logger.Error().Err(err).Msg("paginate automatic reports")
		return nil, fmt.Errorf("paginate automatic reports: %w", err)
	}

	return keyboard, nil
}

func (h handlerService) getListAutomaticReportsKeyboard(ctx context.Context, opts getAutomaticReportsKeyboardOptions) (string, []InlineKeyboardRow, error) {
	logger := h.logger.With().Str("name", "handlerService.getListAutomaticReportsKeyboard").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	automaticReportsCount, err := h.stores.AutomaticReport.Count(ctx, ListAutomaticReportFilter{
		BalanceIDs: opts.user.GetBalancesIDs(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("count automatic reports in store")
		return "", nil, fmt.Errorf("count automatic reports in store: %w", err)
	}
	if automaticReportsCount == 0 {
		logger.Info().Msg("automatic reports not found")
		return "", nil, ErrNoAutomaticReportsFound
	}

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: opts.user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list categories")
		return "", nil, fmt.Errorf("list categories: %w", err)
	}

	message, keyboard, err := paginateTextUsingInlineKeybaord(
		inlineKeyboardPaginatorOptions{
			totalCount:     automaticReportsCount,
			maxPerKeyboard: automaticReportsPerKeyboard,
			maxPerRow:      automaticReportsPerKeyboardRow,
			currentPage:    opts.page,
		},
		func() (string, error) {
			automaticReports, err := h.stores.AutomaticReport.List(ctx, ListAutomaticReportFilter{
				BalanceIDs:           opts.user.GetBalancesIDs(),
				OrderByCreatedAtDesc: true,
				Pagination: &Pagination{
					Limit: automaticReportsPerKeyboard,
					Page:  opts.page,
				},
			})
			if err != nil {
				logger.Error().Err(err).Msg("list automatic reports from store")
				return "", fmt.Errorf("list automatic reports from store: %w", err)
			}
			if len(automaticReports) == 0 {
				logger.Info().Msg("automatic reports not found")
				return "", ErrNoAutomaticReportsFound
			}

			var outputMessage string
			for _, automaticReport := range automaticReports {
				balanceNames := make([]string, 0, len(automaticReport.BalanceIDs))
				for _, balanceID := range automaticReport.BalanceIDs {
					balance := opts.user.GetBalance(balanceID)
					if balance != nil {
						balanceNames = append(balanceNames, balance.Name)
					}
				}

				categoryTitles := make([]string, 0, len(automaticReport.CategoryIDs))
				for _, categoryID := range automaticReport.CategoryIDs {
					categoryIndex := slices.IndexFunc(categories, func(category model.Category) bool {
						return category.ID == categoryID
					})
					if categoryIndex != -1 {
						categoryTitles = append(categoryTitles, categories[categoryIndex].Title)
					}
				}

				outputMessage += fmt.Sprintf(
					"📊 *Title*: %s\n"+
						"⏰ *Frequency*: %s\n"+
						"💰 *Balances*: %s\n"+
						"🏷️ *Categories*: %s\n"+
						"──────────────\n",
					automaticReport.Name,
					automaticReport.Period.GetLabel(),
					strings.Join(balanceNames, ", "),
					strings.Join(categoryTitles, ", "),
				)
			}

			return outputMessage, nil
		},
	)
	if err != nil {
		if errs.IsExpected(err) {
			return "", nil, err
		}
		logger.Error().Err(err).Msg("paginate automatic reports")
		return "", nil, fmt.Errorf("paginate automatic reports: %w", err)
	}

	return message, keyboard, nil
}

const automaticReportSelectionDoneData = "done"

func getSelectionInlineKeyboardRows[T identifiable](data []T, selectedIDs []string, elementLimitPerRow int) []InlineKeyboardRow {
	inlineKeyboardRows := make([]InlineKeyboardRow, 0)

	var currentRow InlineKeyboardRow
	for i, entry := range data {
		text := entry.GetName()
		if slices.Contains(selectedIDs, entry.GetID()) {
			text = "✅ " + text
		}

		currentRow.Buttons = append(currentRow.Buttons, InlineKeyboardButton{
			Text: text,
			Data: entry.GetID(),
		})

		if len(currentRow.Buttons) == elementLimitPerRow || i == len(data)-1 {
			inlineKeyboardRows = append(inlineKeyboardRows, currentRow)
			currentRow = InlineKeyboardRow{}
		}
	}

	inlineKeyboardRows = append(inlineKeyboardRows, InlineKeyboardRow{
		Buttons: []InlineKeyboardButton{
			{
				Text: "Done ✔️",
				Data: automaticReportSelectionDoneData,
			},
		},
	})

	return inlineKeyboardRows
}

func convertOperationsToLocation(operations []model.Operation, location *time.Location) {
	if location == nil {
		return
	}

	for i := range operations {
		operations[i].CreatedAt = operations[i].CreatedAt.In(location)
	}
}
