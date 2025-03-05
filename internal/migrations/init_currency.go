package migrations

import "database/sql"

func initCurrencyTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE currencies (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			code VARCHAR(255) NOT NULL,
			symbol VARCHAR(255) NOT NULL,
		);


		ALTER TABLE ONLY currencies
    		ADD CONSTRAINT currencies_code_fkey UNIQUE (code);
	`)

	return err
}
