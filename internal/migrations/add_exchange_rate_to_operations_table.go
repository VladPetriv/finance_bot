package migrations

import "database/sql"

func addExchangeRateToOperationsTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE operations ADD COLUMN exchange_rate VARCHAR(255);
	`)
	return err
}
