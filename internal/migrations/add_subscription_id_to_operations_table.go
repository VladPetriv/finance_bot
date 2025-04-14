package migrations

import "database/sql"

func addBalanceSubscriptionIDToOperationsTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE operations ADD COLUMN balance_subscription_id VARCHAR(255);
	`)
	return err
}
