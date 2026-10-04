package service

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/google/uuid"
)

// Create Automatic Reports
func (h *handlerService) handleCreateAutomaticReportFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleCreateAutomaticReportFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	return model.EnterAutomaticReportNameFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Enter automatic report name:")
}

func (h *handlerService) handleEnterAutomaticReportNameFlowStepForCreate(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterAutomaticReportNameFlowStepForCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	opts.stateMetaData.Add(model.AutomaticReportNameMetadataKey, opts.message.GetText())
	return model.ChooseAutomaticReportPeriodFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		Message:                 automaticReportPeriodMessage,
		FormatMessageInMarkDown: true,
		InlineKeyboard:          automaticReportPeriodKeyboard,
	})
}

func (h *handlerService) handleChooseAutomaticReportPeriodFlowStepForCreate(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseAutomaticReportPeriodFlowStepForCreate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	period, err := model.ParseAutomaticReportPeriod(opts.message.GetText())
	if err != nil {
		return model.ChooseAutomaticReportPeriodFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:         opts.message.GetChatID(),
			Message:        "Invalid automatic report frequency. Please choose from the options below:",
			InlineKeyboard: automaticReportPeriodKeyboard,
		})
	}

	opts.stateMetaData.Add(model.AutomaticReportPeriodMetadataKey, period)
	return model.ChooseBalanceFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        selectAutomaticReportBalancesMessage,
		UpdatedInlineKeyboard: getSelectionInlineKeyboardRows(opts.user.Balances, nil, 2),
	})
}

const (
	automaticReportPeriodMessage = "How often do you want to receive the report?\n\n" +
		"🔁 *Every ...* — rolling, counted from today\n" +
		"📌 *End of ...* — calendar day, week (Mon-Sun), month, quarter or year"
	selectAutomaticReportBalancesMessage   = "Select balances for the report and press Done:"
	selectAutomaticReportCategoriesMessage = "Select categories for the report and press Done:"
)

func (h *handlerService) handleChooseBalanceFlowStepForCreateAutomaticReport(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForCreateAutomaticReport").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	selectedBalanceIDs := getSelectedIDsFromMetadata(opts.stateMetaData, model.AutomaticReportBalanceIDsMetadataKey)

	if opts.message.GetText() != automaticReportSelectionDoneData {
		if opts.message.GetText() == automaticReportSelectionSelectAllData {
			selectedBalanceIDs = toggleAllSelectedIDs(selectedBalanceIDs, getIDs(opts.user.Balances))
		} else {
			balance := opts.user.GetBalance(opts.message.GetText())
			if balance == nil {
				return "", ErrBalanceNotFound
			}

			selectedBalanceIDs = toggleSelectedID(selectedBalanceIDs, balance.ID)
		}
		opts.stateMetaData.Add(model.AutomaticReportBalanceIDsMetadataKey, strings.Join(selectedBalanceIDs, ","))

		return model.ChooseBalanceFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        selectAutomaticReportBalancesMessage,
			UpdatedInlineKeyboard: getSelectionInlineKeyboardRows(opts.user.Balances, selectedBalanceIDs, 2),
		})
	}

	if len(selectedBalanceIDs) == 0 {
		return "", ErrAutomaticReportBalancesNotSelected
	}

	categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
		UserID: opts.user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("list user categories from store")
		return "", fmt.Errorf("list user categories from store: %w", err)
	}
	if len(categories) == 0 {
		return model.EndFlowStep, ErrCategoriesNotFound
	}

	return model.ChooseCategoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        selectAutomaticReportCategoriesMessage,
		UpdatedInlineKeyboard: getSelectionInlineKeyboardRows(categories, nil, 3),
	})
}

func (h *handlerService) handleChooseCategoryFlowStepForCreateAutomaticReport(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseCategoryFlowStepForCreateAutomaticReport").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	selectedCategoryIDs := getSelectedIDsFromMetadata(opts.stateMetaData, model.AutomaticReportCategoryIDsMetadataKey)

	if opts.message.GetText() != automaticReportSelectionDoneData {
		categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
			UserID: opts.user.ID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("list user categories from store")
			return "", fmt.Errorf("list user categories from store: %w", err)
		}

		if opts.message.GetText() == automaticReportSelectionSelectAllData {
			selectedCategoryIDs = toggleAllSelectedIDs(selectedCategoryIDs, getIDs(categories))
		} else {
			categoryIndex := slices.IndexFunc(categories, func(category model.Category) bool {
				return category.ID == opts.message.GetText()
			})
			if categoryIndex == -1 {
				return "", ErrCategoryNotFound
			}

			selectedCategoryIDs = toggleSelectedID(selectedCategoryIDs, categories[categoryIndex].ID)
		}
		opts.stateMetaData.Add(model.AutomaticReportCategoryIDsMetadataKey, strings.Join(selectedCategoryIDs, ","))

		return model.ChooseCategoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        selectAutomaticReportCategoriesMessage,
			UpdatedInlineKeyboard: getSelectionInlineKeyboardRows(categories, selectedCategoryIDs, 3),
		})
	}

	if len(selectedCategoryIDs) == 0 {
		return "", ErrAutomaticReportCategoriesNotSelected
	}

	name, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.AutomaticReportNameMetadataKey)
	if !ok {
		logger.Error().Msg("automatic report name not found in metadata")
		return "", fmt.Errorf("automatic report name not found in metadata")
	}

	periodFromMetadata, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.AutomaticReportPeriodMetadataKey)
	if !ok {
		logger.Error().Msg("automatic report period not found in metadata")
		return "", fmt.Errorf("automatic report period not found in metadata")
	}

	period, err := model.ParseAutomaticReportPeriod(periodFromMetadata)
	if err != nil {
		return "", fmt.Errorf("parse automatic report period: %w", err)
	}

	automaticReport := model.AutomaticReport{
		ID:          uuid.NewString(),
		BalanceIDs:  getSelectedIDsFromMetadata(opts.stateMetaData, model.AutomaticReportBalanceIDsMetadataKey),
		CategoryIDs: selectedCategoryIDs,
		Name:        name,
		Period:      period,
	}

	err = h.stores.AutomaticReport.Create(ctx, &automaticReport)
	if err != nil {
		logger.Error().Err(err).Msg("create automatic report in store")
		return "", fmt.Errorf("create automatic report in store: %w", err)
	}

	err = h.scheduleAutomaticReportExecution(ctx, automaticReport)
	if err != nil {
		logger.Error().Err(err).Msg("schedule automatic report execution")
		return "", fmt.Errorf("schedule automatic report execution: %w", err)
	}

	return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		UpdatedKeyboard: automaticReportKeyboardRows,
		UpdatedMessage:  fmt.Sprintf("Automatic report '%s' successfully created!\nFrequency: %s.", automaticReport.Name, automaticReport.Period.GetLabel()),
	})
}

// List Automatic Reports
func (h *handlerService) handleListAutomaticReportsFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleListAutomaticReportsFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	opts.stateMetaData.Add(model.PageMetadataKey, firstPage)

	message, keyboard, err := h.getListAutomaticReportsKeyboard(ctx, getAutomaticReportsKeyboardOptions{
		user: opts.user,
		page: firstPage,
	})
	if err != nil {
		if errs.IsExpected(err) {
			return model.EndFlowStep, err
		}
		logger.Error().Err(err).Msg("get list automatic reports keyboard")
		return "", fmt.Errorf("get list automatic reports keyboard: %w", err)
	}

	return model.ChooseAutomaticReportFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		Message:                 message,
		FormatMessageInMarkDown: true,
		InlineKeyboard:          keyboard,
	})
}

func (h *handlerService) handleChooseAutomaticReportFlowStepForList(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseAutomaticReportFlowStepForList").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if !isPaginationNeeded(opts.message.GetText()) {
		return "", nil
	}

	nextPage := calculateNextPage(opts.message.GetText(), opts.stateMetaData)
	opts.stateMetaData.Add(model.PageMetadataKey, nextPage)

	message, keyboard, err := h.getListAutomaticReportsKeyboard(ctx, getAutomaticReportsKeyboardOptions{
		user: opts.user,
		page: nextPage,
	})
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get list automatic reports keyboard")
		return "", fmt.Errorf("get list automatic reports keyboard: %w", err)
	}

	return model.ChooseAutomaticReportFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedInlineKeyboard:   keyboard,
		UpdatedMessage:          message,
	})
}

// Update Automatic Reports
func (h *handlerService) handleUpdateAutomaticReportFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateAutomaticReportFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	opts.stateMetaData.Add(model.PageMetadataKey, firstPage)

	keyboard, err := h.getAutomaticReportsKeyboard(ctx, getAutomaticReportsKeyboardOptions{
		user: opts.user,
		page: firstPage,
	})
	if err != nil {
		if errs.IsExpected(err) {
			return model.EndFlowStep, err
		}
		logger.Error().Err(err).Msg("get automatic reports keyboard")
		return "", fmt.Errorf("get automatic reports keyboard: %w", err)
	}

	return model.ChooseAutomaticReportToUpdateFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose automatic report to update:",
		InlineKeyboard: keyboard,
	})
}

func (h *handlerService) handleChooseAutomaticReportToUpdateFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseAutomaticReportToUpdateFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	messageText := opts.message.GetText()
	if isPaginationNeeded(messageText) {
		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData.Add(model.PageMetadataKey, nextPage)

		keyboard, err := h.getAutomaticReportsKeyboard(ctx, getAutomaticReportsKeyboardOptions{
			user: opts.user,
			page: nextPage,
		})
		if err != nil {
			if errs.IsExpected(err) {
				return "", err
			}
			logger.Error().Err(err).Msg("get automatic reports keyboard")
			return "", fmt.Errorf("get automatic reports keyboard: %w", err)
		}

		return model.ChooseAutomaticReportToUpdateFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        "Choose automatic report to update:",
			UpdatedInlineKeyboard: keyboard,
		})
	}

	automaticReport, err := h.getUserAutomaticReport(ctx, opts.user, messageText)
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get user automatic report")
		return "", fmt.Errorf("get user automatic report: %w", err)
	}

	opts.stateMetaData.Add(model.AutomaticReportIDMetadataKey, automaticReport.ID)
	return model.ChooseUpdateAutomaticReportOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Choose update automatic report option:",
		UpdatedInlineKeyboard: updateAutomaticReportOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseUpdateAutomaticReportOptionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseUpdateAutomaticReportOptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	automaticReport, err := h.getAutomaticReportFromMetadata(ctx, opts)
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get automatic report from metadata")
		return "", fmt.Errorf("get automatic report from metadata: %w", err)
	}

	switch opts.message.GetText() {
	case model.BotUpdateAutomaticReportNameCommand:
		return model.EnterAutomaticReportNameFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedMessage:          fmt.Sprintf("Enter updated automatic report name(Current: `%s`):", automaticReport.Name),
		})
	case model.BotUpdateAutomaticReportPeriodCommand:
		return model.ChooseAutomaticReportPeriodFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedMessage:          fmt.Sprintf("Select updated automatic report frequency(Current: `%s`):", automaticReport.Period.GetLabel()),
			UpdatedInlineKeyboard:   automaticReportPeriodKeyboard,
		})
	case model.BotUpdateAutomaticReportBalancesCommand:
		opts.stateMetaData.Add(model.AutomaticReportBalanceIDsMetadataKey, strings.Join(automaticReport.BalanceIDs, ","))

		return model.ChooseBalanceFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        selectAutomaticReportBalancesMessage,
			UpdatedInlineKeyboard: getSelectionInlineKeyboardRows(opts.user.Balances, automaticReport.BalanceIDs, 2),
		})
	case model.BotUpdateAutomaticReportCategoriesCommand:
		categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
			UserID: opts.user.ID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("list categories from store")
			return "", fmt.Errorf("list categories from store: %w", err)
		}
		if len(categories) == 0 {
			return "", ErrCategoriesNotFound
		}

		opts.stateMetaData.Add(model.AutomaticReportCategoryIDsMetadataKey, strings.Join(automaticReport.CategoryIDs, ","))

		return model.ChooseCategoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        selectAutomaticReportCategoriesMessage,
			UpdatedInlineKeyboard: getSelectionInlineKeyboardRows(categories, automaticReport.CategoryIDs, 3),
		})
	default:
		return "", fmt.Errorf("received unknown update automatic report option: %s", opts.message.GetText())
	}
}

func (h *handlerService) handleEnterAutomaticReportNameFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleEnterAutomaticReportNameFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	automaticReport, err := h.getAutomaticReportFromMetadata(ctx, opts)
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get automatic report from metadata")
		return "", fmt.Errorf("get automatic report from metadata: %w", err)
	}

	automaticReport.Name = opts.message.GetText()

	err = h.stores.AutomaticReport.Update(ctx, automaticReport)
	if err != nil {
		logger.Error().Err(err).Msg("update automatic report")
		return "", fmt.Errorf("update automatic report: %w", err)
	}

	return model.ChooseUpdateAutomaticReportOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		FormatMessageInMarkDown: true,
		Message: fmt.Sprintf(
			"Automatic report name successfully updated!\nNew name: `%s`\nPlease choose other update automatic report option or finish action by canceling it!",
			automaticReport.Name,
		),
		InlineKeyboard: updateAutomaticReportOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseAutomaticReportPeriodFlowStepForUpdate(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseAutomaticReportPeriodFlowStepForUpdate").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	period, err := model.ParseAutomaticReportPeriod(opts.message.GetText())
	if err != nil {
		return model.ChooseAutomaticReportPeriodFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
			ChatID:         opts.message.GetChatID(),
			Message:        "Invalid automatic report frequency. Please choose from the options below:",
			InlineKeyboard: automaticReportPeriodKeyboard,
		})
	}

	automaticReport, err := h.getAutomaticReportFromMetadata(ctx, opts)
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get automatic report from metadata")
		return "", fmt.Errorf("get automatic report from metadata: %w", err)
	}

	automaticReport.Period = period

	err = h.stores.AutomaticReport.Update(ctx, automaticReport)
	if err != nil {
		logger.Error().Err(err).Msg("update automatic report")
		return "", fmt.Errorf("update automatic report: %w", err)
	}

	err = h.rescheduleAutomaticReportExecution(ctx, *automaticReport)
	if err != nil {
		logger.Error().Err(err).Msg("reschedule automatic report execution")
		return "", fmt.Errorf("reschedule automatic report execution: %w", err)
	}

	return model.ChooseUpdateAutomaticReportOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage: fmt.Sprintf(
			"Automatic report frequency successfully updated!\nNew frequency: `%s`\nPlease choose other update automatic report option or finish action by canceling it!",
			automaticReport.Period.GetLabel(),
		),
		UpdatedInlineKeyboard: updateAutomaticReportOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseBalanceFlowStepForUpdateAutomaticReport(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseBalanceFlowStepForUpdateAutomaticReport").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	selectedBalanceIDs := getSelectedIDsFromMetadata(opts.stateMetaData, model.AutomaticReportBalanceIDsMetadataKey)

	if opts.message.GetText() != automaticReportSelectionDoneData {
		if opts.message.GetText() == automaticReportSelectionSelectAllData {
			selectedBalanceIDs = toggleAllSelectedIDs(selectedBalanceIDs, getIDs(opts.user.Balances))
		} else {
			balance := opts.user.GetBalance(opts.message.GetText())
			if balance == nil {
				return "", ErrBalanceNotFound
			}

			selectedBalanceIDs = toggleSelectedID(selectedBalanceIDs, balance.ID)
		}
		opts.stateMetaData.Add(model.AutomaticReportBalanceIDsMetadataKey, strings.Join(selectedBalanceIDs, ","))

		return model.ChooseBalanceFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        selectAutomaticReportBalancesMessage,
			UpdatedInlineKeyboard: getSelectionInlineKeyboardRows(opts.user.Balances, selectedBalanceIDs, 2),
		})
	}

	if len(selectedBalanceIDs) == 0 {
		return "", ErrAutomaticReportBalancesNotSelected
	}

	automaticReport, err := h.getAutomaticReportFromMetadata(ctx, opts)
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get automatic report from metadata")
		return "", fmt.Errorf("get automatic report from metadata: %w", err)
	}

	automaticReport.BalanceIDs = selectedBalanceIDs

	err = h.stores.AutomaticReport.Update(ctx, automaticReport)
	if err != nil {
		logger.Error().Err(err).Msg("update automatic report")
		return "", fmt.Errorf("update automatic report: %w", err)
	}

	return model.ChooseUpdateAutomaticReportOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		InlineMessageID: opts.message.GetInlineMessageID(),
		UpdatedMessage:  "Automatic report balances successfully updated!\nPlease choose other update automatic report option or finish action by canceling it!",

		UpdatedInlineKeyboard: updateAutomaticReportOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseCategoryFlowStepForUpdateAutomaticReport(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseCategoryFlowStepForUpdateAutomaticReport").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	selectedCategoryIDs := getSelectedIDsFromMetadata(opts.stateMetaData, model.AutomaticReportCategoryIDsMetadataKey)

	if opts.message.GetText() != automaticReportSelectionDoneData {
		categories, err := h.stores.Category.List(ctx, &ListCategoriesFilter{
			UserID: opts.user.ID,
		})
		if err != nil {
			logger.Error().Err(err).Msg("list user categories from store")
			return "", fmt.Errorf("list user categories from store: %w", err)
		}

		if opts.message.GetText() == automaticReportSelectionSelectAllData {
			selectedCategoryIDs = toggleAllSelectedIDs(selectedCategoryIDs, getIDs(categories))
		} else {
			categoryIndex := slices.IndexFunc(categories, func(category model.Category) bool {
				return category.ID == opts.message.GetText()
			})
			if categoryIndex == -1 {
				return "", ErrCategoryNotFound
			}

			selectedCategoryIDs = toggleSelectedID(selectedCategoryIDs, categories[categoryIndex].ID)
		}
		opts.stateMetaData.Add(model.AutomaticReportCategoryIDsMetadataKey, strings.Join(selectedCategoryIDs, ","))

		return model.ChooseCategoryFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        selectAutomaticReportCategoriesMessage,
			UpdatedInlineKeyboard: getSelectionInlineKeyboardRows(categories, selectedCategoryIDs, 3),
		})
	}

	if len(selectedCategoryIDs) == 0 {
		return "", ErrAutomaticReportCategoriesNotSelected
	}

	automaticReport, err := h.getAutomaticReportFromMetadata(ctx, opts)
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get automatic report from metadata")
		return "", fmt.Errorf("get automatic report from metadata: %w", err)
	}

	automaticReport.CategoryIDs = selectedCategoryIDs

	err = h.stores.AutomaticReport.Update(ctx, automaticReport)
	if err != nil {
		logger.Error().Err(err).Msg("update automatic report")
		return "", fmt.Errorf("update automatic report: %w", err)
	}

	return model.ChooseUpdateAutomaticReportOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedMessage:        "Automatic report categories successfully updated!\nPlease choose other update automatic report option or finish action by canceling it!",
		UpdatedInlineKeyboard: updateAutomaticReportOptionsKeyboard,
	})
}

// Delete Automatic Reports
func (h *handlerService) handleDeleteAutomaticReportFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleDeleteAutomaticReportFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	opts.stateMetaData.Add(model.PageMetadataKey, firstPage)

	keyboard, err := h.getAutomaticReportsKeyboard(ctx, getAutomaticReportsKeyboardOptions{
		user: opts.user,
		page: firstPage,
	})
	if err != nil {
		if errs.IsExpected(err) {
			return model.EndFlowStep, err
		}
		logger.Error().Err(err).Msg("get automatic reports keyboard")
		return "", fmt.Errorf("get automatic reports keyboard: %w", err)
	}

	return model.ChooseAutomaticReportToDeleteFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:         opts.message.GetChatID(),
		Message:        "Choose automatic report to delete:",
		InlineKeyboard: keyboard,
	})
}

func (h *handlerService) handleChooseAutomaticReportToDeleteFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseAutomaticReportToDeleteFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	messageText := opts.message.GetText()
	if isPaginationNeeded(messageText) {
		nextPage := calculateNextPage(messageText, opts.stateMetaData)
		opts.stateMetaData.Add(model.PageMetadataKey, nextPage)

		keyboard, err := h.getAutomaticReportsKeyboard(ctx, getAutomaticReportsKeyboardOptions{
			user: opts.user,
			page: nextPage,
		})
		if err != nil {
			if errs.IsExpected(err) {
				return "", err
			}
			logger.Error().Err(err).Msg("get automatic reports keyboard")
			return "", fmt.Errorf("get automatic reports keyboard: %w", err)
		}

		return model.ChooseAutomaticReportToDeleteFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                opts.message.GetChatID(),
			MessageID:             opts.message.GetMessageID(),
			InlineMessageID:       opts.message.GetInlineMessageID(),
			UpdatedMessage:        "Choose automatic report to delete:",
			UpdatedInlineKeyboard: keyboard,
		})
	}

	automaticReport, err := h.getUserAutomaticReport(ctx, opts.user, messageText)
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get user automatic report")
		return "", fmt.Errorf("get user automatic report: %w", err)
	}

	opts.stateMetaData.Add(model.AutomaticReportIDMetadataKey, automaticReport.ID)
	return model.ConfirmDeleteAutomaticReportFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                opts.message.GetChatID(),
		MessageID:             opts.message.GetMessageID(),
		InlineMessageID:       opts.message.GetInlineMessageID(),
		UpdatedInlineKeyboard: confirmationInlineKeyboardRows,
		UpdatedMessage:        automaticReport.GetDeletionMessage(),
	})
}

func (h *handlerService) handleConfirmDeleteAutomaticReportFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleConfirmDeleteAutomaticReportFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	confirmDeletion, err := strconv.ParseBool(opts.message.GetText())
	if err != nil {
		logger.Error().Err(err).Msg("parse callback data to bool")
		return "", fmt.Errorf("parse callback data to bool: %w", err)
	}
	if !confirmDeletion {
		logger.Info().Msg("user did not confirm automatic report deletion")
		return model.EndFlowStep, h.notifyCancellationAndShowKeyboard(opts.message, automaticReportKeyboardRows)
	}

	automaticReport, err := h.getAutomaticReportFromMetadata(ctx, opts)
	if err != nil {
		if errs.IsExpected(err) {
			return "", err
		}
		logger.Error().Err(err).Msg("get automatic report from metadata")
		return "", fmt.Errorf("get automatic report from metadata: %w", err)
	}

	err = h.stores.AutomaticReport.Delete(ctx, automaticReport.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete automatic report")
		return "", fmt.Errorf("delete automatic report: %w", err)
	}

	return model.EndFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:          opts.message.GetChatID(),
		MessageID:       opts.message.GetMessageID(),
		UpdatedKeyboard: automaticReportKeyboardRows,
		UpdatedMessage:  "Automatic report successfully deleted!",
	})
}

func (h *handlerService) getUserAutomaticReport(ctx context.Context, user *model.User, id string) (*model.AutomaticReport, error) {
	automaticReport, err := h.stores.AutomaticReport.Get(ctx, GetAutomaticReportFilter{
		ID: id,
	})
	if err != nil {
		return nil, fmt.Errorf("get automatic report from store: %w", err)
	}
	if automaticReport == nil || !slices.ContainsFunc(automaticReport.BalanceIDs, func(balanceID string) bool {
		return slices.Contains(user.GetBalancesIDs(), balanceID)
	}) {
		return nil, ErrAutomaticReportNotFound
	}

	return automaticReport, nil
}

func (h *handlerService) getAutomaticReportFromMetadata(ctx context.Context, opts flowProcessingOptions) (*model.AutomaticReport, error) {
	automaticReportID, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.AutomaticReportIDMetadataKey)
	if !ok {
		return nil, fmt.Errorf("automatic report ID not found in metadata")
	}

	return h.getUserAutomaticReport(ctx, opts.user, automaticReportID)
}

func (h *handlerService) scheduleAutomaticReportExecution(ctx context.Context, automaticReport model.AutomaticReport) error {
	err := h.stores.AutomaticReport.CreateScheduledReportExecution(ctx, &model.ScheduledReportExecution{
		ID:                uuid.NewString(),
		AutomaticReportID: automaticReport.ID,
		ExecutionDate:     automaticReport.Period.CalculateFirstExecutionDate(time.Now()),
	})
	if err != nil {
		return fmt.Errorf("create scheduled report execution in store: %w", err)
	}

	return nil
}

func (h *handlerService) rescheduleAutomaticReportExecution(ctx context.Context, automaticReport model.AutomaticReport) error {
	executions, err := h.stores.AutomaticReport.ListScheduledReportExecutions(ctx, ListScheduledReportExecutionFilter{
		AutomaticReportID: automaticReport.ID,
	})
	if err != nil {
		return fmt.Errorf("list scheduled report executions from store: %w", err)
	}

	for _, execution := range executions {
		err := h.stores.AutomaticReport.DeleteScheduledReportExecution(ctx, execution.ID)
		if err != nil {
			return fmt.Errorf("delete scheduled report execution from store: %w", err)
		}
	}

	return h.scheduleAutomaticReportExecution(ctx, automaticReport)
}

func getSelectedIDsFromMetadata(metadata model.Metadata, key model.MetadataKey) []string {
	selectedIDs, ok := model.GetTypedFromMetadata[string](metadata, key)
	if !ok || selectedIDs == "" {
		return nil
	}

	return strings.Split(selectedIDs, ",")
}

func getIDs[T identifiable](data []T) []string {
	ids := make([]string, 0, len(data))
	for _, entry := range data {
		ids = append(ids, entry.GetID())
	}

	return ids
}

func isAllSelected(selectedIDs, allIDs []string) bool {
	if len(allIDs) == 0 {
		return false
	}

	for _, id := range allIDs {
		if !slices.Contains(selectedIDs, id) {
			return false
		}
	}

	return true
}

func toggleAllSelectedIDs(selectedIDs, allIDs []string) []string {
	if isAllSelected(selectedIDs, allIDs) {
		return nil
	}

	return allIDs
}

func toggleSelectedID(selectedIDs []string, id string) []string {
	index := slices.Index(selectedIDs, id)
	if index == -1 {
		return append(selectedIDs, id)
	}

	return slices.Delete(selectedIDs, index, index+1)
}
