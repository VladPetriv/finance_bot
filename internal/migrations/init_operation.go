package migrations

import "database/sql"

func initOperationTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TYPE operation_type AS ENUM ('incoming', 'spending', 'transfer', 'transfer_in', 'transfer_out');

		CREATE TABLE operations (
		    id VARCHAR(255) PRIMARY KEY,
			category_id VARCHAR(255) NULL,
			balance_id VARCHAR(255) NOT NULL,
			type operation_type,
			description VARCHAR(255) NOT NULL,
			amount VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

      	ALTER TABLE ONLY operations
        	ADD CONSTRAINT operations_balance_id_fkey FOREIGN KEY (balance_id) REFERENCES balances(id);
	`)

	return err
}
