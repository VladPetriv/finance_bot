package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// PostgreSQL is a struct that contains a connection to PostgreSQL.
type PostgreSQL struct {
	db *sqlx.DB
}

// PostgreSQLOptions is a struct that contains options for connecting to PostgreSQL.
type PostgreSQLOptions struct {
	User     string
	Password string
	Database string
	Host     string
	SSLMode  string
}

func (p PostgreSQLOptions) convertToConnectionURL() string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s sslmode=%s",
		p.User, p.Password, p.Database, p.Host, p.SSLMode,
	)
}

// NewPostgreSQL returns a new instance of PostgreSQL.
func NewPostgreSQL(options PostgreSQLOptions) (*PostgreSQL, error) {
	db, err := sqlx.Open("postgres", options.convertToConnectionURL())
	if err != nil {
		return nil, fmt.Errorf("open postgresql connection: %w", err)

	}

	return &PostgreSQL{
		db: db,
	}, nil
}

func (p *PostgreSQL) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgreSQL) Close() error {
	return p.db.Close()
}
