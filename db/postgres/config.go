package postgres

import "time"

type Config struct {
	DSN             string // e.g. "postgres://user:pass@localhost:5432/db?sslmode=disable"
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func DefaultConfig(dsn string) Config {
	return Config{
		DSN:             dsn,
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}
}
