package runtime

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type KV interface {
	// TODO(zhuzilin) Currently, the go-redis library will return string by default.
	// Find a way to return correct type of the val.
	Put(key string, val string) error
	Get(key string) (string, error)
}

type RedisKV struct {
	rdb *redis.Client
}

func (kv *RedisKV) Put(key string, val string) error {
	return kv.rdb.Set(context.Background(), key, val, 0).Err()
}

func (kv *RedisKV) Get(key string) (string, error) {
	return kv.rdb.Get(context.Background(), key).Result()
}

func NewRedisKV(url, password string, db int) KV {
	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       0,
	})

	return &RedisKV{
		rdb: rdb,
	}
}
