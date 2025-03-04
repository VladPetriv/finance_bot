package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type postgreSQL struct {
	db *sqlx.DB
}

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

func NewPostgreSQL(options PostgreSQLOptions) (*postgreSQL, error) {
	db, err := sqlx.Open("postgres", options.convertToConnectionURL())
	if err != nil {
		return nil, fmt.Errorf("open postgresql connection: %w", err)

	}

	return &postgreSQL{
		db: db,
	}, nil
}

func (p *postgreSQL) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *postgreSQL) Close() error {
	return p.db.Close()
}
