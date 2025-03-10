package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// PostgreSQL is a struct that contains a connection to PostgreSQL.
type PostgreSQL struct {
	DB *sqlx.DB
}

// PostgreSQLOptions is a struct that contains options for connecting to PostgreSQL.
type PostgreSQLOptions struct {
	User     string
	Password string
	Database string
	Host     string
	Port     string
	SSLMode  string
}

func (p PostgreSQLOptions) convertToConnectionURL() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s&hostaddr=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode, p.Host,
	)
}

// NewPostgreSQL returns a new instance of PostgreSQL.
func NewPostgreSQL(options PostgreSQLOptions) (*PostgreSQL, error) {
	db, err := sqlx.Open("postgres", options.convertToConnectionURL())
	if err != nil {
		return nil, fmt.Errorf("open postgresql connection: %w", err)
	}

	return &PostgreSQL{
		DB: db,
	}, nil
}

// Ping checks if the PostgreSQL connection is alive.
func (p *PostgreSQL) Ping() error {
	return p.DB.Ping()
}

// Close closes the PostgreSQL connection.
func (p *PostgreSQL) Close() error {
	return p.DB.Close()
}
