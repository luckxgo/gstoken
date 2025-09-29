package core

import (
	"context"
	"time"
)

// AuthEngine 核心认证引擎接口
type AuthEngine interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	Verify(ctx context.Context, token string) (*UserInfo, error)
	CheckPermission(ctx context.Context, userID string, permission string) (bool, error)
	CheckRole(ctx context.Context, userID string, roleID string) (bool, error)
	GetPermissionService() PermissionService
}

// TokenGenerator Token生成器接口
type TokenGenerator interface {
	Generate(extra map[string]interface{}) (string, error)
	Parse(token string) (*TokenInfo, error)
	Refresh(token string) (string, error)
	// 注册自定义Token生成函数
	RegisterCustomFunc(fn CustomTokenFunc) error
	// 设置Token风格
	SetStyle(style TokenStyle) error
}

// Storage 统一存储接口
type Storage interface {
	Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	LogoutByUserID(ctx context.Context, userID string) error
	GetLoginInfo(ctx context.Context, token string) (*LoginInfo, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
}

// UserRoleProvider 用户角色提供者接口（由用户实现）
type UserRoleProvider interface {
	GetUserRoles(ctx context.Context, userID string) ([]Role, error)
}

// PermissionService 权限服务接口
type PermissionService interface {
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)
	CheckRole(ctx context.Context, userID, roleID string) (bool, error)
	// 设置用户角色提供者
	SetUserRoleProvider(provider UserRoleProvider)
}

// SessionService 会话服务接口
type SessionService interface {
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, token string) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, token string) error
	KickOut(ctx context.Context, userID string) error
	KickOutByToken(ctx context.Context, token string) error
}
