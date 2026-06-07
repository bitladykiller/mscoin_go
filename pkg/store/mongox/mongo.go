// Package mongox centralizes MongoDB initialization and lifecycle management.
package mongox

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config describes the MongoDB connection used by services that need timeseries
// or weakly structured storage.
type Config struct {
	URI      string
	Username string
	Password string
	Database string
}

// Client keeps both the raw client and the selected default database.
type Client struct {
	raw *mongo.Client
	db  *mongo.Database
}

// New creates a Mongo client and validates the connection immediately so
// services fail fast during startup instead of failing on the first request.
func New(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)
	if cfg.Username != "" || cfg.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	raw, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	if err := raw.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return &Client{
		raw: raw,
		db:  raw.Database(cfg.Database),
	}, nil
}

// Database returns the default database configured for the service.
func (c *Client) Database() *mongo.Database {
	return c.db
}

// Disconnect closes the Mongo client gracefully.
func (c *Client) Disconnect(ctx context.Context) error {
	return c.raw.Disconnect(ctx)
}
