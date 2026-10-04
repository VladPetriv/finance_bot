package migrations

import "database/sql"

func addTimezoneToUserSettingsTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE user_settings ADD COLUMN timezone VARCHAR(64) NOT NULL DEFAULT 'UTC';
	`)
	return err
}
