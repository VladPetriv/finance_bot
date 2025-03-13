package migrations

import "database/sql"

func initBalanceSubscriptionTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TYPE subscription_period AS ENUM ('weekly', 'monthly', 'yearly');

		CREATE TABLE balance_subscriptions (
			id VARCHAR(255) PRIMARY KEY,
			balance_id VARCHAR(255) NOT NULL,
			category_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			period subscription_period,
			start_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);


		ALTER TABLE balance_subscriptions ADD CONSTRAINT fk_balance_subscriptions_balance_id FOREIGN KEY (balance_id) REFERENCES balances(id);
		ALTER TABLE balance_subscriptions ADD CONSTRAINT fk_balance_subscriptions_category_id FOREIGN KEY (category_id) REFERENCES categories(id);
		`)

	return err
}
