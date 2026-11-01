package migrations

import (
	"database/sql"
	"fmt"
)

func dropLegacyAmountColumns(tx *sql.Tx) error {
	for _, table := range amountMinorTables {
		_, err := tx.Exec(fmt.Sprintf(`ALTER TABLE %s DROP COLUMN amount;`, table))
		if err != nil {
			return fmt.Errorf("drop legacy amount column in %s: %w", table, err)
		}
	}

	return nil
}
