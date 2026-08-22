package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func Connect(ctx context.Context, cfg Config) (*Client, error) {
	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxIdleTime).
		SetConnectTimeout(cfg.ConnectTimeout)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return &Client{
		Client:   client,
		Database: client.Database(cfg.Database),
	}, nil
}

func (c *Client) Close(ctx context.Context) error {
	return c.Client.Disconnect(ctx)
}
