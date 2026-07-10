package bench

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisDriver handles Redis, Valkey, and DragonflyDB (all Redis-protocol compatible).
type RedisDriver struct {
	variant string // "redis", "valkey", or "dragonfly"
	client  *redis.Client
	cfg     *Config
}

func (d *RedisDriver) Name() string { return d.variant }

func (d *RedisDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	var addr string
	switch d.variant {
	case "redis":
		addr = cfg.RedisAddr
	case "valkey":
		addr = cfg.ValkeyAddr
	case "dragonfly":
		addr = cfg.DragonflyAddr
	default:
		return fmt.Errorf("unknown redis variant: %s", d.variant)
	}

	d.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		PoolSize: cfg.Concurrency + 10,
	})

	return d.client.Ping(context.Background()).Err()
}

func (d *RedisDriver) Cleanup() error {
	return d.client.FlushDB(context.Background()).Err()
}

func (d *RedisDriver) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

func (d *RedisDriver) Write(key string, value []byte) error {
	return d.client.Set(context.Background(), key, value, 0).Err()
}

func (d *RedisDriver) Read(key string) ([]byte, error) {
	val, err := d.client.Get(context.Background(), key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}
