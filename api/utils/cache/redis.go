package cache

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RedisCache struct {
	client   *redis.Client
	duration time.Duration
}

func NewRedisCache(duration time.Duration) *RedisCache {
	addr := os.Getenv("REDIS_SERVER_URL")
	password := os.Getenv("REDIS_PASSWORD")
	db := os.Getenv("REDIS_DB")
	if addr == "" {
		logrus.Error("required redis server url")
		return nil
	}
	if password == "" {
		logrus.Error("required redis password")
		return nil
	}
	d, err := strconv.Atoi(db)
	if err != nil {
		logrus.Error("invalid db numbers")
		return nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       d,
	})
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		logrus.Error("error in ping redis :: ", err)
		return nil
	}
	return &RedisCache{
		client:   client,
		duration: duration,
	}
}

func (r *RedisCache) Set(key string, value interface{}, duration ...time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		logrus.Error("redis marshal error :: ", err)
		return
	}
	if len(duration) == 0 {
		duration = append(duration, r.duration)
	}
	err = r.client.Set(context.Background(), key, data, duration[0]).Err()
	if err != nil {
		logrus.Error("redis set error :: ", err)
	}
}

func (r *RedisCache) Get(key string, result interface{}) (ok bool) {
	valueBytes, err := r.client.Get(context.Background(), key).Bytes()
	if err != nil {
		return
	}
	err = json.Unmarshal(valueBytes, &result)
	if err != nil {
		return
	}
	ok = true
	return
}

func (r *RedisCache) Delete(key string) {
	_ = r.client.Del(context.Background(), key)
}

func (r *RedisCache) DeleteMulti(keys []string) {
	for _, key := range keys {
		r.client.Del(context.Background(), key)
	}
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}
