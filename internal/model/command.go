package model

// Commands that we can received from bot.
const (
	// BotStartCommand represents the command to start the bot
	BotStartCommand string = "/start"

	// BotUserSettingsCommand represents the command to manage user settings
	BotUserSettingsCommand string = "⚙️ User Settings"
	// BotBalanceCommand represents the wrapper command for balances action
	BotBalanceCommand string = "💰 Balance"
	// BotCategoryCommand represents the wrapper command for categories action
	BotCategoryCommand string = "📂 Category"
	// BotOperationCommand represents the wrapper command for operations action
	BotOperationCommand string = "💸 Operation"
	// BotBalanceSubscriptionsCommand represents the wrapper command for managing balance subscription actions
	BotBalanceSubscriptionsCommand string = "🔄 Balance Subscriptions"
	// BotAutomaticReportCommand represents the wrapper command for automatic reports action
	BotAutomaticReportCommand string = "📊 Automatic Reports"

	// BotGetUserSettingsCommand represents the command to get user settings
	BotGetUserSettingsCommand string = "Get User Settings ⚙️"
	// BotUpdateUserSettingsCommand represents the command to update user settings
	BotUpdateUserSettingsCommand string = "Update User Settings 🔧"
	// BotUpdateUserAICommand represents the command to update AI parser settings
	BotUpdateUserAIParserCommand string = "Update AI Parser 🤖"
	// BotUpdateUserSubscriptionNotificationsCommand represents the command to update subscription notification settings
	BotUpdateUserSubscriptionNotificationsCommand string = "Update Subscription Notifications 💳"

	// BotCreateBalanceCommand represents the command to create a new balance
	BotCreateBalanceCommand string = "Create Balance 💰"
	// BotUpdateBalanceCommand represents the command to update balance
	BotUpdateBalanceCommand string = "Update Balance 📈"
	// BotUpdateBalanceNameCommand represents the command to update balance name
	BotUpdateBalanceNameCommand string = "Update Balance Name 📝"
	// BotUpdateBalanceAmountCommand represents the command to update balance amount
	BotUpdateBalanceAmountCommand string = "Update Balance Amount 💰"
	// BotUpdateBalanceCurrencyCommand represents the command to update balance currency
	BotUpdateBalanceCurrencyCommand string = "Update Balance Currency 💵"
	// BotGetBalanceCommand represents the command to get information about specific balance
	BotGetBalanceCommand string = "Get Balance Info 📊"
	// BotDeleteBalanceCommand represents the command to delete a balance
	BotDeleteBalanceCommand string = "Delete Balance ❌"

	// BotCreateCategoryCommand represents the command to create a new category
	BotCreateCategoryCommand string = "Create Category ✨"
	// BotListCategoriesCommand represents the command to list all categories
	BotListCategoriesCommand string = "List Categories 📋"
	// BotUpdateCategoryCommand represents the command to update a category
	BotUpdateCategoryCommand string = "Update Category ✏️"
	// BotDeleteCategoryCommand represents the command to delete a category
	BotDeleteCategoryCommand string = "Delete Category ❌"

	// BotCreateOperationCommand represents the command to create a new operation
	BotCreateOperationCommand string = "Create Operation 🤔"
	// BotCreateIncomingOperationCommand represents the command to create an incoming operation
	BotCreateIncomingOperationCommand string = "Incoming 🤑"
	// BotCreateSpendingOperationCommand represents the command to create a spending operation
	BotCreateSpendingOperationCommand string = "Spending 💸"
	// BotCreateTransferOperationCommand represents the command to create a transfer operation
	BotCreateTransferOperationCommand string = "Transfer ➡️"
	// BotGetOperationsHistory represents the command to get operations history
	BotGetOperationsHistory string = "Get Operations History 📖"
	// BotDeleteOperationCommand represents the command to delete an operation
	BotDeleteOperationCommand string = "Delete Operation ❌"
	// BotUpdateOperationCommand represents the command to update an operation
	BotUpdateOperationCommand string = "Update Operation ✏️"
	// BotUpdateOperationAmountCommand represents the command to update operation amount
	BotUpdateOperationAmountCommand string = "Update Amount 💰"
	// BotUpdateOperationDescriptionCommand represents the command to update operation description
	BotUpdateOperationDescriptionCommand string = "Update Description 📝"
	// BotUpdateOperationDateCommand represents the command to update operation date
	BotUpdateOperationDateCommand string = "Update Date 📅"
	// BotUpdateOperationCategoryCommand represents the command to update operation category
	BotUpdateOperationCategoryCommand string = "Update Category 🏷️"

	// BotCreateBalanceSubscriptionCommand represents the command to create a balance subscription
	BotCreateBalanceSubscriptionCommand string = "Create Balance Subscription 📈"
	// BotListBalanceSubscriptionsCommand represents the command to list balance subscriptions
	BotListBalanceSubscriptionsCommand string = "List Balance Subscriptions 📋"
	// BotUpdateBalanceSubscriptionCommand represents the command to update a balance subscription
	BotUpdateBalanceSubscriptionCommand string = "Update Balance Subscription 📝"
	// BotUpdateBalanceSubscriptionNameCommand represents the command to update balance subscription name
	BotUpdateBalanceSubscriptionNameCommand string = "Update Balance Subscription Name 📝"
	// BotUpdateBalanceSubscriptionAmountCommand represents the command to update balance subscription amount
	BotUpdateBalanceSubscriptionAmountCommand string = "Update Balance Subscription Amount 📝"
	// BotUpdateBalanceSubscriptionCategoryCommand represents the command to update balance subscription category
	BotUpdateBalanceSubscriptionCategoryCommand string = "Update Balance Subscription Category 🏷️"
	// BotUpdateBalanceSubscriptionPeriodCommand represents the command to update balance subscription period
	BotUpdateBalanceSubscriptionPeriodCommand string = "Update Balance Subscription Period 📅"
	// BotDeleteBalanceSubscriptionCommand represents the command to delete a balance subscription
	BotDeleteBalanceSubscriptionCommand string = "Delete Balance Subscription 🗑️"

	// BotCreateAutomaticReportCommand represents the command to create a new automatic report
	BotCreateAutomaticReportCommand string = "Create Automatic Report 📈"
	// BotListAutomaticReportsCommand represents the command to list all automatic reports
	BotListAutomaticReportsCommand string = "List Automatic Reports 📋"
	// BotUpdateAutomaticReportCommand represents the command to update an automatic report
	BotUpdateAutomaticReportCommand string = "Update Automatic Report ✏️"
	// BotDeleteAutomaticReportCommand represents the command to delete an automatic report
	BotDeleteAutomaticReportCommand string = "Delete Automatic Report 🗑️"

	// BotUpdateAutomaticReportNameCommand represents the command to update automatic report name
	BotUpdateAutomaticReportNameCommand string = "Update Report Name 🏷️"
	// BotUpdateAutomaticReportPeriodCommand represents the command to update automatic report period
	BotUpdateAutomaticReportPeriodCommand string = "Update Report Period 📅"
	// BotUpdateAutomaticReportBalancesCommand represents the command to update automatic report balances
	BotUpdateAutomaticReportBalancesCommand string = "Update Report Balances 💰"
	// BotUpdateAutomaticReportCategoriesCommand represents the command to update automatic report categories
	BotUpdateAutomaticReportCategoriesCommand string = "Update Report Categories 🗂️"

	// BotPreviousCommand represents the command to go back to the previous page
	BotPreviousCommand string = "Previous ⬅️"
	// BotNextCommand represents the command to go to the next page
	BotNextCommand string = "Next ➡️"
	// BotCancelCommand represents the command that will cancel the current flow
	BotCancelCommand string = "Cancel action ⬅️"
	// BotBackCommand represents the command to go back to the previous menu
	BotBackCommand string = "Back ⬅️"
	// BotEnableCommand represents the command to enable any state
	BotEnableCommand string = "Enable ✅"
	// BotDisableCommand represents the command to disable any state
	BotDisableCommand string = "Disable ❌"
)

// AvailableCommands is a list of all available bot commands.
var AvailableCommands = []string{
	// General commands
	BotEnableCommand, BotDisableCommand, BotStartCommand, BotCancelCommand,
	BotBackCommand, BotNextCommand, BotPreviousCommand,

	// Wrappers
	BotUserSettingsCommand, BotBalanceCommand, BotCategoryCommand, BotOperationCommand,
	BotBalanceSubscriptionsCommand, BotAutomaticReportCommand,

	// User settings
	BotGetUserSettingsCommand, BotUpdateUserSettingsCommand,

	// Balance
	BotGetBalanceCommand, BotCreateBalanceCommand, BotUpdateBalanceCommand, BotDeleteBalanceCommand,
	BotUpdateBalanceNameCommand, BotUpdateBalanceAmountCommand, BotUpdateBalanceCurrencyCommand,

	// Category
	BotCreateCategoryCommand, BotListCategoriesCommand, BotUpdateCategoryCommand, BotDeleteCategoryCommand,

	// Operation
	BotCreateOperationCommand, BotCreateIncomingOperationCommand, BotCreateSpendingOperationCommand, BotGetOperationsHistory,
	BotCreateTransferOperationCommand, BotDeleteOperationCommand, BotUpdateOperationCommand, BotUpdateOperationAmountCommand,
	BotUpdateOperationDescriptionCommand, BotUpdateOperationDateCommand, BotUpdateOperationCategoryCommand,

	// Balance Subscription
	BotCreateBalanceSubscriptionCommand, BotListBalanceSubscriptionsCommand, BotDeleteBalanceSubscriptionCommand, BotUpdateBalanceSubscriptionCommand,
	BotUpdateBalanceSubscriptionNameCommand, BotUpdateBalanceSubscriptionCategoryCommand, BotUpdateBalanceSubscriptionAmountCommand, BotUpdateBalanceSubscriptionPeriodCommand,

	// Automatic Report
	BotCreateAutomaticReportCommand, BotListAutomaticReportsCommand, BotUpdateAutomaticReportCommand, BotDeleteAutomaticReportCommand,
	BotUpdateAutomaticReportNameCommand, BotUpdateAutomaticReportPeriodCommand, BotUpdateAutomaticReportBalancesCommand, BotUpdateAutomaticReportCategoriesCommand,
}

// CommandToEvent maps bot commands to their corresponding events
var CommandToEvent = map[string]Event{
	// General
	BotStartCommand:  StartEvent,
	BotCancelCommand: CancelEvent,
	BotBackCommand:   BackEvent,

	// Wrappers
	BotUserSettingsCommand:         UserSettingsEvent,
	BotBalanceCommand:              BalanceEvent,
	BotCategoryCommand:             CategoryEvent,
	BotOperationCommand:            OperationEvent,
	BotBalanceSubscriptionsCommand: BalanceSubscriptionEvent,
	BotAutomaticReportCommand:      AutomaticReportEvent,

	// User Settings
	BotGetUserSettingsCommand:    GetUserSettingsEvent,
	BotUpdateUserSettingsCommand: UpdateUserSettingsEvent,

	// Balance
	BotCreateBalanceCommand: CreateBalanceEvent,
	BotUpdateBalanceCommand: UpdateBalanceEvent,
	BotGetBalanceCommand:    GetBalanceEvent,
	BotDeleteBalanceCommand: DeleteBalanceEvent,

	// Category
	BotCreateCategoryCommand: CreateCategoryEvent,
	BotListCategoriesCommand: ListCategoriesEvent,
	BotUpdateCategoryCommand: UpdateCategoryEvent,
	BotDeleteCategoryCommand: DeleteCategoryEvent,

	// Operation
	BotCreateOperationCommand: CreateOperationEvent,
	BotGetOperationsHistory:   GetOperationsHistoryEvent,
	BotDeleteOperationCommand: DeleteOperationEvent,
	BotUpdateOperationCommand: UpdateOperationEvent,

	// Balance Subscriptions
	BotCreateBalanceSubscriptionCommand: CreateBalanceSubscriptionEvent,
	BotListBalanceSubscriptionsCommand:  ListBalanceSubscriptionEvent,
	BotUpdateBalanceSubscriptionCommand: UpdateBalanceSubscriptionEvent,
	BotDeleteBalanceSubscriptionCommand: DeleteBalanceSubscriptionEvent,

	// Automatic Report
	BotCreateAutomaticReportCommand: CreateAutomaticReportEvent,
	BotListAutomaticReportsCommand:  ListAutomaticReportsEvent,
	BotUpdateAutomaticReportCommand: UpdateAutomaticReportEvent,
	BotDeleteAutomaticReportCommand: DeleteAutomaticReportEvent,
}

// CommandToFistFlowStep maps commands to their initial flow steps
var CommandToFistFlowStep = map[string]FlowStep{
	// User
	BotGetUserSettingsCommand:    GetUserSettingsFlowStep,
	BotUpdateUserSettingsCommand: UpdateUserSettingsFlowStep,

	// Balance
	BotCreateBalanceCommand: CreateBalanceFlowStep,
	BotUpdateBalanceCommand: UpdateBalanceFlowStep,
	BotGetBalanceCommand:    GetBalanceFlowStep,
	BotDeleteBalanceCommand: DeleteBalanceFlowStep,

	// Category
	BotCreateCategoryCommand: CreateCategoryFlowStep,
	BotListCategoriesCommand: ListCategoriesFlowStep,
	BotUpdateCategoryCommand: UpdateCategoryFlowStep,
	BotDeleteCategoryCommand: DeleteCategoryFlowStep,

	// Operation
	BotCreateOperationCommand: CreateOperationFlowStep,
	BotGetOperationsHistory:   GetOperationsHistoryFlowStep,
	BotDeleteOperationCommand: DeleteOperationFlowStep,
	BotUpdateOperationCommand: UpdateOperationFlowStep,

	// Balance Subscription
	BotCreateBalanceSubscriptionCommand: CreateBalanceSubscriptionFlowStep,
	BotListBalanceSubscriptionsCommand:  ListBalanceSubscriptionFlowStep,
	BotUpdateBalanceSubscriptionCommand: UpdateBalanceSubscriptionFlowStep,
	BotDeleteBalanceSubscriptionCommand: DeleteBalanceSubscriptionFlowStep,

	// Automatic Report
	BotCreateAutomaticReportCommand: CreateAutomaticReportFlowStep,
	BotListAutomaticReportsCommand:  ListAutomaticReportsFlowStep,
	BotUpdateAutomaticReportCommand: UpdateAutomaticReportFlowStep,
	BotDeleteAutomaticReportCommand: DeleteAutomaticReportFlowStep,
}

// OperationCommandToOperationType maps operation commands to their corresponding operation types
var OperationCommandToOperationType = map[string]OperationType{
	BotCreateIncomingOperationCommand: OperationTypeIncoming,
	BotCreateSpendingOperationCommand: OperationTypeSpending,
	BotCreateTransferOperationCommand: OperationTypeTransfer,
}
