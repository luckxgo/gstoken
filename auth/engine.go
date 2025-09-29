package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gstoken/core"
)

// Engine 核心认证引擎实现
type Engine struct {
	config            *core.Config
	storage           core.Storage
	tokenGenerator    core.TokenGenerator
	authService       core.AuthService
	sessionService    core.SessionService
	permissionService core.PermissionService
}

// NewEngine 创建新的认证引擎
func NewEngine(config *core.Config, storage core.Storage, tokenGenerator core.TokenGenerator) *Engine {
	engine := &Engine{
		config:         config,
		storage:        storage,
		tokenGenerator: tokenGenerator,
	}

	// 初始化各个服务
	engine.sessionService = NewSessionService(storage, config)
	engine.authService = NewAuthService(storage, tokenGenerator, engine.sessionService, config)
	engine.permissionService = NewPermissionService(storage)

	return engine
}

// Login 用户登录
func (e *Engine) Login(ctx context.Context, req *core.LoginRequest) (*core.LoginResponse, error) {
	if req == nil {
		return nil, errors.New("登录请求不能为空")
	}

	if req.UserID == "" {
		return nil, errors.New("用户ID不能为空")
	}

	// 调用认证服务进行登录
	return e.authService.Login(ctx, req)
}

// Logout 用户登出
func (e *Engine) Logout(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("Token不能为空")
	}

	return e.authService.Logout(ctx, token)
}

// Verify 验证Token并获取用户信息
func (e *Engine) Verify(ctx context.Context, token string) (*core.UserInfo, error) {
	if token == "" {
		return nil, errors.New("Token不能为空")
	}

	// 获取登录信息
	loginInfo, err := e.authService.GetLoginInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("获取登录信息失败: %w", err)
	}

	// 检查Token是否过期
	if time.Now().After(loginInfo.LastAccess.Add(e.config.TokenExpire)) {
		return nil, errors.New("Token已过期")
	}

	// 更新最后访问时间
	session, err := e.sessionService.GetSession(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("获取会话信息失败: %w", err)
	}

	session.LastAccess = time.Now()
	if err := e.sessionService.UpdateSession(ctx, session); err != nil {
		// 更新失败不影响验证结果，只记录错误
		// 在实际项目中可以使用日志记录
	}

	// 构造用户信息
	userInfo := &core.UserInfo{
		ID:       loginInfo.UserID,
		Username: loginInfo.UserID, // 这里简化处理，实际项目中应该从用户服务获取
		Extra:    make(map[string]interface{}),
		Roles:    []string{}, // 角色信息需要通过用户自定义的 UserRoleProvider 获取
	}

	return userInfo, nil
}

// CheckPermission 检查用户权限
func (e *Engine) CheckPermission(ctx context.Context, userID string, permission string) (bool, error) {
	if userID == "" {
		return false, errors.New("用户ID不能为空")
	}

	if permission == "" {
		return false, errors.New("权限标识不能为空")
	}

	return e.permissionService.CheckPermission(ctx, userID, permission)
}

// CheckRole 检查用户角色
func (e *Engine) CheckRole(ctx context.Context, userID, roleID string) (bool, error) {
	return e.permissionService.CheckRole(ctx, userID, roleID)
}

// GetAuthService 获取认证服务
func (e *Engine) GetAuthService() core.AuthService {
	return e.authService
}

// GetSessionService 获取会话服务
func (e *Engine) GetSessionService() core.SessionService {
	return e.sessionService
}

// GetPermissionService 获取权限服务
func (e *Engine) GetPermissionService() core.PermissionService {
	return e.permissionService
}

// LogoutByUserID 根据用户ID登出所有会话
func (e *Engine) LogoutByUserID(ctx context.Context, userID string) error {
	return e.authService.LogoutByUserID(ctx, userID)
}

// GetLoginInfo 获取登录信息
func (e *Engine) GetLoginInfo(ctx context.Context, token string) (*core.LoginInfo, error) {
	return e.authService.GetLoginInfo(ctx, token)
}

// RefreshToken 刷新Token
func (e *Engine) RefreshToken(ctx context.Context, refreshToken string) (*core.LoginResponse, error) {
	return e.authService.RefreshAccessToken(ctx, refreshToken)
}
