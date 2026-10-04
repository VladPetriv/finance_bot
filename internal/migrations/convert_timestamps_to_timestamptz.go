package migrations

import "database/sql"

func convertTimestampsToTimestamptz(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE operations
			ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
			ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

		ALTER TABLE user_settings
			ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
			ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

		ALTER TABLE balances
			ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
			ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

		ALTER TABLE balance_subscriptions
			ALTER COLUMN start_at TYPE TIMESTAMPTZ USING start_at AT TIME ZONE 'UTC',
			ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
			ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

		ALTER TABLE automatic_reports
			ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
			ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

		ALTER TABLE states
			ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
			ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

		ALTER TABLE scheduled_report_executions
			ALTER COLUMN execution_date TYPE TIMESTAMPTZ USING execution_date AT TIME ZONE 'UTC';

		ALTER TABLE scheduled_operations
			ALTER COLUMN creation_date TYPE TIMESTAMPTZ USING creation_date AT TIME ZONE 'UTC';
	`)
	return err
}
