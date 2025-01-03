package database

import "context"

// Database represents a database connection.
type Database interface {
	// Ping pings the database.
	Ping(ctx context.Context) error
	// Close closes the connection with database.
	Close() error
}
