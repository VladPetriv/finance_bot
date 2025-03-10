package migrations

import "database/sql"

func initStateTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE states (
		    id VARCHAR(255) PRIMARY KEY,
			user_username VARCHAR(255) NULL,
			flow VARCHAR(255) NOT NULL,
			steps JSONB NULL,
			metadata JSONB NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`)

	return err
}
