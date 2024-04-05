package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	client *mongo.Client
}

func NewDB(connectionString string) (*DB, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, err
	}
	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return &DB{client}, nil
}

func (db *DB) Close() error {
	return db.client.Disconnect(context.Background())
}

func (db *DB) Client() *mongo.Client {
	return db.client
}
