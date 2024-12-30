package models

// Commands that we can received from bot.
const (
	// BotStartCommand represents the command to start the bot
	BotStartCommand string = "/start"
	// BotCreateBalanceCommand represents the command to create a new balance
	BotCreateBalanceCommand string = "Create Balance 💰"
	// BotUpdateBalanceCommand represents the command to update balance
	BotUpdateBalanceCommand string = "Update Balance 📈"
	// BotGetBalanceCommand represents the command to get information about specific balance
	BotGetBalanceCommand string = "Get Balance Info 📊"
	// BotCreateCategoryCommand represents the command to create a new category
	BotCreateCategoryCommand string = "Create Category ✨"
	// BotListCategoriesCommand represents the command to list all categories
	BotListCategoriesCommand string = "List Categories 📋"

	// BotCreateOperationCommand represents the command to create a new operation
	BotCreateOperationCommand string = "Create Operation 🤔"
	// BotCreateIncomingOperationCommand represents the command to create an incoming operation
	BotCreateIncomingOperationCommand string = "Create Incoming Operation 🤑"
	// BotCreateSpendingOperationCommand represents the command to create a spending operation
	BotCreateSpendingOperationCommand string = "Create Spending Operation 💸"
	// BotGetOperationsHistory represents the command to get operations history
	BotGetOperationsHistory string = "Get Operations History 📖"

	// BotBackCommand represents the command to go back to previous state
	BotBackCommand string = "Back ❌"
)

// AvailableCommands is a list of all available bot commands.
var AvailableCommands = []string{
	BotStartCommand,
	BotGetBalanceCommand, BotCreateBalanceCommand, BotUpdateBalanceCommand,
	BotCreateCategoryCommand, BotListCategoriesCommand,
	BotCreateOperationCommand, BotCreateIncomingOperationCommand, BotCreateSpendingOperationCommand, BotGetOperationsHistory,
	BotBackCommand,
}

// CommandToEvent maps bot commands to their corresponding events
var CommandToEvent = map[string]Event{
	BotStartCommand:           StartEvent,
	BotCreateBalanceCommand:   CreateBalanceEvent,
	BotUpdateBalanceCommand:   UpdateBalanceEvent,
	BotGetBalanceCommand:      GetBalanceEvent,
	BotCreateCategoryCommand:  CreateCategoryEvent,
	BotListCategoriesCommand:  ListCategoriesEvent,
	BotCreateOperationCommand: CreateOperationEvent,
	BotGetOperationsHistory:   GetOperationsHistoryEvent,
	BotBackCommand:            BackEvent,
}

// CommadToFistFlowStep maps commands to their initial flow steps
var CommadToFistFlowStep = map[string]FlowStep{
	BotCreateBalanceCommand:   CreateBalanceFlowStep,
	BotUpdateBalanceCommand:   UpdateBalanceFlowStep,
	BotGetBalanceCommand:      GetBalanceFlowStep,
	BotCreateCategoryCommand:  CreateCategoryFlowStep,
	BotListCategoriesCommand:  ListCategoriesFlowStep,
	BotCreateOperationCommand: CreateOperationFlowStep,
	BotGetOperationsHistory:   GetOperationsHistoryFlowStep,
}

// OperationCommandToOperationType maps operation commands to their corresponding operation types
var OperationCommandToOperationType = map[string]OperationType{
	BotCreateIncomingOperationCommand: OperationTypeIncoming,
	BotCreateSpendingOperationCommand: OperationTypeSpending,
}
