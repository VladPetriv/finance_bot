package migrations

import "database/sql"

func addNotifyAboutSubscriptionPaymentsToUserSettings(db *sql.Tx) error {
	_, err := db.Exec(`
		ALTER TABLE user_settings ADD COLUMN notify_about_subscription_payments BOOLEAN DEFAULT FALSE;
	`)
	return err
}
