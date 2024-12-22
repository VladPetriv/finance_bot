package models

// Commands that we can received from bot.
const (
	// BotStartCommand represents the command to start the bot
	BotStartCommand string = "/start"
	// BotCreateBalanceCommand represents the command to create a new balance
	BotCreateBalanceCommand string = "Create Balance 💲"

	// BotBackCommand represents the command to go back to previous state
	BotBackCommand string = "Back ❌"
	// BotCreateCategoryCommand represents the command to create a new category
	BotCreateCategoryCommand string = "Create Category 📊"
	// BotListCategoriesCommand represents the command to list all categories
	BotListCategoriesCommand string = "List Categories 🗂️"
	// BotUpdateBalanceCommand represents the command to update balance
	BotUpdateBalanceCommand string = "Update Balance 💲"
	// BotUpdateBalanceAmountCommand represents the command to update balance amount
	BotUpdateBalanceAmountCommand string = "Update Balance Amount 💵"
	// BotUpdateBalanceCurrencyCommand represents the command to update balance currency
	BotUpdateBalanceCurrencyCommand string = "Update Balance Currency 💱"
	// BotGetBalanceCommand represents the command to get balance information
	BotGetBalanceCommand string = "Get Balance Info 🏦"
	// BotCreateOperationCommand represents the command to create a new operation
	BotCreateOperationCommand string = "Create Operation 🤔"
	// BotCreateIncomingOperationCommand represents the command to create an incoming operation
	BotCreateIncomingOperationCommand string = "Create Incoming Operation 🤑"
	// BotCreateSpendingOperationCommand represents the command to create a spending operation
	BotCreateSpendingOperationCommand string = "Create Spending Operation 💸"
	// BotUpdateOperationAmountCommand represents the command to update operation amount
	BotUpdateOperationAmountCommand string = "Update Operation Amount 💵"
	// BotGetOperationsHistory represents the command to get operations history
	BotGetOperationsHistory string = "Get Operations History 📖"
)

// AvailableCommands is a list of all available bot commands.
var AvailableCommands = []string{
	BotStartCommand, BotBackCommand, BotCreateCategoryCommand,
	BotListCategoriesCommand, BotUpdateBalanceCommand, BotUpdateBalanceAmountCommand,
	BotCreateOperationCommand, BotUpdateBalanceCurrencyCommand, BotGetBalanceCommand,
	BotCreateIncomingOperationCommand, BotCreateIncomingOperationCommand, BotCreateSpendingOperationCommand,
	BotUpdateOperationAmountCommand, BotGetOperationsHistory, BotCreateBalanceCommand,
}

// CommandToEvent maps bot commands to their corresponding events
var CommandToEvent = map[string]Event{
	BotStartCommand:         StartEvent,
	BotCreateBalanceCommand: CreateBalanceEvent,
}

// CommadToFistFlowStep maps commands to their initial flow steps
var CommadToFistFlowStep = map[string]FlowStep{
	BotCreateBalanceCommand: CreateBalanceFlowStep,
}
