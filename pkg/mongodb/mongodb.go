package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	DB *mongo.Database
}

// New return new instance of MongoDB.
func New(uri, dbName string) (*mongoDB, error) {
	var database *mongoDB

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect to mongodb: %w", err)
	}

	database.DB = client.Database(dbName)

	return database, nil
}
