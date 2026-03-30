package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client.
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Redis cache client.
func NewClient(addr, password string, db int) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Client{rdb: rdb}
}

// Ping checks connectivity.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Set stores a key/value pair with expiry.
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiry).Err()
}

// Get retrieves a value by key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

// Del deletes one or more keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// SetNX sets a key only if it does not exist (used for idempotency locks).
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiry time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, key, value, expiry).Result()
}

// Incr increments an integer counter.
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// Expire sets a TTL on an existing key.
func (c *Client) Expire(ctx context.Context, key string, expiry time.Duration) error {
	return c.rdb.Expire(ctx, key, expiry).Err()
}

// TTL returns the remaining time-to-live of a key.
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

// Exists checks whether a key exists.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.rdb.Exists(ctx, key).Result()
	return n > 0, err
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// RDB exposes the underlying redis.Client for advanced use.
func (c *Client) RDB() *redis.Client {
	return c.rdb
}
