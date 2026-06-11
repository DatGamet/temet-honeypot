package database

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Connection struct {
	client   *mongo.Client
	database *mongo.Database
}

func (c *Connection) Client() *mongo.Client {
	return c.client
}

func (c *Connection) Database() *mongo.Database {
	return c.database
}

func Connect(ctx context.Context) (*Connection, error) {
	client, err := mongo.Connect(
		options.Client().
			ApplyURI(os.Getenv("MONGO_URI")).
			SetConnectTimeout(5 * time.Second).
			SetServerSelectionTimeout(3 * time.Second).
			SetRetryWrites(true).
			SetMaxPoolSize(200).
			SetMinPoolSize(10).
			SetMaxConnIdleTime(2 * time.Minute),
	)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return &Connection{
		client:   client,
		database: client.Database("account-service"),
	}, nil
}
