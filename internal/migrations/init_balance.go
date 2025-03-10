package migrations

import "database/sql"

func initBalanceTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE balances (
		    id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NULL,
			currency_id VARCHAR(255) NULL,
			name VARCHAR(255) NOT NULL,
			amount VARCHAR(255) NOT NULL
		);

		ALTER TABLE ONLY balances
    		ADD CONSTRAINT balances_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
	`)

	return err
}
