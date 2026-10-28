package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
)

var (
	errUnknownSearchFilter = errs.New("Unknown filter! Please choose one from the keyboard.")
	errUnknownSearchType   = errs.New("Unknown operation type! Please choose one from the keyboard.")
)

type searchFilterAction string

const (
	searchFilterActionQuery      searchFilterAction = "query"
	searchFilterActionType       searchFilterAction = "type"
	searchFilterActionCategory   searchFilterAction = "category"
	searchFilterActionAmountFrom searchFilterAction = "amount_from"
	searchFilterActionAmountTo   searchFilterAction = "amount_to"
	searchFilterActionDateFrom   searchFilterAction = "date_from"
	searchFilterActionDateTo     searchFilterAction = "date_to"
	searchFilterActionReset      searchFilterAction = "reset"
	searchFilterActionShow       searchFilterAction = "show"
)

func parseSearchFilterAction(raw string) (searchFilterAction, error) {
	switch searchFilterAction(raw) {
	case searchFilterActionQuery:
		return searchFilterActionQuery, nil
	case searchFilterActionType:
		return searchFilterActionType, nil
	case searchFilterActionCategory:
		return searchFilterActionCategory, nil
	case searchFilterActionAmountFrom:
		return searchFilterActionAmountFrom, nil
	case searchFilterActionAmountTo:
		return searchFilterActionAmountTo, nil
	case searchFilterActionDateFrom:
		return searchFilterActionDateFrom, nil
	case searchFilterActionDateTo:
		return searchFilterActionDateTo, nil
	case searchFilterActionReset:
		return searchFilterActionReset, nil
	case searchFilterActionShow:
		return searchFilterActionShow, nil
	default:
		return "", errUnknownSearchFilter
	}
}

var searchFilterKeyboard = []InlineKeyboardRow{
	{
		Buttons: []InlineKeyboardButton{
			{
				Text: "📝 Text",
				Data: string(searchFilterActionQuery),
			},
			{
				Text: "📌 Type",
				Data: string(searchFilterActionType),
			},
			{
				Text: "📂 Category",
				Data: string(searchFilterActionCategory),
			},
		},
	},
	{
		Buttons: []InlineKeyboardButton{
			{
				Text: "💰 Amount from",
				Data: string(searchFilterActionAmountFrom),
			},
			{
				Text: "💰 Amount to",
				Data: string(searchFilterActionAmountTo),
			},
		},
	},
	{
		Buttons: []InlineKeyboardButton{
			{
				Text: "📅 Date from",
				Data: string(searchFilterActionDateFrom),
			},
			{
				Text: "📅 Date to",
				Data: string(searchFilterActionDateTo),
			},
		},
	},
	{
		Buttons: []InlineKeyboardButton{
			{
				Text: "♻️ Reset",
				Data: string(searchFilterActionReset),
			},
			{
				Text: "🔎 Show results",
				Data: string(searchFilterActionShow),
			},
		},
	},
}

var searchOperationTypeKeyboard = []InlineKeyboardRow{
	{
		Buttons: []InlineKeyboardButton{
			{
				Text: "Incoming 💰",
				Data: string(model.OperationTypeIncoming),
			},
			{
				Text: "Spending 💸",
				Data: string(model.OperationTypeSpending),
			},
			{
				Text: "Transfer ➡️",
				Data: string(model.OperationTypeTransfer),
			},
		},
	},
}

var searchMetadataKeys = []model.MetadataKey{
	model.SearchQueryMetadataKey,
	model.SearchOperationTypeMetadataKey,
	model.SearchCategoryIDMetadataKey,
	model.SearchCategoryTitleMetadataKey,
	model.SearchAmountFromMetadataKey,
	model.SearchAmountToMetadataKey,
	model.SearchDateFromMetadataKey,
	model.SearchDateToMetadataKey,
}

func (h handlerService) handleSearchOperationsFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.ChooseBalanceFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose balance to search operations in:",
		InlineKeyboard: getInlineKeyboardRows(opts.user.Balances, 2),
	})
}

func (h handlerService) handleChooseBalanceFlowStepForSearchOperations(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	opts.stateMetaData.Add(model.BalanceNameMetadataKey, opts.message.GetText())

	return model.ChooseSearchFilterFlowStep, h.showSearchFilters(opts, true)
}

func (h handlerService) handleChooseSearchFilterFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	action, err := parseSearchFilterAction(opts.message.GetText())
	if err != nil {
		return "", err
	}

	switch action {
	case searchFilterActionQuery:
		return model.EnterSearchTextFlowStep, h.updateSearchMessage(opts, "Enter text to search in operation descriptions:", nil)
	case searchFilterActionType:
		return model.ChooseSearchOperationTypeFlowStep, h.updateSearchMessage(opts, "Choose operation type:", searchOperationTypeKeyboard)
	case searchFilterActionCategory:
		categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
			UserID: opts.user.ID,
		})
		if err != nil {
			return "", fmt.Errorf("list categories from store: %w", err)
		}
		if len(categories) == 0 {
			return "", ErrCategoriesNotFound
		}

		return model.ChooseSearchCategoryFlowStep, h.updateSearchMessage(opts, "Choose category:", getSearchCategoryKeyboard(categories))
	case searchFilterActionAmountFrom:
		return model.EnterSearchAmountFromFlowStep, h.updateSearchMessage(opts, "Enter minimal amount:", nil)
	case searchFilterActionAmountTo:
		return model.EnterSearchAmountToFlowStep, h.updateSearchMessage(opts, "Enter maximal amount:", nil)
	case searchFilterActionDateFrom:
		return model.ChooseSearchDateFromFlowStep, h.updateSearchMessage(opts, "Pick start date:", getSearchCalendarKeyboard(time.Now().In(opts.user.GetLocation())))
	case searchFilterActionDateTo:
		return model.ChooseSearchDateToFlowStep, h.updateSearchMessage(opts, "Pick end date:", getSearchCalendarKeyboard(time.Now().In(opts.user.GetLocation())))
	case searchFilterActionReset:
		for _, key := range searchMetadataKeys {
			delete(opts.stateMetaData, key)
		}

		return model.ChooseSearchFilterFlowStep, h.showSearchFilters(opts, true)
	case searchFilterActionShow:
		return h.showSearchResults(ctx, opts)
	default:
		return "", errUnknownSearchFilter
	}
}

func (h handlerService) handleEnterSearchTextFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	opts.stateMetaData.Add(model.SearchQueryMetadataKey, strings.TrimSpace(opts.message.GetText()))

	return model.ChooseSearchFilterFlowStep, h.showSearchFilters(opts, false)
}

func (h handlerService) handleChooseSearchOperationTypeFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	operationType := model.OperationType(opts.message.GetText())
	switch operationType {
	case model.OperationTypeIncoming, model.OperationTypeSpending, model.OperationTypeTransfer:
		opts.stateMetaData.Add(model.SearchOperationTypeMetadataKey, string(operationType))
	default:
		return "", errUnknownSearchType
	}

	return model.ChooseSearchFilterFlowStep, h.showSearchFilters(opts, true)
}

func (h handlerService) handleChooseSearchCategoryFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
		ID:     opts.message.GetText(),
		UserID: opts.user.ID,
	})
	if err != nil {
		return "", fmt.Errorf("get category from store: %w", err)
	}
	if category == nil {
		return "", ErrCategoryNotFound
	}

	opts.stateMetaData.Add(model.SearchCategoryIDMetadataKey, category.ID)
	opts.stateMetaData.Add(model.SearchCategoryTitleMetadataKey, category.Title)

	return model.ChooseSearchFilterFlowStep, h.showSearchFilters(opts, true)
}

func (h handlerService) handleEnterSearchAmountFromFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	return h.enterSearchAmount(opts, model.SearchAmountFromMetadataKey)
}

func (h handlerService) handleEnterSearchAmountToFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	return h.enterSearchAmount(opts, model.SearchAmountToMetadataKey)
}

func (h handlerService) enterSearchAmount(opts flowProcessingOptions, key model.MetadataKey) (model.FlowStep, error) {
	amount := strings.TrimSpace(opts.message.GetText())
	_, err := money.NewFromString(amount)
	if err != nil {
		return "", ErrInvalidAmountFormat
	}

	opts.stateMetaData.Add(key, amount)

	return model.ChooseSearchFilterFlowStep, h.showSearchFilters(opts, false)
}

func (h handlerService) handleChooseSearchDateFromFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	return h.chooseSearchDate(opts, model.SearchDateFromMetadataKey)
}

func (h handlerService) handleChooseSearchDateToFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	return h.chooseSearchDate(opts, model.SearchDateToMetadataKey)
}

const (
	searchDateLayout          = time.DateOnly
	searchCalendarMonthPrefix = "cal:"
	searchCalendarMonthLayout = "2006-01"
	searchCalendarIgnoreData  = "ignore"
)

func (h handlerService) chooseSearchDate(opts flowProcessingOptions, key model.MetadataKey) (model.FlowStep, error) {
	currentStep := model.ChooseSearchDateFromFlowStep
	if key == model.SearchDateToMetadataKey {
		currentStep = model.ChooseSearchDateToFlowStep
	}

	text := opts.message.GetText()
	switch {
	case text == searchCalendarIgnoreData:
		return currentStep, nil
	case strings.HasPrefix(text, searchCalendarMonthPrefix):
		month, err := time.Parse(searchCalendarMonthLayout, strings.TrimPrefix(text, searchCalendarMonthPrefix))
		if err != nil {
			return "", ErrInvalidDateFormat
		}

		return currentStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        "Pick date:",
			UpdatedInlineKeyboard: getSearchCalendarKeyboard(month),
		})
	}

	_, err := time.Parse(searchDateLayout, text)
	if err != nil {
		return "", ErrInvalidDateFormat
	}

	opts.stateMetaData.Add(key, text)

	return model.ChooseSearchFilterFlowStep, h.showSearchFilters(opts, true)
}

func (h handlerService) handleSearchOperationsResultsFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	if !isPaginationNeeded(opts.message.GetText()) {
		return "", errUnknownSearchFilter
	}

	opts.stateMetaData.Add(model.PageMetadataKey, calculateNextPage(opts.message.GetText(), opts.stateMetaData))

	return h.showSearchResults(ctx, opts)
}

func (h handlerService) showSearchResults(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	balanceName, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.BalanceNameMetadataKey)
	if !ok {
		return "", fmt.Errorf("balance name not found in metadata")
	}

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		Name:            balanceName,
		UserID:          opts.user.ID,
		PreloadCurrency: true,
	})
	if err != nil {
		return "", fmt.Errorf("get balance from store: %w", err)
	}
	if balance == nil {
		return "", ErrBalanceNotFound
	}

	page := firstPage
	if isPaginationNeeded(opts.message.GetText()) {
		page = calculateNextPage(opts.message.GetText(), opts.stateMetaData)
	}
	opts.stateMetaData.Add(model.PageMetadataKey, page)

	filter := buildSearchOperationsFilter(opts.stateMetaData, balance.ID, opts.user.GetLocation())

	message, keyboard, err := h.getOperationsHistoryKeyboard(ctx, getOperationsHistoryKeyboardOptions{
		balance:  balance,
		page:     page,
		location: opts.user.GetLocation(),
		search:   &filter,
	})
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}

		return "", fmt.Errorf("get operations keyboard: %w", err)
	}

	return model.SearchOperationsResultsFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedInlineKeyboard:   keyboard,
		UpdatedMessage:          message,
	})
}

func buildSearchOperationsFilter(metadata model.Metadata, balanceID string, location *time.Location) ListOperationsFilter {
	filter := ListOperationsFilter{
		BalanceID:   balanceID,
		Location:    location,
		Description: searchMetadataValue(metadata, model.SearchQueryMetadataKey),
		CategoryID:  searchMetadataValue(metadata, model.SearchCategoryIDMetadataKey),
		AmountFrom:  searchMetadataValue(metadata, model.SearchAmountFromMetadataKey),
		AmountTo:    searchMetadataValue(metadata, model.SearchAmountToMetadataKey),
	}

	switch model.OperationType(searchMetadataValue(metadata, model.SearchOperationTypeMetadataKey)) {
	case model.OperationTypeIncoming:
		filter.Types = []model.OperationType{model.OperationTypeIncoming}
	case model.OperationTypeSpending:
		filter.Types = []model.OperationType{model.OperationTypeSpending}
	case model.OperationTypeTransfer:
		filter.Types = []model.OperationType{model.OperationTypeTransferIn, model.OperationTypeTransferOut}
	}

	dateFrom, hasDateFrom := parseSearchDate(searchMetadataValue(metadata, model.SearchDateFromMetadataKey), location)
	dateTo, hasDateTo := parseSearchDate(searchMetadataValue(metadata, model.SearchDateToMetadataKey), location)
	if hasDateFrom || hasDateTo {
		between := BetweenFilter{
			From: time.Unix(0, 0),
			To:   time.Now().AddDate(100, 0, 0),
		}
		if hasDateFrom {
			between.From = dateFrom
		}
		if hasDateTo {
			between.To = dateTo.AddDate(0, 0, 1)
		}
		filter.BetweenFilter = &between
	}

	return filter
}

func parseSearchDate(raw string, location *time.Location) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}

	date, err := time.ParseInLocation(searchDateLayout, raw, location)
	if err != nil {
		return time.Time{}, false
	}

	return date, true
}

func searchMetadataValue(metadata model.Metadata, key model.MetadataKey) string {
	value, _ := model.GetTypedFromMetadata[string](metadata, key)
	return value
}

func (h handlerService) showSearchFilters(opts flowProcessingOptions, updateMessage bool) error {
	message := "Search filters:\n\n" + buildSearchSummary(opts.stateMetaData) + "\n\nChoose a filter to change or show results:"

	if updateMessage {
		return h.updateSearchMessage(opts, message, searchFilterKeyboard)
	}

	return h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        message,
		InlineKeyboard: searchFilterKeyboard,
	})
}

func (h handlerService) updateSearchMessage(opts flowProcessingOptions, message string, keyboard []InlineKeyboardRow) error {
	return h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        message,
		UpdatedInlineKeyboard: keyboard,
	})
}

func buildSearchSummary(metadata model.Metadata) string {
	valueOrAny := func(value string) string {
		if value == "" {
			return "any"
		}

		return value
	}

	return fmt.Sprintf(
		"📝 Text: %s\n📌 Type: %s\n📂 Category: %s\n💰 Amount: %s — %s\n📅 Date: %s — %s",
		valueOrAny(searchMetadataValue(metadata, model.SearchQueryMetadataKey)),
		valueOrAny(searchMetadataValue(metadata, model.SearchOperationTypeMetadataKey)),
		valueOrAny(searchMetadataValue(metadata, model.SearchCategoryTitleMetadataKey)),
		valueOrAny(searchMetadataValue(metadata, model.SearchAmountFromMetadataKey)),
		valueOrAny(searchMetadataValue(metadata, model.SearchAmountToMetadataKey)),
		valueOrAny(searchMetadataValue(metadata, model.SearchDateFromMetadataKey)),
		valueOrAny(searchMetadataValue(metadata, model.SearchDateToMetadataKey)),
	)
}

func getSearchCategoryKeyboard(categories []model.Category) []InlineKeyboardRow {
	const categoriesPerRow = 2

	rows := make([]InlineKeyboardRow, 0, len(categories)/categoriesPerRow+1)
	for i, category := range categories {
		if i%categoriesPerRow == 0 {
			rows = append(rows, InlineKeyboardRow{})
		}

		last := len(rows) - 1
		rows[last].Buttons = append(rows[last].Buttons, InlineKeyboardButton{
			Text: category.Title,
			Data: category.ID,
		})
	}

	return rows
}

func getSearchCalendarKeyboard(month time.Time) []InlineKeyboardRow {
	firstDay := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	previousMonth := firstDay.AddDate(0, -1, 0)
	nextMonth := firstDay.AddDate(0, 1, 0)

	rows := []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: "◀️",
					Data: searchCalendarMonthPrefix + previousMonth.Format(searchCalendarMonthLayout),
				},
				{
					Text: firstDay.Format("January 2006"),
					Data: searchCalendarIgnoreData,
				},
				{
					Text: "▶️",
					Data: searchCalendarMonthPrefix + nextMonth.Format(searchCalendarMonthLayout),
				},
			},
		},
	}

	weekdays := InlineKeyboardRow{}
	for _, weekday := range []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"} {
		weekdays.Buttons = append(weekdays.Buttons, InlineKeyboardButton{
			Text: weekday,
			Data: searchCalendarIgnoreData,
		})
	}
	rows = append(rows, weekdays)

	emptyCell := InlineKeyboardButton{
		Text: " ",
		Data: searchCalendarIgnoreData,
	}
	week := InlineKeyboardRow{}
	for i := 0; i < (int(firstDay.Weekday())+6)%7; i++ {
		week.Buttons = append(week.Buttons, emptyCell)
	}

	for day := firstDay; day.Month() == firstDay.Month(); day = day.AddDate(0, 0, 1) {
		week.Buttons = append(week.Buttons, InlineKeyboardButton{
			Text: fmt.Sprint(day.Day()),
			Data: day.Format(searchDateLayout),
		})

		if len(week.Buttons) == 7 {
			rows = append(rows, week)
			week = InlineKeyboardRow{}
		}
	}

	if len(week.Buttons) > 0 {
		for len(week.Buttons) < 7 {
			week.Buttons = append(week.Buttons, emptyCell)
		}
		rows = append(rows, week)
	}

	return rows
}
