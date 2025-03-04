package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB represents a connection with MongoDB.
type MongoDB struct {
	DB *mongo.Database
}

var _ Database = (*MongoDB)(nil)

// NewMongoDB return new instance of MongoDB.
func NewMongoDB(ctx context.Context, uri, dbName string) (*MongoDB, error) {
	var database MongoDB

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect to mongodb: %w", err)
	}

	database.DB = client.Database(dbName)

	return &database, nil
}

// Close closes the connection with MongoDB.
func (m MongoDB) Close() error {
	if m.DB != nil {
		err := m.DB.Client().Disconnect(context.Background())
		if err != nil {
			return fmt.Errorf("can't discount: %w", err)
		}
	}

	return nil
}

// Ping pings the database to check connection with MongoDB.
func (m MongoDB) Ping(ctx context.Context) error {
	return m.DB.Client().Ping(ctx, nil)
}
