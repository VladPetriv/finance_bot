package migrations

import "database/sql"

func initCurrencyTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE currencies (
		    id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			code VARCHAR(255) NOT NULL,
			symbol VARCHAR(255) NOT NULL
		);


		ALTER TABLE ONLY currencies
    		ADD CONSTRAINT currencies_code_fkey UNIQUE (code);
	`)

	return err
}
