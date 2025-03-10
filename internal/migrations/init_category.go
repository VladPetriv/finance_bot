package migrations

import "database/sql"

func initCategoryTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE categories (
		    id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NULL,
			title VARCHAR(255) NOT NULL
		);


		ALTER TABLE ONLY categories
    		ADD CONSTRAINT categories_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
	`)

	return err
}
