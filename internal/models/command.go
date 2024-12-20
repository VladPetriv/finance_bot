package models

// Commands that we can received from bot.
const (
	BotStartCommand         string = "/start"
	BotCreateBalanceCommand string = "Create Balance 💲"

	BotBackCommand                    string = "Back ❌"
	BotCreateCategoryCommand          string = "Create Category 📊"
	BotListCategoriesCommand          string = "List Categories 🗂️"
	BotUpdateBalanceCommand           string = "Update Balance 💲"
	BotUpdateBalanceAmountCommand     string = "Update Balance Amount 💵"
	BotUpdateBalanceCurrencyCommand   string = "Update Balance Currency 💱"
	BotGetBalanceCommand              string = "Get Balance Info 🏦"
	BotCreateOperationCommand         string = "Create Operation 🤔"
	BotCreateIncomingOperationCommand string = "Create Incoming Operation 🤑"
	BotCreateSpendingOperationCommand string = "Create Spending Operation 💸"
	BotUpdateOperationAmountCommand   string = "Update Operation Amount 💵"
	BotGetOperationsHistory           string = "Get Operations History 📖"
)

// AvailableCommands is a list of all available bot commands.
var AvailableCommands = []string{
	BotStartCommand, BotBackCommand, BotCreateCategoryCommand,
	BotListCategoriesCommand, BotUpdateBalanceCommand, BotUpdateBalanceAmountCommand,
	BotCreateOperationCommand, BotUpdateBalanceCurrencyCommand, BotGetBalanceCommand,
	BotCreateIncomingOperationCommand, BotCreateIncomingOperationCommand, BotCreateSpendingOperationCommand,
	BotUpdateOperationAmountCommand, BotGetOperationsHistory, BotCreateBalanceCommand,
}

var CommandToEvent = map[string]Event{
	BotStartCommand:         StartEvent,
	BotCreateBalanceCommand: CreateBalanceEvent,
}

var CommadToFistFlowStep = map[string]FlowStep{
	BotCreateBalanceCommand: CreateBalanceFlowStep,
}
