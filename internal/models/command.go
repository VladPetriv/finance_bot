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

	// BotCreateBalanceCommand represents the command to create a new balance
	BotCreateBalanceCommand string = "Create Balance 💰"
	// BotUpdateBalanceCommand represents the command to update balance
	BotUpdateBalanceCommand string = "Update Balance 📈"
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
	// BotCreateagsingOperationCommand represents the command to create a transfer operation
	BotCreateTransferOperationCommand string = "Transfer ➡️"
	// BotGetOperationsHistory represents the command to get operations history
	BotGetOperationsHistory string = "Get Operations History 📖"
	// BotDeleteOperationCommand represents the command to delete an operation
	BotDeleteOperationCommand string = "Delete Operation ❌"
	// BotShowMoreOperationsForDeleteCommand represents the command to select more operations
	BotShowMoreOperationsForDeleteCommand string = "Show More Operations For Delete ➡️"
	// BotUpdateOperationCommand represents the command to update a category
	BotUpdateOperationCommand string = "Update Operation ✏️"

	// BotCancelCommand represents the command that will cancel the current flow
	BotCancelCommand string = "Cancel action ⬅️"
)

// AvailableCommands is a list of all available bot commands.
var AvailableCommands = []string{
	BotStartCommand,
	BotBalanceCommand, BotCategoryCommand, BotOperationCommand,
	BotGetBalanceCommand, BotCreateBalanceCommand, BotUpdateBalanceCommand, BotDeleteBalanceCommand,
	BotCreateCategoryCommand, BotListCategoriesCommand, BotUpdateCategoryCommand, BotDeleteCategoryCommand,
	BotCreateOperationCommand, BotCreateIncomingOperationCommand, BotCreateSpendingOperationCommand, BotGetOperationsHistory, BotCreateTransferOperationCommand,
	BotDeleteOperationCommand, BotShowMoreOperationsForDeleteCommand, BotUpdateOperationCommand,
	BotCancelCommand,
}

// CommandToEvent maps bot commands to their corresponding events
var CommandToEvent = map[string]Event{
	// General
	BotStartCommand:  StartEvent,
	BotCancelCommand: CancelEvent,

	// Wrappers
	BotBalanceCommand:   BalanceEvent,
	BotCategoryCommand:  CategoryEvent,
	BotOperationCommand: OperationEvent,

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
}

// OperationCommandToOperationType maps operation commands to their corresponding operation types
var OperationCommandToOperationType = map[string]OperationType{
	BotCreateIncomingOperationCommand: OperationTypeIncoming,
	BotCreateSpendingOperationCommand: OperationTypeSpending,
	BotCreateTransferOperationCommand: OperationTypeTransfer,
}
