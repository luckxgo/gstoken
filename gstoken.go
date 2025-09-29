package gstoken

import (
	"context"
	"fmt"
	"github.com/luckxgo/gstoken/auth"
	"github.com/luckxgo/gstoken/core"
	"github.com/luckxgo/gstoken/storage"
	"github.com/luckxgo/gstoken/token"
)

// GSToken 主要的认证框架实例
type GSToken struct {
	config     *core.Config
	storage    core.Storage
	generator  core.TokenGenerator
	engine     core.AuthEngine
	keyService *core.KeyService
}

// New 创建新的GSToken实例
func New(config *core.Config) *GSToken {
	gs := &GSToken{
		config: config,
	}

	// 初始化键服务
	gs.keyService = core.NewKeyService(config.KeyPrefix)

	// 初始化存储
	gs.initStorage()

	// 初始化Token生成器
	gs.generator = token.NewGenerator(config.TokenStyle)

	// 初始化认证引擎
	gs.engine = auth.NewEngine(config, gs.storage, gs.generator, gs.keyService)

	// 如果配置中设置了用户角色提供者，自动配置
	if config.UserRoleProvider != nil {
		gs.engine.GetPermissionService().SetUserRoleProvider(config.UserRoleProvider)
	}

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

// GetAuthEngine 获取认证引擎
func (gs *GSToken) GetAuthEngine() core.AuthEngine {
	return gs.engine
}

// 以下是便捷方法，直接调用认证引擎的方法

// Login 用户登录
func (gs *GSToken) Login(ctx context.Context, req *core.LoginRequest) (*core.LoginResponse, error) {
	return gs.engine.Login(ctx, req)
}

// Logout 用户登出
func (gs *GSToken) Logout(ctx context.Context, token string) error {
	return gs.engine.Logout(ctx, token)
}

// LogoutByUserID 根据用户ID登出所有会话
func (gs *GSToken) LogoutByUserID(ctx context.Context, userID string) error {
	// 直接通过引擎实现调用
	if engine, ok := gs.engine.(*auth.Engine); ok {
		return engine.LogoutByUserID(ctx, userID)
	}
	return fmt.Errorf("LogoutByUserID功能不可用")
}

// GetLoginInfo 获取登录信息
func (gs *GSToken) GetLoginInfo(ctx context.Context, token string) (*core.LoginInfo, error) {
	// 直接通过引擎实现调用
	if engine, ok := gs.engine.(*auth.Engine); ok {
		return engine.GetLoginInfo(ctx, token)
	}
	return nil, fmt.Errorf("GetLoginInfo功能不可用")
}

// IsLogin 检查是否已登录
func (gs *GSToken) IsLogin(ctx context.Context, token string) bool {
	_, err := gs.engine.Verify(ctx, token)
	return err == nil
}

// CheckPermission 检查权限
func (gs *GSToken) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	return gs.engine.CheckPermission(ctx, userID, permission)
}

// RefreshToken 刷新Token
func (gs *GSToken) RefreshToken(ctx context.Context, refreshToken string) (*core.LoginResponse, error) {
	// 直接通过引擎实现调用
	if engine, ok := gs.engine.(*auth.Engine); ok {
		return engine.RefreshToken(ctx, refreshToken)
	}
	return nil, fmt.Errorf("RefreshToken功能不可用")
}

// CheckRole 检查用户角色
func (gs *GSToken) CheckRole(ctx context.Context, userID string, roleID string) (bool, error) {
	return gs.engine.CheckRole(ctx, userID, roleID)
}

// GetPermissionService 获取权限服务
func (gs *GSToken) GetPermissionService() core.PermissionService {
	return gs.engine.GetPermissionService()
}
