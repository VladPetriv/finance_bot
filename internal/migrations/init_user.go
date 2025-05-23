package migrations

import "database/sql"

func initUserTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE users (
		    id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) NOT NULL
		);


		ALTER TABLE ONLY users
    		ADD CONSTRAINT users_username_fkey UNIQUE (username);

      	CREATE INDEX idx_users_username ON users (username);
	`)

	return err
}
