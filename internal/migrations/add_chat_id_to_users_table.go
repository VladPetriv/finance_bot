package migrations

import "database/sql"

func addChatIDToUsersTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE users ADD COLUMN chat_id BIGINT DEFAULT 0;
	`)
	return err
}
