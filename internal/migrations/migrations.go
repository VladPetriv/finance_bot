package migrations

import "github.com/lopezator/migrator"

var Migrations = []any{
	&migrator.MigrationNoTx{
		Name: "Init UUID extension",
		Func: initUUIDExtension,
	},
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
}
