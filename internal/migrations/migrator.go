package migrations

import (
	"fmt"

	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/lopezator/migrator"
)

// MigrateDBOptions represents options for the MigrateDB function.
type MigrateDBOptions struct {
	Logger     *logger.Logger
	DB         *sqlx.DB
	DBName     string
	Migrations []any
}

// MigrateDB applies migrations to the database
func MigrateDB(opts MigrateDBOptions) error {
	logger := logger.NewDummy().With().Logger()
	if opts.Logger != nil {
		logger = opts.Logger.With().Str("name", "MigrateDB").Logger()
		logger.Debug().Str("dbName", opts.DBName).Msg("migrating database ...")
	}

	migratorOpts := []migrator.Option{
		migrator.Migrations(opts.Migrations...),
	}
	if opts.Logger == nil {
		migratorOpts = append(migratorOpts, migrator.WithLogger(&dummyMigratorLogger{}))
	}

	migrator, err := migrator.New(migratorOpts...)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	databaseVersion := len(opts.Migrations)
	pending, err := migrator.Pending(opts.DB.DB)
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

		err := migrator.Migrate(opts.DB.DB)
		if err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}

		logger.Info().Int("updatedDatabaseVersion", len(opts.Migrations)).Msg("migrations were successfully completed")
		return nil
	}

	logger.Info().Msg("no new migrations were found")
	return nil
}

type dummyMigratorLogger struct{}

func (d *dummyMigratorLogger) Printf(text string, args ...any) {}
