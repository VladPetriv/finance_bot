package migrations

import "database/sql"

func initBalanceTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE balances (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			currency_id UUID NOT NULL,
			name VARCHAR(255) NOT NULL,
			amount VARCHAR(255) NOT NULL
		);


		ALTER TABLE ONLY balances
    		ADD CONSTRAINT balances_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

		ALTER TABLE ONLY balances
    		ADD CONSTRAINT balances_currency_id_fkey FOREIGN KEY (currency_id) REFERENCES currencies(id);
	`)

	return err
}
