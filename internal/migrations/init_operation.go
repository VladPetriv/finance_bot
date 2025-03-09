package migrations

import "database/sql"

func initOperationTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TYPE operation_type AS ENUM ('incoming', 'spending', 'transfer', 'transfer_in', 'transfer_out');

		CREATE TABLE operations (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			category_id UUID NOT NULL,
			balance_id UUID NOT NULL,
			type operation_type,
			description VARCHAR(255) NOT NULL,
			amount VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);


      	ALTER TABLE ONLY operations
        	ADD CONSTRAINT operations_balance_id_fkey FOREIGN KEY (balance_id) REFERENCES balances(id);

        ALTER TABLE ONLY operations
        	ADD CONSTRAINT operations_category_id_fkey FOREIGN KEY (category_id) REFERENCES categories(id);
	`)

	return err
}
