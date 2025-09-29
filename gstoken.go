package gstoken

import (
	"gstoken/internal/core"
	"gstoken/internal/storage"
	"gstoken/internal/token"
)

// GSToken 主要的认证框架实例
type GSToken struct {
	config    *core.Config
	storage   core.Storage
	generator core.TokenGenerator
}

// New 创建新的GSToken实例
func New(config *core.Config) *GSToken {
	gs := &GSToken{
		config: config,
	}

	// 初始化存储
	gs.initStorage()

	// 初始化Token生成器
	gs.generator = token.NewGenerator(config.TokenStyle)

	return gs
}

// initStorage 初始化存储
func (gs *GSToken) initStorage() {
	switch gs.config.Storage.Type {
	case "redis":
		gs.storage = storage.NewRedisStorage(gs.config.Redis)
	case "memory":
		gs.storage = storage.NewMemoryStorage()
	default:
		// 默认使用内存存储
		gs.storage = storage.NewMemoryStorage()
	}
}

// GetConfig 获取配置
func (gs *GSToken) GetConfig() *core.Config {
	return gs.config
}

// GetStorage 获取存储实例
func (gs *GSToken) GetStorage() core.Storage {
	return gs.storage
}

// GetTokenGenerator 获取Token生成器
func (gs *GSToken) GetTokenGenerator() core.TokenGenerator {
	return gs.generator
}
