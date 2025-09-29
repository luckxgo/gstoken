package web

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinContext Gin 框架的 WebContext 实现
type GinContext struct {
	*gin.Context
}

// NewGinContext 创建 Gin 上下文适配器
func NewGinContext(c *gin.Context) *GinContext {
	return &GinContext{Context: c}
}

// GetHeader 获取请求头
func (c *GinContext) GetHeader(key string) string {
	return c.Context.GetHeader(key)
}

// SetHeader 设置响应头
func (c *GinContext) SetHeader(key, value string) {
	c.Context.Header(key, value)
}

// GetQuery 获取查询参数
func (c *GinContext) GetQuery(key string) string {
	return c.Context.Query(key)
}

// GetParam 获取路径参数
func (c *GinContext) GetParam(key string) string {
	return c.Context.Param(key)
}

// GetBody 获取请求体
func (c *GinContext) GetBody() ([]byte, error) {
	return io.ReadAll(c.Context.Request.Body)
}

// JSON 返回 JSON 响应
func (c *GinContext) JSON(code int, obj interface{}) {
	c.Context.JSON(code, obj)
}

// String 返回字符串响应
func (c *GinContext) String(code int, format string, values ...interface{}) {
	c.Context.String(code, format, values...)
}

// Status 设置状态码
func (c *GinContext) Status(code int) {
	c.Context.Status(code)
}

// Abort 中止请求处理
func (c *GinContext) Abort() {
	c.Context.Abort()
}

// AbortWithStatus 中止请求并设置状态码
func (c *GinContext) AbortWithStatus(code int) {
	c.Context.AbortWithStatus(code)
}

// AbortWithJSON 中止请求并返回 JSON
func (c *GinContext) AbortWithJSON(code int, obj interface{}) {
	c.Context.JSON(code, obj)
	c.Context.Abort()
}

// Set 设置上下文值
func (c *GinContext) Set(key string, value interface{}) {
	c.Context.Set(key, value)
}

// Get 获取上下文值
func (c *GinContext) Get(key string) (interface{}, bool) {
	return c.Context.Get(key)
}

// GetContext 获取原始 context.Context
func (c *GinContext) GetContext() context.Context {
	return c.Context.Request.Context()
}

// GetRequest 获取原始 HTTP 请求
func (c *GinContext) GetRequest() *http.Request {
	return c.Context.Request
}

// GetResponseWriter 获取原始 HTTP 响应写入器
func (c *GinContext) GetResponseWriter() http.ResponseWriter {
	return c.Context.Writer
}

// Next 调用下一个中间件或处理器
func (c *GinContext) Next() {
	c.Context.Next()
}

// GinAuthMiddleware Gin 框架的认证中间件
type GinAuthMiddleware struct {
	*BaseAuthMiddleware
}

// NewGinAuthMiddleware 创建 Gin 认证中间件
func NewGinAuthMiddleware(gsToken GSTokenAdapter, config *AuthConfig) *GinAuthMiddleware {
	return &GinAuthMiddleware{
		BaseAuthMiddleware: NewBaseAuthMiddleware(gsToken, config),
	}
}

// RequireAuth 要求认证的 Gin 中间件
func (m *GinAuthMiddleware) RequireAuth() gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.RequireAuth()
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}

// RequirePermission 要求特定权限的 Gin 中间件
func (m *GinAuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.RequirePermission(permission)
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}

// RequireRole 要求特定角色的 Gin 中间件
func (m *GinAuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.RequireRole(role)
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}

// RequireAnyPermission 要求任意权限的 Gin 中间件
func (m *GinAuthMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.RequireAnyPermission(permissions...)
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}

// RequireAllPermissions 要求所有权限的 Gin 中间件
func (m *GinAuthMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.RequireAllPermissions(permissions...)
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}

// RequireAnyRole 要求任意角色的 Gin 中间件
func (m *GinAuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.RequireAnyRole(roles...)
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}

// RequireAllRoles 要求所有角色的 Gin 中间件
func (m *GinAuthMiddleware) RequireAllRoles(roles ...string) gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.RequireAllRoles(roles...)
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}

// OptionalAuth 可选认证的 Gin 中间件
func (m *GinAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.OptionalAuth()
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}

// GinHelper Gin 框架的辅助函数
type GinHelper struct{}

// GetUserID 从 Gin 上下文中获取用户 ID
func (h *GinHelper) GetUserID(c *gin.Context) (string, bool) {
	if userID, exists := c.Get(ContextKeyUserID); exists {
		if userIDStr, ok := userID.(string); ok {
			return userIDStr, true
		}
	}
	return "", false
}

// GetToken 从 Gin 上下文中获取 Token
func (h *GinHelper) GetToken(c *gin.Context) (string, bool) {
	if token, exists := c.Get(ContextKeyToken); exists {
		if tokenStr, ok := token.(string); ok {
			return tokenStr, true
		}
	}
	return "", false
}

// GetUserInfo 从 Gin 上下文中获取用户信息
func (h *GinHelper) GetUserInfo(c *gin.Context) (interface{}, bool) {
	return c.Get(ContextKeyUserInfo)
}

// MustGetUserID 从 Gin 上下文中获取用户 ID，如果不存在则 panic
func (h *GinHelper) MustGetUserID(c *gin.Context) string {
	userID, exists := h.GetUserID(c)
	if !exists {
		panic(ContextKeyUserID + " not found in context")
	}
	return userID
}

// MustGetToken 从 Gin 上下文中获取 Token，如果不存在则 panic
func (h *GinHelper) MustGetToken(c *gin.Context) string {
	token, exists := h.GetToken(c)
	if !exists {
		panic(ContextKeyToken + " not found in context")
	}
	return token
}

// 全局 Gin 辅助实例
var Gin = &GinHelper{}
