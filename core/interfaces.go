package core

import (
	"context"
	"time"
)

// AuthEngine 核心认证引擎接口
type AuthEngine interface {
	// Login 用户登录，返回Token和用户信息
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)

	// Logout 用户登出，使指定Token失效
	Logout(ctx context.Context, token string) error

	// Verify 验证Token有效性，返回用户信息
	Verify(ctx context.Context, token string) (*UserInfo, error)

	// CheckPermission 检查用户是否拥有指定权限
	CheckPermission(ctx context.Context, userID string, permission string) (bool, error)

	// CheckRole 检查用户是否拥有指定角色
	CheckRole(ctx context.Context, userID string, roleID string) (bool, error)

	// GetPermissionService 获取权限服务实例
	GetPermissionService() PermissionService
}

// TokenGenerator Token生成器接口
type TokenGenerator interface {
	// Generate 生成Token，可传入额外参数
	Generate(extra map[string]interface{}) (string, error)

	// Parse 解析Token，返回Token信息
	Parse(token string) (*TokenInfo, error)

	// Refresh 刷新Token，生成新的Token
	Refresh(token string) (string, error)

	// RegisterCustomFunc 注册自定义Token生成函数
	RegisterCustomFunc(fn CustomTokenFunc) error

	// SetStyle 设置Token生成风格
	SetStyle(style TokenStyle) error
}

// Storage 统一存储接口
type Storage interface {
	// Set 设置键值对，支持过期时间
	Set(ctx context.Context, key string, value interface{}, expire time.Duration) error

	// Get 根据键获取值
	Get(ctx context.Context, key string) (interface{}, error)

	// Delete 删除指定键
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Keys 根据模式匹配获取键列表
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// AuthService 认证服务接口
type AuthService interface {
	// Login 用户登录，处理登录逻辑并返回Token
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)

	// Logout 单个Token登出，清理会话信息
	Logout(ctx context.Context, token string) error

	// LogoutByUserID 根据用户ID登出所有会话
	LogoutByUserID(ctx context.Context, userID string) error

	// GetLoginInfo 获取Token对应的登录信息
	GetLoginInfo(ctx context.Context, token string) (*LoginInfo, error)

	// RefreshAccessToken 使用刷新Token获取新的访问Token
	RefreshAccessToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
}

// UserRoleProvider 用户角色提供者接口（由用户实现）
type UserRoleProvider interface {
	// GetUserRoles 获取指定用户的所有角色信息
	// 此方法需要由业务系统实现，框架不提供默认实现
	GetUserRoles(ctx context.Context, userID string) ([]Role, error)
}

// PermissionService 权限服务接口
type PermissionService interface {
	// CheckPermission 检查用户是否拥有指定权限
	// 通过用户角色提供者获取角色，然后验证权限
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)

	// CheckRole 检查用户是否拥有指定角色
	CheckRole(ctx context.Context, userID, roleID string) (bool, error)

	// SetUserRoleProvider 设置用户角色提供者
	// 必须在使用权限检查功能前调用此方法
	SetUserRoleProvider(provider UserRoleProvider)
}

// SessionService 会话服务接口
type SessionService interface {
	// CreateSession 创建新的用户会话
	CreateSession(ctx context.Context, session *Session) error

	// GetSession 根据Token获取会话信息
	GetSession(ctx context.Context, token string) (*Session, error)

	// UpdateSession 更新会话信息（如最后访问时间等）
	UpdateSession(ctx context.Context, session *Session) error

	// DeleteSession 删除指定Token的会话
	DeleteSession(ctx context.Context, token string) error

	// KickOut 踢出用户的所有会话（根据用户ID）
	KickOut(ctx context.Context, userID string) error

	// KickOutByToken 踢出指定Token的会话
	KickOutByToken(ctx context.Context, token string) error
}
