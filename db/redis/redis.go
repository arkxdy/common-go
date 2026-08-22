package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	Client *redis.Client
}

func Connect(ctx context.Context, cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Client{Client: rdb}, nil
}

func (c *Client) Close() error {
	return c.Client.Close()
}
