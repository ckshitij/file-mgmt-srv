// Package dbmongo provides a wrapper for managing MongoDB connections,
// including connection lifecycle, health checks, and access to databases.
package dbmongo

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBClient encapsulates a MongoDB client connection.
type MongoDBClient struct {
	Client *mongo.Client // The underlying MongoDB driver client
}

// NewMongoDBClient initializes a new MongoDB client using the provided URI.
// It establishes a connection and returns a wrapped client instance.
func NewMongoDBClient(ctx context.Context, URI string) (*MongoDBClient, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URI))
	if err != nil {
		return nil, err
	}
	return &MongoDBClient{Client: client}, nil
}

// Disconnect cleanly shuts down the MongoDB client connection.
// It logs both success and error messages.
func (db *MongoDBClient) Disconnect(ctx context.Context) error {
	if err := db.Client.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
		return err
	}
	log.Println("Disconnected from MongoDB successfully")
	return nil
}

// Ping verifies that the MongoDB server is reachable.
// Returns an error if the connection is lost or not healthy.
func (db *MongoDBClient) Ping(ctx context.Context) error {
	if err := db.Client.Ping(ctx, nil); err != nil {
		log.Printf("Error pinging MongoDB: %v", err)
		return err
	}
	log.Println("Ping to MongoDB successful")
	return nil
}

// GetDatabase returns a reference to the named database from the client.
// If the client is uninitialized, it logs a warning and returns nil.
func (db *MongoDBClient) GetDatabase(dbName string) *mongo.Database {
	if db.Client == nil {
		log.Println("MongoDB client is not initialized")
		return nil
	}
	database := db.Client.Database(dbName)
	if database == nil {
		log.Printf("Database %s does not exist", dbName)
		return nil
	}
	log.Printf("Connected to database: %s", dbName)
	return database
}
