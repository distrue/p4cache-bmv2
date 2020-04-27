package util

import (
	"errors"

	"github.com/go-redis/redis/v7"
)

type RedisClient interface {
	GetItem(string) (string, error)
	SetItem(string, interface{}) error
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient() RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &redisClient{client}
}

func (c *redisClient) GetItem(key string) (string, error) {
	val, err := c.client.Get(key).Result()
	if err == redis.Nil {
		return "", errors.New("No item found")
	} else if err != nil {
		panic(err)
	}
	return val, nil
}

func (c *redisClient) SetItem(key string, val interface{}) error {
	return c.client.Set(key, val, 0).Err()
}
