package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb *redis.Client
}

func New(addr string) *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // нет пароля
		DB:       0,  // база по умолчанию
	})
	return &Cache{rdb: rdb}
}

func (c *Cache) Set(key string, value string, ttl time.Duration) error {
	ctx := context.Background()
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Get(key string) (string, error) {
	ctx := context.Background()
	return c.rdb.Get(ctx, key).Result()
}

func (c *Cache) TTL(key string) (time.Duration, error) {
	ctx := context.Background()
	return c.rdb.TTL(ctx, key).Result()
}

func (c *Cache) Close() error {
	return c.rdb.Close()
}
