package config

import (
	"time"

	"github.com/luckxgo/gstoken/core"
)

// ConfigBuilder 配置构建器
type ConfigBuilder struct {
	config *core.Config
}

// NewBuilder 创建配置构建器
func NewBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: DefaultConfig(),
	}
}

// WithUserRoleProvider 设置用户角色提供者
func (b *ConfigBuilder) WithUserRoleProvider(provider core.UserRoleProvider) *ConfigBuilder {
	b.config.UserRoleProvider = provider
	return b
}

// WithTokenExpire 设置Token过期时间
func (b *ConfigBuilder) WithTokenExpire(expire time.Duration) *ConfigBuilder {
	b.config.TokenExpire = expire
	return b
}

// WithRefreshExpire 设置RefreshToken过期时间
func (b *ConfigBuilder) WithRefreshExpire(expire time.Duration) *ConfigBuilder {
	b.config.RefreshExpire = expire
	return b
}

// WithTokenStyle 设置Token风格
func (b *ConfigBuilder) WithTokenStyle(style core.TokenStyle) *ConfigBuilder {
	b.config.TokenStyle = style
	return b
}

// WithLoginMode 设置登录模式
func (b *ConfigBuilder) WithLoginMode(mode core.LoginMode) *ConfigBuilder {
	b.config.LoginMode = mode
	return b
}

// WithAutoRenew 设置自动续期
func (b *ConfigBuilder) WithAutoRenew(autoRenew bool) *ConfigBuilder {
	b.config.AutoRenew = autoRenew
	return b
}

// WithRememberDays 设置记住登录天数
func (b *ConfigBuilder) WithRememberDays(days int) *ConfigBuilder {
	b.config.RememberDays = days
	return b
}

// WithRedisStorage 设置Redis存储
func (b *ConfigBuilder) WithRedisStorage(addr, password string, db int) *ConfigBuilder {
	b.config.Storage.Type = core.StorageTypeRedis
	b.config.Redis.Addr = addr
	b.config.Redis.Password = password
	b.config.Redis.DB = db
	// 关闭集群模式
	b.config.Redis.ClusterEnabled = false
	b.config.Redis.ClusterAddrs = nil
	return b
}

// WithRedisUsername 设置用户名
func (b *ConfigBuilder) WithRedisUsername(username string) *ConfigBuilder {
	b.config.Redis.Username = username
	return b
}

// WithRedisPool 设置连接池相关参数
func (b *ConfigBuilder) WithRedisPool(poolSize, minIdle int, idleTimeout time.Duration) *ConfigBuilder {
	b.config.Redis.PoolSize = poolSize
	b.config.Redis.MinIdleConns = minIdle
	b.config.Redis.ConnMaxIdleTime = idleTimeout
	return b
}

// WithRedisTimeouts 设置连接与读写超时以及池等待超时
func (b *ConfigBuilder) WithRedisTimeouts(dial, read, write, pool time.Duration) *ConfigBuilder {
	b.config.Redis.DialTimeout = dial
	b.config.Redis.ReadTimeout = read
	b.config.Redis.WriteTimeout = write
	b.config.Redis.PoolTimeout = pool
	return b
}

// WithRedisRetries 设置重试参数
func (b *ConfigBuilder) WithRedisRetries(max int, minBackoff, maxBackoff time.Duration) *ConfigBuilder {
	b.config.Redis.MaxRetries = max
	b.config.Redis.MinRetryBackoff = minBackoff
	b.config.Redis.MaxRetryBackoff = maxBackoff
	return b
}

// WithRedisClientName 设置客户端名称
func (b *ConfigBuilder) WithRedisClientName(name string) *ConfigBuilder {
	b.config.Redis.ClientName = name
	return b
}

// WithRedisTLS 设置 TLS 开关（基础）
func (b *ConfigBuilder) WithRedisTLS(enable bool, skipVerify bool) *ConfigBuilder {
	b.config.Redis.TLSEnabled = enable
	b.config.Redis.TLSSkipVerify = skipVerify
	return b
}

// WithRedisCluster 设置Redis集群存储
func (b *ConfigBuilder) WithRedisCluster(addrs []string, password string) *ConfigBuilder {
	b.config.Storage.Type = core.StorageTypeRedis
	b.config.Redis.ClusterEnabled = true
	b.config.Redis.ClusterAddrs = addrs
	b.config.Redis.Password = password
	// 集群模式不使用单机的 Addr/DB
	b.config.Redis.Addr = ""
	b.config.Redis.DB = 0
	return b
}

// WithMemoryStorage 设置内存存储
func (b *ConfigBuilder) WithMemoryStorage() *ConfigBuilder {
	b.config.Storage.Type = core.StorageTypeMemory
	return b
}

// WithDatabaseStorage 设置数据库存储
func (b *ConfigBuilder) WithDatabaseStorage(driver, host string, port int, username, password, database string) *ConfigBuilder {
	b.config.Storage.Type = core.StorageTypeDatabase
	b.config.Database.Driver = driver
	b.config.Database.Host = host
	b.config.Database.Port = port
	b.config.Database.Username = username
	b.config.Database.Password = password
	b.config.Database.Database = database
	return b
}

// Build 构建配置
func (b *ConfigBuilder) Build() *core.Config {
	return b.config
}
