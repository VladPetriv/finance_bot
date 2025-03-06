package migrations

import "database/sql"

func initCategoryTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE categories (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			title VARCHAR(255) NOT NULL
		);


		ALTER TABLE ONLY categories
    		ADD CONSTRAINT categories_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
	`)

	return err
}
