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
	ctx context.Context // 用于存储标准库 context 值
}

// NewGinContext 创建 Gin 上下文适配器
func NewGinContext(c *gin.Context) *GinContext {
	return &GinContext{
		Context: c,
		ctx:     c.Request.Context(),
	}
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
	// 设置到 Gin 上下文
	c.Context.Set(key, value)
	// 同时设置到标准库 context
	c.ctx = context.WithValue(c.ctx, key, value)
	// 更新请求的 context
	c.Context.Request = c.Context.Request.WithContext(c.ctx)
}

// Get 获取上下文值
func (c *GinContext) Get(key string) (interface{}, bool) {
	// 优先从 Gin 上下文获取
	if value, exists := c.Context.Get(key); exists {
		return value, true
	}
	// 如果 Gin 上下文中没有，尝试从标准库 context 获取
	if value := c.ctx.Value(key); value != nil {
		return value, true
	}
	return nil, false
}

// GetContext 获取原始 context.Context
func (c *GinContext) GetContext() context.Context {
	return c.ctx
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

/*
*
OptionalAuth 可选认证的 Gin 中间件
*/
func (m *GinAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.OptionalAuth()
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}

// RequireRoleOrPermission 任意满足角色或权限即放行（Gin 适配）
func (m *GinAuthMiddleware) RequireRoleOrPermission(roles []string, perms []string) gin.HandlerFunc {
	middlewareFunc := m.BaseAuthMiddleware.RequireRoleOrPermission(roles, perms)
	return func(c *gin.Context) {
		middlewareFunc(NewGinContext(c))
	}
}
