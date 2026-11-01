package migrations

import (
	"database/sql"
	"fmt"
)

func addAmountMinorToBalancesTable(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE balances ADD COLUMN amount_minor BIGINT;`)
	if err != nil {
		return fmt.Errorf("add amount_minor column: %w", err)
	}

	_, err = tx.Exec(`ALTER TABLE balances ALTER COLUMN amount DROP NOT NULL;`)
	if err != nil {
		return fmt.Errorf("drop not null on amount column: %w", err)
	}

	return nil
}
