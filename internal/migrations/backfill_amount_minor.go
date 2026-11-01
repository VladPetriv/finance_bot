package migrations

import (
	"database/sql"
	"fmt"
)

var amountMinorTables = [...]string{"balances", "operations", "balance_subscriptions"}

func backfillAmountMinor(tx *sql.Tx) error {
	for _, table := range amountMinorTables {
		var rounded int64

		err := tx.QueryRow(fmt.Sprintf(
			`SELECT COUNT(*) FROM %[1]s
			WHERE amount_minor IS NULL AND amount IS NOT NULL AND amount <> ''
			AND amount::numeric * 100 <> ROUND(amount::numeric * 100);`,
			table,
		)).Scan(&rounded)
		if err != nil {
			return fmt.Errorf("count rows needing rounding in %s: %w", table, err)
		}

		_, err = tx.Exec(fmt.Sprintf(
			`UPDATE %[1]s SET amount_minor = ROUND(amount::numeric * 100)::bigint
			WHERE amount_minor IS NULL AND amount IS NOT NULL AND amount <> '';`,
			table,
		))
		if err != nil {
			return fmt.Errorf("backfill amount_minor in %s: %w", table, err)
		}

		_, err = tx.Exec(fmt.Sprintf(
			`DO $$ BEGIN RAISE NOTICE 'backfill amount_minor: %d rows rounded in %s'; END $$;`,
			rounded, table,
		))
		if err != nil {
			return fmt.Errorf("log rounded rows in %s: %w", table, err)
		}
	}

	return nil
}
