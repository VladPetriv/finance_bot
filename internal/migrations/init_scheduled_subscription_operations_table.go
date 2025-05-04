package migrations

import "database/sql"

func initScheduledOperationTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE scheduled_operations (
			id VARCHAR(255) PRIMARY KEY,
			subscription_id VARCHAR(255) NOT NULL,
			creation_date TIMESTAMP NOT NULL
		);

		ALTER TABLE scheduled_operations
	        ADD CONSTRAINT fk_scheduled_operations_subscription_id FOREIGN KEY (subscription_id) REFERENCES balance_subscriptions(id) ON DELETE CASCADE;
	`)

	return err
}
