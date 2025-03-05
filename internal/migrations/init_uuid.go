package migrations

import "database/sql"

func initUUIDExtension(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	`)

	return err
}
