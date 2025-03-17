package migrations

import "database/sql"

func initUserSettingsTable(db *sql.DB) error {
	query := `
		CREATE TABLE user_settings (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			ai_parser_enabled BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		ALTER TABLE ONLY user_settings
    		ADD CONSTRAINT user_settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
	`
	_, err := db.Exec(query)
	return err
}
