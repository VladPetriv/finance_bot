package service

import "github.com/VladPetriv/finance_bot/internal/model"

var (
	defaultKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{model.BotBalanceCommand, model.BotCategoryCommand},
		},
		{
			Buttons: []string{model.BotOperationCommand, model.BotBalanceSubscriptionsCommand},
		},
		{
			Buttons: []string{model.BotAutomaticReportCommand},
		},
		{
			Buttons: []string{model.BotUserSettingsCommand},
		},
	}

	rowKeyboardWithCancelButtonOnly = []KeyboardRow{
		{
			Buttons: []string{model.BotCancelCommand},
		},
	}

	confirmationInlineKeyboardRows = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: "Yes",
					Data: "true",
				},
				{
					Text: "No",
					Data: "false",
				},
			},
		},
	}

	userSettingsKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{model.BotGetUserSettingsCommand, model.BotUpdateUserSettingsCommand},
		},
		{
			Buttons: []string{model.BotBackCommand},
		},
	}

	balanceKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{model.BotCreateBalanceCommand, model.BotGetBalanceCommand},
		},
		{
			Buttons: []string{model.BotUpdateBalanceCommand, model.BotDeleteBalanceCommand},
		},
		{
			Buttons: []string{model.BotBackCommand},
		},
	}

	categoryKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{model.BotCreateCategoryCommand, model.BotListCategoriesCommand},
		},
		{
			Buttons: []string{model.BotUpdateCategoryCommand, model.BotDeleteCategoryCommand},
		},
		{
			Buttons: []string{model.BotBackCommand},
		},
	}

	operationKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{model.BotCreateOperationCommand, model.BotGetOperationsHistory},
		},
		{
			Buttons: []string{model.BotUpdateOperationCommand, model.BotDeleteOperationCommand},
		},
		{
			Buttons: []string{model.BotBackCommand},
		},
	}

	balanceSubscriptionKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{model.BotCreateBalanceSubscriptionCommand, model.BotListBalanceSubscriptionsCommand},
		},
		{
			Buttons: []string{model.BotUpdateBalanceSubscriptionCommand, model.BotDeleteBalanceSubscriptionCommand},
		},
		{
			Buttons: []string{model.BotBackCommand},
		},
	}

	automaticReportKeyboardRows = []KeyboardRow{
		{
			Buttons: []string{model.BotCreateAutomaticReportCommand, model.BotListAutomaticReportsCommand},
		},
		{
			Buttons: []string{model.BotUpdateAutomaticReportCommand, model.BotDeleteAutomaticReportCommand},
		},
		{
			Buttons: []string{model.BotBackCommand},
		},
	}

	updateUserSettingsOptionsKeyboard = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateUserAIParserCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateUserSubscriptionNotificationsCommand,
				},
			},
		},
	}

	updateOperationOptionsKeyboardForIncomingAndSpendingOperations = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateOperationAmountCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateOperationDescriptionCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateOperationCategoryCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateOperationDateCommand,
				},
			},
		},
	}

	updateOperationOptionsKeyboardForTransferOperations = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateOperationAmountCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateOperationDateCommand,
				},
			},
		},
	}

	updateBalanceOptionsKeyboard = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateBalanceNameCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateBalanceAmountCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateBalanceCurrencyCommand,
				},
			},
		},
	}

	updateBalanceSubscriptionOptionsKeyboard = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateBalanceSubscriptionNameCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateBalanceSubscriptionAmountCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateBalanceSubscriptionCategoryCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateBalanceSubscriptionPeriodCommand,
				},
			},
		},
	}

	updateAutomaticReportOptionsKeyboard = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateAutomaticReportNameCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateAutomaticReportPeriodCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateAutomaticReportBalancesCommand,
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: model.BotUpdateAutomaticReportCategoriesCommand,
				},
			},
		},
	}

	automaticReportPeriodKeyboard = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: "🔁 Daily",
					Data: string(model.AutomaticReportPeriodDaily),
				},
				{
					Text: "🔁 Weekly",
					Data: string(model.AutomaticReportPeriodWeekly),
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: "🔁 Monthly",
					Data: string(model.AutomaticReportPeriodMonthly),
				},
				{
					Text: "🔁 Quarterly",
					Data: string(model.AutomaticReportPeriodQuarterly),
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: "🔁 Yearly",
					Data: string(model.AutomaticReportPeriodYearly),
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: "📌 End of day",
					Data: string(model.AutomaticReportPeriodEndOfDay),
				},
				{
					Text: "📌 End of week",
					Data: string(model.AutomaticReportPeriodEndOfWeek),
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: "📌 End of month",
					Data: string(model.AutomaticReportPeriodEndOfMonth),
				},
				{
					Text: "📌 End of quarter",
					Data: string(model.AutomaticReportPeriodEndOfQuarter),
				},
			},
		},
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: "📌 End of year",
					Data: string(model.AutomaticReportPeriodEndOfYear),
				},
			},
		},
	}

	operationHistoryPeriodKeyboard = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: string(model.CreationPeriodDay),
				},
				{
					Text: string(model.CreationPeriodWeek),
				},
				{
					Text: string(model.CreationPeriodMonth),
				},
				{
					Text: string(model.CreationPeriodYear),
				},
			},
		},
	}

	balanceSubscriptionFrequencyKeyboard = []InlineKeyboardRow{
		{
			Buttons: []InlineKeyboardButton{
				{
					Text: string(model.SubscriptionPeriodWeekly),
				},
				{
					Text: string(model.SubscriptionPeriodMonthly),
				},
				{
					Text: string(model.SubscriptionPeriodYearly),
				},
			},
		},
	}
)
