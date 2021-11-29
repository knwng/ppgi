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

type redisKV struct {
	rdb *redis.Client
}

func (kv *redisKV) Put(key string, val string) error {
	return kv.rdb.Set(context.Background(), key, val, 0).Err()
}

func (kv *redisKV) Get(key string) (string, error) {
	return kv.rdb.Get(context.Background(), key).Result()
}

func NewKV(URL string) KV {
	rdb := redis.NewClient(&redis.Options{
		Addr:     URL,
		Password: "",
		DB:       0,
	})

	return &redisKV{
		rdb: rdb,
	}
}
