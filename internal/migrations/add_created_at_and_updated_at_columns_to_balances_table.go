package migrations

import (
	"database/sql"
	"fmt"
)

func addCreatedAtAndUpdatedAtColumnsToBalancesTable(db *sql.Tx) error {
	_, err := db.Exec(`
		ALTER TABLE balances ADD COLUMN created_at TIMESTAMP DEFAULT NOW();
	`)
	if err != nil {
		return fmt.Errorf("add created_at column to balances table: %w", err)
	}

	_, err = db.Exec(`
			ALTER TABLE balances ADD COLUMN updated_at TIMESTAMP DEFAULT NOW();
		`)
	if err != nil {
		return fmt.Errorf("add updated_at column to balances table: %w", err)
	}

	return nil
}
