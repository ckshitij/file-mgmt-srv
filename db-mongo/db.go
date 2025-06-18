package dbmongo

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBClient struct {
	Client *mongo.Client
}

func NewMongoDBClient(ctx context.Context, URI string) (*MongoDBClient, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URI))
	if err != nil {
		return nil, err
	}
	return &MongoDBClient{Client: client}, nil
}

func (db *MongoDBClient) Disconnect(ctx context.Context) error {
	if err := db.Client.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
	}
	log.Println("Disconnected from MongoDB successfully")
	return nil
}

func (db *MongoDBClient) Ping(ctx context.Context) error {
	if err := db.Client.Ping(ctx, nil); err != nil {
		log.Printf("Error pinging MongoDB: %v", err)
		return err
	}
	log.Println("Ping to MongoDB successful")
	return nil
}

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
