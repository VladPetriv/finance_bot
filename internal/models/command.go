package models

// Commands that we can received from bot.
const (
	// BotStartCommand represents the command to start the bot
	BotStartCommand string = "/start"

	// BotBalanceCommand represents the wrapper command for balances action
	BotBalanceCommand string = "💰 Balance"
	// BotCategoryCommand represents the wrapper command for categories action
	BotCategoryCommand string = "📂 Category"
	// BotOperationCommand represents the wrapper command for operations action
	BotOperationCommand string = "💸 Operation"
	// BotBalanceSubscriptionsCommand represents the wrapper command for managing balance subscription actions
	BotBalanceSubscriptionsCommand string = "🔄 Balance Subscriptions"

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
	BotCreateBalanceSubscriptionCommand string = "Create 📈"
	// BotListBalanceSubscriptionsCommand represents the command to list balance subscriptions
	BotListBalanceSubscriptionsCommand string = "List 📋"
	// BotUpdateBalanceSubscriptionCommand represents the command to update a balance subscription
	BotUpdateBalanceSubscriptionCommand string = "Update 📝"
	// BotUpdateBalanceSubscriptionNameCommand represents the command to update balance subscription name
	BotUpdateBalanceSubscriptionNameCommand string = "Update Name 📝"
	// BotUpdateBalanceSubscriptionAmountCommand represents the command to update balance subscription amount
	BotUpdateBalanceSubscriptionAmountCommand string = "Update Amount 📝"
	// BotUpdateBalanceSubscriptionCategoryCommand represents the command to update balance subscription category
	BotUpdateBalanceSubscriptionCategoryCommand string = "Update Category 🏷️"
	// BotUpdateBalanceSubscriptionPeriodCommand represents the command to update balance subscription period
	BotUpdateBalanceSubscriptionPeriodCommand string = "Update Period 📅"
	// BotDeleteBalanceSubscriptionCommand represents the command to delete a balance subscription
	BotDeleteBalanceSubscriptionCommand string = "Delete 🗑️"

	// BotShowMoreCommand represents the command to select more models.
	BotShowMoreCommand string = "Show More ➡️"
	// BotCancelCommand represents the command that will cancel the current flow
	BotCancelCommand string = "Cancel action ⬅️"
)

// AvailableCommands is a list of all available bot commands.
var AvailableCommands = []string{
	BotStartCommand,
	BotBalanceCommand, BotCategoryCommand, BotOperationCommand, BotBalanceSubscriptionsCommand,
	BotGetBalanceCommand, BotCreateBalanceCommand, BotUpdateBalanceCommand, BotDeleteBalanceCommand,
	BotUpdateBalanceNameCommand, BotUpdateBalanceAmountCommand, BotUpdateBalanceCurrencyCommand,
	BotCreateCategoryCommand, BotListCategoriesCommand, BotUpdateCategoryCommand, BotDeleteCategoryCommand,
	BotCreateOperationCommand, BotCreateIncomingOperationCommand, BotCreateSpendingOperationCommand, BotGetOperationsHistory, BotCreateTransferOperationCommand,
	BotDeleteOperationCommand, BotUpdateOperationCommand, BotUpdateOperationAmountCommand, BotUpdateOperationDescriptionCommand,
	BotUpdateOperationDateCommand, BotUpdateOperationCategoryCommand,
	BotCreateBalanceSubscriptionCommand, BotListBalanceSubscriptionsCommand, BotDeleteBalanceSubscriptionCommand, BotUpdateBalanceSubscriptionCommand,
	BotUpdateBalanceSubscriptionNameCommand, BotUpdateBalanceSubscriptionCategoryCommand, BotUpdateBalanceSubscriptionAmountCommand, BotUpdateBalanceSubscriptionPeriodCommand,
	BotCancelCommand, BotShowMoreCommand,
}

// CommandToEvent maps bot commands to their corresponding events
var CommandToEvent = map[string]Event{
	// General
	BotStartCommand:  StartEvent,
	BotCancelCommand: CancelEvent,

	// Wrappers
	BotBalanceCommand:              BalanceEvent,
	BotCategoryCommand:             CategoryEvent,
	BotOperationCommand:            OperationEvent,
	BotBalanceSubscriptionsCommand: BalanceSubscriptionEvent,

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
}

// CommandToFistFlowStep maps commands to their initial flow steps
var CommandToFistFlowStep = map[string]FlowStep{
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
}

// OperationCommandToOperationType maps operation commands to their corresponding operation types
var OperationCommandToOperationType = map[string]OperationType{
	BotCreateIncomingOperationCommand: OperationTypeIncoming,
	BotCreateSpendingOperationCommand: OperationTypeSpending,
	BotCreateTransferOperationCommand: OperationTypeTransfer,
}
