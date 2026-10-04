package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
)

func (h *handlerService) handleGetUserSettingsFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleGetUserSettingsFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	return model.EndFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		Message:                 opts.user.Settings.GetDetails(),
		FormatMessageInMarkDown: true,
		Keyboard:                userSettingsKeyboardRows,
	})
}

func (h *handlerService) handleUpdateUserSettingsFlowStep(_ context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateUserSettingsFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	err := h.showCancelButton(opts.message.GetChatID(), "")
	if err != nil {
		logger.Error().Err(err).Msg("show cancel button")
		return "", fmt.Errorf("show cancel button: %w", err)
	}

	outputMessage := fmt.Sprintf("Current user settings:\n\n%s", opts.user.Settings.GetDetails())

	return model.ChooseUpdateUserSettingsOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		Message:                 outputMessage,
		FormatMessageInMarkDown: true,
		InlineKeyboard:          updateUserSettingsOptionsKeyboard,
	})
}

func (h *handlerService) handleChooseUpdateUserSettingsOptionFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleChooseUpdateUserSettingsOptionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	switch opts.message.GetText() {
	case model.BotUpdateUserAIParserCommand:
		toggleSettingResult := generateToggleSetting(toggleSettingConfig{
			Title:       "AI Parser Settings",
			Icon:        "🤖",
			Description: "AI Parser",
			IsEnabled:   opts.user.Settings.AIParserEnabled,
		})

		return model.UpdateAIParserEnabledUserSettingFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedInlineKeyboard:   []InlineKeyboardRow{toggleSettingResult.KeyboardRow},
			UpdatedMessage:          toggleSettingResult.OutputMessage,
		})
	case model.BotUpdateUserSubscriptionNotificationsCommand:
		toggleSettingResult := generateToggleSetting(toggleSettingConfig{
			Title:       "Subscription Notifications Settings",
			Icon:        "🔔",
			Description: "Subscription Notifications",
			IsEnabled:   opts.user.Settings.NotifyAboutSubscriptionPayments,
		})

		return model.UpdateSubscriptionNotificationUserSettingFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedInlineKeyboard:   []InlineKeyboardRow{toggleSettingResult.KeyboardRow},
			UpdatedMessage:          toggleSettingResult.OutputMessage,
		})
	case model.BotUpdateUserTimezoneCommand:
		return model.UpdateTimezoneUserSettingFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
			ChatID:                  opts.message.GetChatID(),
			MessageID:               opts.message.GetMessageID(),
			InlineMessageID:         opts.message.GetInlineMessageID(),
			FormatMessageInMarkDown: true,
			UpdatedInlineKeyboard:   timezoneOptionsKeyboard,
			UpdatedMessage: fmt.Sprintf(
				"🌍 *Timezone Settings*\nCurrent timezone: `%s`\nChoose your region below or enter any IANA timezone name (e.g. Europe/Kyiv).",
				opts.user.Settings.GetLocation().String(),
			),
		})
	default:
		logger.Debug().Str("option", opts.message.GetText()).Msg("received unknown update user settings option")
		return "", fmt.Errorf("received unknown update user settings option: %s", opts.message.GetText())
	}
}

func (h *handlerService) handleUpdateAIParserEnabledUserSettingFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateAIParserEnabledUserSettingFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	var state bool
	outputMessage := "AI Parser successfully *disabled*"
	if opts.message.GetText() == model.BotEnableCommand {
		state = true
		outputMessage = "AI Parser successfully *enabled*"
	}

	settings := opts.user.Settings
	settings.AIParserEnabled = state

	err := h.stores.User.UpdateSettings(ctx, settings)
	if err != nil {
		logger.Error().Err(err).Msg("update user settings in store")
		return "", fmt.Errorf("update user settings in store: %w", err)
	}

	return model.ChooseUpdateUserSettingsOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage:          fmt.Sprintf("%s\nPlease choose other update user settings option or finish action by canceling it!", outputMessage),
		UpdatedInlineKeyboard:   updateUserSettingsOptionsKeyboard,
	})
}

func (h *handlerService) handleUpdateSubscriptionNotificationUserSettingFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleUpdateSubscriptionNotificationUserSettingFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	var state bool
	outputMessage := "Subscription notification successfully *disabled*"
	if opts.message.GetText() == model.BotEnableCommand {
		state = true
		outputMessage = "Subscription notification successfully *enabled*"
	}

	settings := opts.user.Settings
	settings.NotifyAboutSubscriptionPayments = state

	err := h.stores.User.UpdateSettings(ctx, settings)
	if err != nil {
		logger.Error().Err(err).Msg("update user settings in store")
		return "", fmt.Errorf("update user settings in store: %w", err)
	}

	return model.ChooseUpdateUserSettingsOptionFlowStep, h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedMessage:          fmt.Sprintf("%s\nPlease choose other update user settings option or finish action by canceling it!", outputMessage),
		UpdatedInlineKeyboard:   updateUserSettingsOptionsKeyboard,
	})
}

func (h *handlerService) handleUpdateTimezoneUserSettingFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	return h.handleTimezoneRegionFlowStep(ctx, opts, timezoneFlowStepsConfig{
		regionStep: model.UpdateTimezoneUserSettingFlowStep,
		zoneStep:   model.ChooseTimezoneUserSettingFlowStep,
		onApplied:  h.sendUpdatedTimezoneMessage,
	})
}

func (h *handlerService) handleChooseTimezoneUserSettingFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	return h.handleTimezoneZoneFlowStep(ctx, opts, timezoneFlowStepsConfig{
		regionStep: model.UpdateTimezoneUserSettingFlowStep,
		zoneStep:   model.ChooseTimezoneUserSettingFlowStep,
		onApplied:  h.sendUpdatedTimezoneMessage,
	})
}

func (h *handlerService) handleEnterInitialTimezoneFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	return h.handleTimezoneRegionFlowStep(ctx, opts, timezoneFlowStepsConfig{
		regionStep: model.EnterInitialTimezoneFlowStep,
		zoneStep:   model.ChooseInitialTimezoneFlowStep,
		onApplied:  h.askInitialBalanceName,
	})
}

func (h *handlerService) handleChooseInitialTimezoneFlowStep(ctx context.Context, opts flowProcessingOptions) (model.FlowStep, error) {
	return h.handleTimezoneZoneFlowStep(ctx, opts, timezoneFlowStepsConfig{
		regionStep: model.EnterInitialTimezoneFlowStep,
		zoneStep:   model.ChooseInitialTimezoneFlowStep,
		onApplied:  h.askInitialBalanceName,
	})
}

type timezoneFlowStepsConfig struct {
	regionStep model.FlowStep
	zoneStep   model.FlowStep
	onApplied  func(opts flowProcessingOptions, timezone string) (model.FlowStep, error)
}

func (h *handlerService) handleTimezoneRegionFlowStep(ctx context.Context, opts flowProcessingOptions, config timezoneFlowStepsConfig) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleTimezoneRegionFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	region := opts.message.GetText()
	if !slices.Contains(model.TimezoneRegions, region) {
		return h.applyTimezone(ctx, opts, config.regionStep, config)
	}

	opts.stateMetaData.Add(model.TimezoneRegionMetadataKey, region)
	opts.stateMetaData.Add(model.PageMetadataKey, firstPage)

	return config.zoneStep, h.sendTimezonesOfRegion(opts, region, firstPage)
}

func (h *handlerService) handleTimezoneZoneFlowStep(ctx context.Context, opts flowProcessingOptions, config timezoneFlowStepsConfig) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.handleTimezoneZoneFlowStep").Logger()
	logger.Debug().Any("opts", opts).Msg("got args")

	if !isPaginationNeeded(opts.message.GetText()) {
		return h.applyTimezone(ctx, opts, config.zoneStep, config)
	}

	region, ok := model.GetTypedFromMetadata[string](opts.stateMetaData, model.TimezoneRegionMetadataKey)
	if !ok {
		logger.Error().Msg("timezone region not found in metadata")
		return "", fmt.Errorf("timezone region not found in metadata")
	}

	nextPage := calculateNextPage(opts.message.GetText(), opts.stateMetaData)
	opts.stateMetaData.Add(model.PageMetadataKey, nextPage)

	return config.zoneStep, h.sendTimezonesOfRegion(opts, region, nextPage)
}

func (h *handlerService) applyTimezone(ctx context.Context, opts flowProcessingOptions, currentStep model.FlowStep, config timezoneFlowStepsConfig) (model.FlowStep, error) {
	logger := h.logger.With().Str("name", "handlerService.applyTimezone").Logger()

	location, err := parseTimezone(opts.message.GetText())
	if err != nil {
		logger.Info().Err(err).Str("timezone", opts.message.GetText()).Msg("parse timezone")
		return currentStep, err
	}

	settings := opts.user.Settings
	settings.Timezone = location.String()

	err = h.stores.User.UpdateSettings(ctx, settings)
	if err != nil {
		logger.Error().Err(err).Msg("update user settings in store")
		return "", fmt.Errorf("update user settings in store: %w", err)
	}

	return config.onApplied(opts, settings.Timezone)
}

func (h *handlerService) sendTimezonesOfRegion(opts flowProcessingOptions, region string, page int) error {
	keyboard, err := getTimezonesKeyboard(region, page)
	if err != nil {
		return fmt.Errorf("get timezones keyboard: %w", err)
	}

	return h.apis.Messenger.UpdateMessage(UpdateMessageOptions{
		ChatID:                  opts.message.GetChatID(),
		MessageID:               opts.message.GetMessageID(),
		InlineMessageID:         opts.message.GetInlineMessageID(),
		FormatMessageInMarkDown: true,
		UpdatedInlineKeyboard:   keyboard,
		UpdatedMessage:          fmt.Sprintf("🌍 Choose your timezone in *%s* or enter any IANA timezone name:", region),
	})
}

func (h *handlerService) sendUpdatedTimezoneMessage(opts flowProcessingOptions, timezone string) (model.FlowStep, error) {
	return model.ChooseUpdateUserSettingsOptionFlowStep, h.apis.Messenger.SendWithKeyboard(SendWithKeyboardOptions{
		ChatID:                  opts.message.GetChatID(),
		FormatMessageInMarkDown: true,
		InlineKeyboard:          updateUserSettingsOptionsKeyboard,
		Message: fmt.Sprintf(
			"Timezone successfully updated to `%s`\nPlease choose other update user settings option or finish action by canceling it!",
			timezone,
		),
	})
}

func (h *handlerService) askInitialBalanceName(opts flowProcessingOptions, _ string) (model.FlowStep, error) {
	return model.CreateInitialBalanceFlowStep, h.apis.Messenger.SendMessage(opts.message.GetChatID(), "Please enter the name of your initial balance!:")
}

func parseTimezone(raw string) (*time.Location, error) {
	if raw == "" || raw == "Local" {
		return nil, ErrInvalidTimezone
	}

	location, err := time.LoadLocation(raw)
	if err != nil {
		return nil, ErrInvalidTimezone
	}

	return location, nil
}
