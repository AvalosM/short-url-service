package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	namespace = "short_url:%s"
)

// Cache contains resource to interact with cache
type Cache struct {
	client *redis.Client
}

// NewCache creates a new Cache instance with the provided configuration
func NewCache(config *Config) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
		Protocol: config.Protocol,
	})

	client.Ping(context.Background())

	return &Cache{
		client: client,
	}
}

// Healthy checks cache connection health
func (c *Cache) Healthy() bool {
	_, err := c.client.Ping(context.Background()).Result()

	return err == nil
}

// Close closes the cache connection
func (c *Cache) Close() error {
	return c.client.Close()
}

// Get retrieves a value from the cache by its key
func (c *Cache) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := c.client.Get(ctx, fmt.Sprintf(namespace, key)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}

	return val, true, nil
}

// Set adds a key-value pair to the cache with a ttl expiration time
func (c *Cache) Set(ctx context.Context, key string, value string, duration time.Duration) error {
	_, err := c.client.Set(ctx, fmt.Sprintf(namespace, key), value, duration).Result()
	if err != nil {
		return err
	}

	return nil
}

// Delete removes a key from the cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	_, err := c.client.Del(ctx, fmt.Sprintf(namespace, key)).Result()
	if err != nil {
		return err
	}

	return nil
}
