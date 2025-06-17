package migrations

import "database/sql"

func addNotifiedToScheduledOperationsTable(db *sql.Tx) error {
	_, err := db.Exec(`
		ALTER TABLE scheduled_operations ADD COLUMN notified BOOLEAN DEFAULT FALSE;
	`)
	return err
}
