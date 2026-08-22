package mongo

import "time"

type Config struct {
	URI            string
	Database       string
	ConnectTimeout time.Duration
	MaxPoolSize    uint64
	MinPoolSize    uint64
	MaxIdleTime    time.Duration
}

func DefaultConfig(uri, database string) Config {
	return Config{
		URI:            uri,
		Database:       database,
		ConnectTimeout: 10 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    0,
		MaxIdleTime:    60 * time.Second,
	}
}
