package web

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/luckxgo/gstoken/core"
)

// WebContext 通用的 Web 上下文接口，用于适配不同的 Web 框架
type WebContext interface {
	// GetHeader 获取请求头
	GetHeader(key string) string

	// SetHeader 设置响应头
	SetHeader(key, value string)

	// GetQuery 获取查询参数
	GetQuery(key string) string

	// GetParam 获取路径参数
	GetParam(key string) string

	// GetBody 获取请求体
	GetBody() ([]byte, error)

	// JSON 返回 JSON 响应
	JSON(code int, obj interface{})

	// String 返回字符串响应
	String(code int, format string, values ...interface{})

	// Status 设置状态码
	Status(code int)

	// Abort 中止请求处理
	Abort()

	// AbortWithStatus 中止请求并设置状态码
	AbortWithStatus(code int)

	// AbortWithJSON 中止请求并返回 JSON
	AbortWithJSON(code int, obj interface{})

	// Set 设置上下文值
	Set(key string, value interface{})

	// Get 获取上下文值
	Get(key string) (interface{}, bool)

	// GetContext 获取原始 context.Context
	GetContext() context.Context

	// GetRequest 获取原始 HTTP 请求
	GetRequest() *http.Request

	// GetResponseWriter 获取原始 HTTP 响应写入器
	GetResponseWriter() http.ResponseWriter

	// Next 调用下一个中间件或处理器
	Next()
}

// AuthMiddleware 通用认证中间件接口
type AuthMiddleware interface {
	// RequireAuth 要求认证的中间件
	RequireAuth() MiddlewareFunc

	// RequirePermission 要求特定权限的中间件
	RequirePermission(permission string) MiddlewareFunc

	// RequireRole 要求特定角色的中间件
	RequireRole(role string) MiddlewareFunc

	// RequireAnyPermission 要求任意权限的中间件
	RequireAnyPermission(permissions ...string) MiddlewareFunc

	// RequireAllPermissions 要求所有权限的中间件
	RequireAllPermissions(permissions ...string) MiddlewareFunc

	// RequireAnyRole 要求任意角色的中间件
	RequireAnyRole(roles ...string) MiddlewareFunc

	// RequireAllRoles 要求所有角色的中间件
	RequireAllRoles(roles ...string) MiddlewareFunc

	// OptionalAuth 可选认证的中间件（不强制要求登录）
	OptionalAuth() MiddlewareFunc
}

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(WebContext)

// AuthConfig 认证配置
type AuthConfig struct {
	// Token 提取配置
	TokenHeader string // Token 在请求头中的字段名，默认 "Authorization"
	TokenQuery  string // Token 在查询参数中的字段名，默认 "token"
	TokenPrefix string // Token 前缀，默认 "Bearer "

	// 跳过认证的路径
	SkipPaths []string

	// 错误处理
	UnauthorizedHandler func(WebContext, error)
	ForbiddenHandler    func(WebContext, error)

	// 用户信息提取器
	UserInfoExtractor func(ctx context.Context, token string) (*core.UserInfo, error)
}

// DefaultAuthConfig 默认认证配置
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		TokenHeader: HeaderAuthorization,
		TokenQuery:  QueryParamToken,
		TokenPrefix: BearerPrefix,
		SkipPaths:   []string{},
		UnauthorizedHandler: func(c WebContext, err error) {
			c.AbortWithJSON(http.StatusUnauthorized, map[string]interface{}{
				"error":      ErrorUnauthorized,
				ErrorMessage: err.Error(),
			})
		},
		ForbiddenHandler: func(c WebContext, err error) {
			c.AbortWithJSON(http.StatusForbidden, map[string]interface{}{
				"error":      ErrorForbidden,
				ErrorMessage: err.Error(),
			})
		},
	}
}

// BaseAuthMiddleware 基础认证中间件实现
type BaseAuthMiddleware struct {
	gsToken GSTokenAdapter
	config  *AuthConfig
}

// GSTokenAdapter GSToken 适配器接口
type GSTokenAdapter interface {
	// Verify 验证 Token
	Verify(ctx context.Context, token string) (*core.UserInfo, error)

	// CheckPermission 检查权限
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)

	// CheckRole 检查角色
	CheckRole(ctx context.Context, userID, role string) (bool, error)

	// GetLoginInfo 获取登录信息
	GetLoginInfo(ctx context.Context, token string) (*core.LoginInfo, error)
}

// NewBaseAuthMiddleware 创建基础认证中间件
func NewBaseAuthMiddleware(gsToken GSTokenAdapter, config *AuthConfig) *BaseAuthMiddleware {
	if config == nil {
		config = DefaultAuthConfig()
	}

	return &BaseAuthMiddleware{
		gsToken: gsToken,
		config:  config,
	}
}

// extractToken 从请求中提取 Token
func (m *BaseAuthMiddleware) extractToken(c WebContext) string {
	// 从请求头提取
	if token := c.GetHeader(m.config.TokenHeader); token != "" {
		if m.config.TokenPrefix != "" && len(token) > len(m.config.TokenPrefix) {
			if token[:len(m.config.TokenPrefix)] == m.config.TokenPrefix {
				return token[len(m.config.TokenPrefix):]
			}
		} else if m.config.TokenPrefix == "" {
			return token
		}
	}

	// 从查询参数提取
	if m.config.TokenQuery != "" {
		if token := c.GetQuery(m.config.TokenQuery); token != "" {
			return token
		}
	}

	return ""
}

 // softAuth 在不强制鉴权情况下尽可能提取用户信息（不阻断流程）
func (m *BaseAuthMiddleware) softAuth(c WebContext) {
	token := m.extractToken(c)
	if token == "" {
		return
	}
	if userInfo, err := m.gsToken.Verify(c.GetContext(), token); err == nil && userInfo != nil {
		// 将用户信息存储到上下文
		c.Set(ContextKeyUserID, userInfo.ID)
		c.Set(ContextKeyToken, token)
		c.Set(ContextKeyUserInfo, userInfo)

		// 如果配置了用户信息提取器，获取完整用户信息
		if m.config.UserInfoExtractor != nil {
			if ui, err := m.config.UserInfoExtractor(c.GetContext(), token); err == nil && ui != nil {
				c.Set(ContextKeyUserInfo, ui)
			}
		}
	}
}

// shouldSkip 检查是否应该跳过认证
func (m *BaseAuthMiddleware) shouldSkip(c WebContext) bool {
	reqPath := c.GetRequest().URL.Path
	for _, pat := range m.config.SkipPaths {
		// 精确匹配
		if reqPath == pat {
			return true
		}
		// 通配符匹配（* 或 ?）
		if strings.ContainsAny(pat, "*?") {
			if ok, _ := path.Match(pat, reqPath); ok {
				return true
			}
		}
		// 前缀匹配：以 /* 结尾表示前缀
		if strings.HasSuffix(pat, "/*") {
			prefix := strings.TrimSuffix(pat, "/*")
			if strings.HasPrefix(reqPath, prefix+"/") || reqPath == prefix {
				return true
			}
		}
	}
	return false
}

// RequireAuth 要求认证的中间件
func (m *BaseAuthMiddleware) RequireAuth() MiddlewareFunc {
	return func(c WebContext) {
		if m.shouldSkip(c) {
			// 跳过强制鉴权，但若携带 token，则尝试提取用户信息
			m.softAuth(c)
			c.Next()
			return
		}

		token := m.extractToken(c)
		if token == "" {
			m.config.UnauthorizedHandler(c, core.ErrTokenNotFound)
			return
		}

		userInfo, err := m.gsToken.Verify(c.GetContext(), token)
		if err != nil {
			m.config.UnauthorizedHandler(c, err)
			return
		}

		// 将用户信息存储到上下文
		c.Set(ContextKeyUserID, userInfo.ID)
		c.Set(ContextKeyToken, token)
		c.Set(ContextKeyUserInfo, userInfo)

		// 如果配置了用户信息提取器，获取完整用户信息
		if m.config.UserInfoExtractor != nil {
			userInfo, err := m.config.UserInfoExtractor(c.GetContext(), token)
			if err == nil {
				c.Set(ContextKeyUserInfo, userInfo)
			}
		}

		c.Next()
	}
}

// RequirePermission 要求特定权限的中间件
func (m *BaseAuthMiddleware) RequirePermission(permission string) MiddlewareFunc {
	return func(c WebContext) {
		if m.shouldSkip(c) {
			// 跳过强制鉴权，但若携带 token，则尝试提取用户信息
			m.softAuth(c)
			c.Next()
			return
		}

		token := m.extractToken(c)
		if token == "" {
			m.config.UnauthorizedHandler(c, core.ErrTokenNotFound)
			return
		}

		userInfo, err := m.gsToken.Verify(c.GetContext(), token)
		if err != nil {
			m.config.UnauthorizedHandler(c, err)
			return
		}

		hasPermission, err := m.gsToken.CheckPermission(c.GetContext(), userInfo.ID, permission)
		if err != nil {
			m.config.ForbiddenHandler(c, err)
			return
		}

		if !hasPermission {
			m.config.ForbiddenHandler(c, core.ErrPermissionDenied)
			return
		}

		// 将用户信息存储到上下文
		c.Set(ContextKeyUserID, userInfo.ID)
		c.Set(ContextKeyToken, token)
		c.Set(ContextKeyUserInfo, userInfo)

		c.Next()
	}
}

// RequireRole 要求特定角色的中间件
func (m *BaseAuthMiddleware) RequireRole(role string) MiddlewareFunc {
	return func(c WebContext) {
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		token := m.extractToken(c)
		if token == "" {
			m.config.UnauthorizedHandler(c, core.ErrTokenNotFound)
			return
		}

		userInfo, err := m.gsToken.Verify(c.GetContext(), token)
		if err != nil {
			m.config.UnauthorizedHandler(c, err)
			return
		}

		hasRole, err := m.gsToken.CheckRole(c.GetContext(), userInfo.ID, role)
		if err != nil {
			m.config.ForbiddenHandler(c, err)
			return
		}

		if !hasRole {
			m.config.ForbiddenHandler(c, core.ErrRoleNotFound)
			return
		}

		// 将用户信息存储到上下文
		c.Set(ContextKeyUserID, userInfo.ID)
		c.Set(ContextKeyToken, token)
		c.Set(ContextKeyUserInfo, userInfo)

		c.Next()
	}
}

// RequireAnyPermission 要求任意权限的中间件
func (m *BaseAuthMiddleware) RequireAnyPermission(permissions ...string) MiddlewareFunc {
	return func(c WebContext) {
		// 先执行认证
		authMiddleware := m.RequireAuth()
		authMiddleware(c)

		if userID, exists := c.Get(ContextKeyUserID); exists {
			if userIDStr, ok := userID.(string); ok {
				for _, permission := range permissions {
					hasPermission, err := m.gsToken.CheckPermission(c.GetContext(), userIDStr, permission)
					if err == nil && hasPermission {
						c.Next()
						return
					}
				}

				m.config.ForbiddenHandler(c, core.ErrPermissionDenied)
				return
			}
		}

		c.Next()
	}
}

// RequireAllPermissions 要求所有权限的中间件
func (m *BaseAuthMiddleware) RequireAllPermissions(permissions ...string) MiddlewareFunc {
	return func(c WebContext) {
		// 先执行认证
		authMiddleware := m.RequireAuth()
		authMiddleware(c)

		if userID, exists := c.Get(ContextKeyUserID); exists {
			if userIDStr, ok := userID.(string); ok {
				for _, permission := range permissions {
					hasPermission, err := m.gsToken.CheckPermission(c.GetContext(), userIDStr, permission)
					if err != nil || !hasPermission {
						m.config.ForbiddenHandler(c, core.ErrPermissionDenied)
						return
					}
				}
			}
		}

		c.Next()
	}
}

// RequireAnyRole 要求任意角色的中间件
func (m *BaseAuthMiddleware) RequireAnyRole(roles ...string) MiddlewareFunc {
	return func(c WebContext) {
		// 先执行认证
		authMiddleware := m.RequireAuth()
		authMiddleware(c)

		if userID, exists := c.Get("user_id"); exists {
			if userIDStr, ok := userID.(string); ok {
				for _, role := range roles {
					hasRole, err := m.gsToken.CheckRole(c.GetContext(), userIDStr, role)
					if err == nil && hasRole {
						c.Next()
						return
					}
				}

				m.config.ForbiddenHandler(c, core.ErrRoleNotFound)
				return
			}
		}

		c.Next()
	}
}

// RequireAllRoles 要求所有角色的中间件
func (m *BaseAuthMiddleware) RequireAllRoles(roles ...string) MiddlewareFunc {
	return func(c WebContext) {
		// 先执行认证
		authMiddleware := m.RequireAuth()
		authMiddleware(c)

		if userID, exists := c.Get("user_id"); exists {
			if userIDStr, ok := userID.(string); ok {
				for _, role := range roles {
					hasRole, err := m.gsToken.CheckRole(c.GetContext(), userIDStr, role)
					if err != nil || !hasRole {
						m.config.ForbiddenHandler(c, core.ErrRoleNotFound)
						return
					}
				}
			}
		}

		c.Next()
	}
}

// OptionalAuth 可选认证的中间件（不强制要求登录）
func (m *BaseAuthMiddleware) OptionalAuth() MiddlewareFunc {
	return func(c WebContext) {
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		userInfo, err := m.gsToken.Verify(c.GetContext(), token)
		if err != nil {
			c.Next()
			return
		}

		// 将用户信息存储到上下文
		c.Set(ContextKeyUserID, userInfo.ID)
		c.Set(ContextKeyToken, token)
		c.Set(ContextKeyUserInfo, userInfo)

		// 如果配置了用户信息提取器，获取完整用户信息
		if m.config.UserInfoExtractor != nil {
			userInfo, err := m.config.UserInfoExtractor(c.GetContext(), token)
			if err == nil {
				c.Set(ContextKeyUserInfo, userInfo)
			}
		}

		c.Next()
	}
}
