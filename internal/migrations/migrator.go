package migrations

import (
	"fmt"

	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/lopezator/migrator"
)

// MigrateDB applies migrations to the database
func MigrateDB(log *logger.Logger, db *sqlx.DB, dbName string, migrations []any) error {
	logger := log.With().Str("name", "MigrateDB").Logger()
	logger.Debug().Str("dbName", dbName).Msg("migrating database ...")

	migrator, err := migrator.New(migrator.Migrations(migrations...))
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	databaseVersion := len(migrations)

	pending, err := migrator.Pending(db.DB)
	switch err != nil {
	case true:
		logger.Error().Err(err).Msg("got pending error")
		databaseVersion = 0
	case false:
		databaseVersion -= len(pending)
	}

	logger.Info().Int("dbVersion", databaseVersion).Msg("current database version")

	if len(pending) > 0 || databaseVersion == 0 {
		logger.Info().Msg("new migrations were found, running migrations ...")

		err := migrator.Migrate(db.DB)
		if err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}

		logger.Info().Int("updatedDatabaseVersion", len(migrations)).Msg("migrations were successfully completed")
		return nil
	}

	logger.Info().Msg("no new migrations were found")
	return nil
}
