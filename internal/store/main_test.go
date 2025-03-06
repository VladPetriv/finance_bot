package store_test

import (
	"fmt"
	"testing"

	"github.com/VladPetriv/finance_bot/config"
	"github.com/VladPetriv/finance_bot/internal/migrations"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

var (
	postgresDB   *database.PostgreSQL
	log          *logger.Logger
	cfg          *config.Config
	postgresPort string
)

func TestMain(m *testing.M) {
	cfg = config.Get()
	log = logger.New(logger.LoggergerOptions{
		LogLevel:        "debug",
		PrettyLogOutput: true,
	})

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatal().Msgf("could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatal().Msgf("could not connect to Docker: %s", err)
	}

	resource, err := pool.Run(
		"postgres",
		"latest",
		[]string{
			"POSTGRES_USER=" + cfg.PostgreSQL.User,
			"POSTGRES_PASSWORD=" + cfg.PostgreSQL.Password,
			"POSTGRES_DB=" + cfg.PostgreSQL.Database,
		})
	if err != nil {
		log.Fatal().Msgf("could not start resource: %s", err)
	}

	postgresPort = resource.GetPort("5432/tcp")
	retryErr := pool.Retry(func() error {
		var err error
		postgresDB, err = database.NewPostgreSQL(database.PostgreSQLOptions{
			User:     cfg.PostgreSQL.User,
			Password: cfg.PostgreSQL.Password,
			Database: cfg.PostgreSQL.Database,
			Host:     "localhost",
			Port:     postgresPort,
			SSLMode:  "disable",
		})
		if err != nil {
			return err
		}

		return postgresDB.Ping()
	})
	if retryErr != nil {
		log.Error().Msgf("could not connect to database: %s", err)
	}

	defer func() {
		if postgresDB != nil {
			err := postgresDB.Close()
			if err != nil {
				log.Error().Msgf("could not close database: %s", err)
			}
		}

		if pool != nil {
			err = pool.Purge(resource)
			if err != nil {
				log.Fatal().Msgf("could not purge resource: %s", err)
			}
		}
	}()

	m.Run()
}

func createTestDB(t *testing.T, testCaseName string) *database.PostgreSQL {
	t.Helper()

	_, err := postgresDB.DB.Exec(fmt.Sprintf("CREATE DATABASE %s;", testCaseName))
	require.NoError(t, err)

	testDB, err := database.NewPostgreSQL(database.PostgreSQLOptions{
		User:     cfg.PostgreSQL.User,
		Password: cfg.PostgreSQL.Password,
		Database: testCaseName,
		Host:     "localhost",
		Port:     postgresPort,
		SSLMode:  "disable",
	})
	require.NoError(t, err)

	err = migrations.MigrateDB(log, testDB.DB, testCaseName, migrations.Migrations)
	require.NoError(t, err)

	t.Cleanup(func() {
		testDB.Close()

		// Drop the test database
		_, err = postgresDB.DB.Exec(fmt.Sprintf("DROP DATABASE %s;", testCaseName))
		require.NoError(t, err)
	})

	return testDB
}
