package migrations

import (
	"database/sql"
)

func initAutomaticReport(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TYPE automatic_report_period AS ENUM ('daily', 'weekly', 'monthly', 'quarterly', 'yearly');

		CREATE TABLE automatic_reports (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			period automatic_report_period NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE TABLE automatic_report_balances (
			automatic_report_id VARCHAR(255) NOT NULL REFERENCES automatic_reports(id) ON DELETE CASCADE,
			balance_id VARCHAR(255) NOT NULL REFERENCES balances(id) ON DELETE CASCADE,
			PRIMARY KEY (automatic_report_id, balance_id)
		);

		CREATE TABLE automatic_report_categories (
			automatic_report_id VARCHAR(255) NOT NULL REFERENCES automatic_reports(id) ON DELETE CASCADE,
			category_id VARCHAR(255) NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			PRIMARY KEY (automatic_report_id, category_id)
		);

		CREATE TABLE scheduled_report_executions (
			id VARCHAR(255) PRIMARY KEY,
			automatic_report_id VARCHAR(255) NOT NULL REFERENCES automatic_reports(id) ON DELETE CASCADE,
			execution_date TIMESTAMP NOT NULL
		);
	`)

	return err
}
