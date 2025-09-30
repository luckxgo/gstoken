package storage

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"time"

	"github.com/luckxgo/gstoken/core"

	"github.com/redis/go-redis/v9"
)

// RedisStorage Redis存储实现
type RedisStorage struct {
	client redis.UniversalClient
}

// NewRedisStorage 创建Redis存储
func NewRedisStorage(config core.RedisConfig) *RedisStorage {
	var rdb redis.UniversalClient
	if config.ClusterEnabled && len(config.ClusterAddrs) > 0 {
		opts := &redis.ClusterOptions{
			Addrs:           config.ClusterAddrs,
			Username:        config.Username,
			Password:        config.Password,
			MaxRetries:      config.MaxRetries,
			MinRetryBackoff: config.MinRetryBackoff,
			MaxRetryBackoff: config.MaxRetryBackoff,
			DialTimeout:     config.DialTimeout,
			ReadTimeout:     config.ReadTimeout,
			WriteTimeout:    config.WriteTimeout,
			PoolSize:        config.PoolSize,
			MinIdleConns:    config.MinIdleConns,
			PoolTimeout:     config.PoolTimeout,
			ConnMaxIdleTime: config.ConnMaxIdleTime,
		}
		if config.ClientName != "" {
			opts.ClientName = config.ClientName
		}
		if config.TLSEnabled {
			opts.TLSConfig = &tls.Config{InsecureSkipVerify: config.TLSSkipVerify}
		}
		rdb = redis.NewClusterClient(opts)
	} else {
		opts := &redis.Options{
			Network:         "tcp",
			Addr:            config.Addr,
			Username:        config.Username,
			Password:        config.Password,
			DB:              config.DB,
			MaxRetries:      config.MaxRetries,
			MinRetryBackoff: config.MinRetryBackoff,
			MaxRetryBackoff: config.MaxRetryBackoff,
			DialTimeout:     config.DialTimeout,
			ReadTimeout:     config.ReadTimeout,
			WriteTimeout:    config.WriteTimeout,
			PoolSize:        config.PoolSize,
			MinIdleConns:    config.MinIdleConns,
			PoolTimeout:     config.PoolTimeout,
			ConnMaxIdleTime: config.ConnMaxIdleTime,
		}
		if config.ClientName != "" {
			opts.ClientName = config.ClientName
		}
		if config.TLSEnabled {
			opts.TLSConfig = &tls.Config{InsecureSkipVerify: config.TLSSkipVerify}
		}
		rdb = redis.NewClient(opts)
	}

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

	// 返回字节数组，保持与 Set 方法存储格式的一致性
	// Set 方法存储的是 JSON 序列化后的字节数组，Get 应该返回相同格式
	return []byte(data), nil
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
	// 使用 SCAN 遍历，避免 KEYS 的阻塞与集群不兼容问题
	if pattern == "" {
		pattern = "*"
	}

	// 去重集合
	seen := make(map[string]struct{})
	keys := make([]string, 0, 256)

	// 集群模式：遍历所有主分片
	if cc, ok := r.client.(*redis.ClusterClient); ok {
		err := cc.ForEachMaster(ctx, func(ctx context.Context, cli *redis.Client) error {
			var cursor uint64
			for {
				res, next, err := cli.Scan(ctx, cursor, pattern, 1000).Result()
				if err != nil {
					return err
				}
				for _, k := range res {
					if _, exists := seen[k]; !exists {
						seen[k] = struct{}{}
						keys = append(keys, k)
					}
				}
				if next == 0 {
					break
				}
				cursor = next
			}
			return nil
		})
		return keys, err
	}

	// 单机或哨兵统一客户端：直接 SCAN
	var cursor uint64
	for {
		res, next, err := r.client.Scan(ctx, cursor, pattern, 1000).Result()
		if err != nil {
			return nil, err
		}
		for _, k := range res {
			if _, exists := seen[k]; !exists {
				seen[k] = struct{}{}
				keys = append(keys, k)
			}
		}
		if next == 0 {
			break
		}
		cursor = next
	}
	return keys, nil
}

// Close 关闭连接
func (r *RedisStorage) Close() error {
	return r.client.Close()
}
