package config

import (
	"github.com/luckxgo/gstoken/core"
	"time"
)

// DefaultConfig 默认配置
func DefaultConfig() *core.Config {
	return &core.Config{
		// Token配置
		TokenExpire:   24 * time.Hour,     // Token过期时间24小时
		RefreshExpire: 7 * 24 * time.Hour, // 刷新Token过期时间7天
		TokenStyle:    core.StyleUUID,     // 默认UUID风格

		// 登录配置
		LoginMode:    core.MultiLogin, // 默认多端登录
		AutoRenew:    true,            // 自动续期
		RememberDays: 7,               // 记住登录7天

		// 键前缀配置
		KeyPrefix: "gstoken", // 默认键前缀

		// 存储配置
		Storage: core.StorageConfig{
			Type: "memory", // 默认内存存储
		},

		// Redis配置
		Redis: core.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			PoolSize: 10,
		},

		// 数据库配置
		Database: core.DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "gstoken",
		},
	}
}

// RedisConfig Redis存储配置
func RedisConfig() *core.Config {
	config := DefaultConfig()
	config.Storage.Type = "redis"
	return config
}
