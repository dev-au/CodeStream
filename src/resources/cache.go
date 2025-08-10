package resources

import (
	"CodeStream/src"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

type Cache struct {
	Ctx    context.Context
	Client *redis.Client
}

func NewCacheContext() *Cache {
	return &Cache{
		Ctx:    context.Background(),
		Client: RedisClient,
	}
}

func SetupRedis() {

	addr, err := redis.ParseURL(src.Config.RedisUrl)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: addr.Addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("failed to connect to Redis: %s", err))
		return
	}
	RedisClient = rdb
}

func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {

	err := c.Client.Set(c.Ctx, key, value, expiration).Err()
	if err != nil {
		panic(fmt.Sprintf("failed to set key: %s", err))
	}
}
func (c *Cache) Get(key string) interface{} {
	val, err := c.Client.Get(c.Ctx, key).Result()
	if err != nil || val == "" {
		return nil
	}
	return val
}

func (c *Cache) Delete(key string) {
	c.Client.Del(c.Ctx, key)
	return
}

func (c *Cache) Exists(key string) bool {
	count, _ := c.Client.Exists(c.Ctx, key).Result()
	return count > 0
}

func (c *Cache) SetExpiration(key string, expiration time.Duration) {
	c.Client.Expire(c.Ctx, key, expiration)
}

func (c *Cache) GetTTL(key string) time.Duration {
	r, err := c.Client.TTL(c.Ctx, key).Result()
	if err != nil {
		panic(err)
		return 0
	}
	return r
}
