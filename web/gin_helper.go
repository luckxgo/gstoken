package web

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

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

// LogoutFromGinContext 从 Gin 上下文中获取 token 并执行登出
// 需要传入 GSToken 实例来执行实际的登出操作
func (h *GinHelper) LogoutFromGinContext(c *gin.Context, gs interface {
	LogoutFromContext(ctx context.Context) error
}) error {
	// 从 Gin 上下文获取 token 并设置到标准库 context 中
	token, exists := h.GetToken(c)
	if !exists {
		return fmt.Errorf("token not found in gin context")
	}

	// 创建包含 token 的上下文
	ctx := context.WithValue(c.Request.Context(), "token", token)

	// 执行登出
	return gs.LogoutFromContext(ctx)
}

// 全局 Gin 辅助实例
var Helper = &GinHelper{}
