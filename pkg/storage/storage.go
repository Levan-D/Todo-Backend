package storage

import (
	"fmt"
	"github.com/Levan-D/Todo-Backend/pkg/config"
	"github.com/go-redis/redis/v7"
	"time"
)

var client *redis.Client

func Initialize() {
	dsn := fmt.Sprintf("%s:%d", config.Get().Redis.Host, config.Get().Redis.Port)

	client = redis.NewClient(&redis.Options{
		Addr:     dsn,
		Password: config.Get().Redis.Password,
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
}

func Get(key string) (string, error) {
	return client.Get(key).Result()
}

func FindByPattern(pattern string) ([]string, error) {
	find, err := client.Keys(pattern).Result()
	if err != nil {
		return nil, err
	}

	return find, nil
}

func Set(key string, value interface{}, expiration time.Duration) error {
	return client.Set(key, value, expiration).Err()
}

func Delete(key string) (int64, error) {
	return client.Del(key).Result()
}
