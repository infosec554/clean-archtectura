package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/infosec554/clean-archtectura/config"
)

var (
	once   sync.Once
	client *redis.Client
)

type ICache interface {
	Set(key string, value any, duration time.Duration) error
	Get(key string) (string, error)
	Scan(key string, val any) error
}

type cache struct {
	client *redis.Client
}

func (c cache) Set(key string, value any, duration time.Duration) error {
	res := c.client.Set(
		context.Background(),
		key,
		value,
		duration,
	)

	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (c cache) Get(key string) (string, error) {
	res, err := c.client.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return res, nil
}

func (c cache) Scan(key string, val any) error {
	err := c.client.Get(context.Background(), key).Scan(val)
	if err != nil {
		return err
	}
	return nil
}

func NewCache(cfg config.Config) ICache {
	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		res := client.Ping(context.Background())
		if res.Err() != nil {
			panic(res.Err())
		}
	})

	return &cache{client: client}
}
