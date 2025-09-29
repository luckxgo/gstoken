package storage

import (
	"context"
	"encoding/json"
	"github.com/luckxgo/gstoken/core"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage Redis存储实现
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage 创建Redis存储
func NewRedisStorage(config core.RedisConfig) *RedisStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	return &RedisStorage{
		client: rdb,
	}
}

// Set 设置键值
func (r *RedisStorage) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expire).Err()
}

// Get 获取值
func (r *RedisStorage) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal([]byte(data), &result)
	return result, err
}

// Delete 删除键
func (r *RedisStorage) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func (r *RedisStorage) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// Keys 获取匹配的键列表
func (r *RedisStorage) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Close 关闭连接
func (r *RedisStorage) Close() error {
	return r.client.Close()
}
