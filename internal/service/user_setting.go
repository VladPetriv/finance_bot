package service

import (
	"context"
	"fmt"

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
