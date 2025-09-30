package config

import (
	"time"

	"github.com/luckxgo/gstoken/core"
)

// DefaultConfig 默认配置
func DefaultConfig() *core.Config {
	return &core.Config{
		// Token配置
		TokenExpire:   24 * time.Hour,     // Token过期时间24小时
		RefreshExpire: 7 * 24 * time.Hour, // 刷新Token过期时间7天
		TokenStyle:    core.StyleUUID,     // 默认UUID风格

		// 登录配置
		LoginMode:    core.MultiLogin,          // 默认多端登录
		AutoRenew:    true,                     // 自动续期
		RememberDays: core.DefaultRememberDays, // 记住登录7天

		// 键前缀配置
		KeyPrefix: core.DefaultKeyPrefix, // 默认键前缀

		// 存储配置
		Storage: core.StorageConfig{
			Type: core.StorageTypeMemory, // 默认内存存储
		},

		// Redis配置
		Redis: core.RedisConfig{
			Addr:            core.DefaultRedisAddr,
			Username:        core.DefaultRedisUsername,
			Password:        core.DefaultRedisPassword,
			DB:              core.DefaultRedisDB,
			PoolSize:        core.DefaultRedisPoolSize,
			MinIdleConns:    core.DefaultRedisMinIdleConns,
			ConnMaxIdleTime: core.DefaultRedisConnMaxIdleTime,
			ClientName:      core.DefaultRedisClientName,
			MaxRetries:      core.DefaultRedisMaxRetries,
			MinRetryBackoff: core.DefaultRedisMinRetryBackoff,
			MaxRetryBackoff: core.DefaultRedisMaxRetryBackoff,
			DialTimeout:     core.DefaultRedisDialTimeout,
			ReadTimeout:     core.DefaultRedisReadTimeout,
			WriteTimeout:    core.DefaultRedisWriteTimeout,
			PoolTimeout:     core.DefaultRedisPoolTimeout,
			ClusterEnabled:  false,
			ClusterAddrs:    nil,
			TLSEnabled:      core.DefaultRedisTLSEnabled,
			TLSSkipVerify:   core.DefaultRedisTLSSkipVerify,
		},
	}
}

// RedisConfig Redis存储配置
func RedisConfig() *core.Config {
	config := DefaultConfig()
	config.Storage.Type = core.StorageTypeRedis
	return config
}
