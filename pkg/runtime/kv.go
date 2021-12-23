package runtime

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func GetExistingStringAndIndex(raw []interface{}) ([]string, []int) {
	str := make([]string, 0)
	index := make([]int, 0)
	for i, data := range raw {
		switch data.(type) {
		case string:
			str = append(str, data.(string))
			index = append(index, i)
		}
	}
	return str, index
}

type KV interface {
	// TODO(zhuzilin) Currently, the go-redis library will return string by default.
	// Find a way to return correct type of the val.
	Put(key string, val string) error
	Get(key string) (string, error)
	Del(key string) error

	HashPut(key string, data map[string]string) error
	HashGet(key, field string) (string, error)
	HashMultiGet(key string, fields []string) ([]interface{}, error)
	HashDel(key string, fields []string) error

	SetAdd(key string, members []string) error
	SetCheck(key string, members []string) ([]bool, error)
	SetDel(key string, members []string) error
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

func (kv *RedisKV) Del(key string) error {
	return kv.rdb.Del(context.Background(), key).Err()
}

func (kv *RedisKV) HashDel(key string, fields []string) error {
	return kv.rdb.HDel(context.Background(), key, fields...).Err()
}

func (kv *RedisKV) HashPut(key string, data map[string]string) error {
	return kv.rdb.HSet(context.Background(), key, data).Err()
}

func (kv *RedisKV) HashGet(key, field string) (string, error) {
	return kv.rdb.HGet(context.Background(), key, field).Result()
}

func (kv *RedisKV) HashMultiGet(key string, fields []string) ([]interface{}, error) {
	return kv.rdb.HMGet(context.Background(), key, fields...).Result()
}

func (kv *RedisKV) SetAdd(key string, members []string) error {
	return kv.rdb.SAdd(context.Background(), key, members).Err()
}

func (kv *RedisKV) SetCheck(key string, members []string) ([]bool, error) {
	return kv.rdb.SMIsMember(context.Background(), key, members).Result()
}

func (kv *RedisKV) SetDel(key string, members []string) error {
	return kv.rdb.SRem(context.Background(), key, members).Err()
}

func NewRedisKV(url, password string, db int) *RedisKV {
	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       db,
	})

	return &RedisKV{
		rdb: rdb,
	}
}
