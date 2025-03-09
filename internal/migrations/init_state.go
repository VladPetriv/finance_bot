package migrations

import "database/sql"

func initStateTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE states (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_username VARCHAR(255) NOT NULL,
			flow VARCHAR(255) NOT NULL,
			steps JSONB NULL,
			metadata JSONB NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);


		ALTER TABLE ONLY states
    		ADD CONSTRAINT states_user_username_fkey FOREIGN KEY (user_username) REFERENCES users(username);
	`)

	return err
}
