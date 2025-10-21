package migrations

import "database/sql"

func addParentOperationIDToOperationsTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE operations ADD COLUMN parent_operation_id VARCHAR(255);
	`)
	return err
}
