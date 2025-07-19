package migrations

import "github.com/lopezator/migrator"

// Migrations contains all the migrations to be applied to the database.
var Migrations = []any{
	&migrator.MigrationNoTx{
		Name: "Init currency table",
		Func: initCurrencyTable,
	},
	&migrator.MigrationNoTx{
		Name: "Init user table",
		Func: initUserTable,
	},
	&migrator.MigrationNoTx{
		Name: "Init balance table",
		Func: initBalanceTable,
	},
	&migrator.MigrationNoTx{
		Name: "Init category table",
		Func: initCategoryTable,
	},
	&migrator.MigrationNoTx{
		Name: "Init operation table",
		Func: initOperationTable,
	},
	&migrator.MigrationNoTx{
		Name: "Init state table",
		Func: initStateTable,
	},
	&migrator.MigrationNoTx{
		Name: "Init user settings table",
		Func: initUserSettingsTable,
	},
	&migrator.MigrationNoTx{
		Name: "Init balance_subscription table",
		Func: initBalanceSubscriptionTable,
	},
	&migrator.Migration{
		Name: "Add balance_subscription_id column to operations table",
		Func: addBalanceSubscriptionIDToOperationsTable,
	},
	&migrator.MigrationNoTx{
		Name: "Init scheduled operations table",
		Func: initScheduledOperationTable,
	},
	&migrator.Migration{
		Name: "Add chat id to users table",
		Func: addChatIDToUsersTable,
	},
	&migrator.Migration{
		Name: "Add notify_about_subscription_payments to user_settings table",
		Func: addNotifyAboutSubscriptionPaymentsToUserSettings,
	},
	&migrator.Migration{
		Name: "Add notified to scheduled_operations table",
		Func: addNotifiedToScheduledOperationsTable,
	},
	&migrator.Migration{
		Name: "Add created_at and updated_at column to balances table",
		Func: addCreatedAtAndUpdatedAtColumnsToBalancesTable,
	},
}
