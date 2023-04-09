package database

// Database represents a database connection.
type Database interface {
	// Close closes the connection with database.
	Close() error
}
