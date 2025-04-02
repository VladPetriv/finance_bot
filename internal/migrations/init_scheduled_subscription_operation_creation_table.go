package migrations

import "database/sql"

func initScheduledOperationCreationTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE scheduled_operation_creations (
			id VARCHAR(255) PRIMARY KEY,
			subscription_id VARCHAR(255) NOT NULL,
			creation_date TIMESTAMP NOT NULL
		);

		ALTER TABLE scheduled_operation_creations
	        ADD CONSTRAINT fk_scheduled_operation_creations_subscription_id FOREIGN KEY (subscription_id) REFERENCES subscription(id) ON DELETE CASCADE;
	`)

	return err
}
