package web

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/luckxgo/gstoken/core"
)

/*
ContextHelper 通用 Web 上下文辅助工具

当前实现基于 Gin 的 *gin.Context；后续可通过适配器支持其他 Web 框架。
*/
type ContextHelper struct{}

// GetUserID 从 Gin 上下文中获取用户 ID
func (h *ContextHelper) GetUserID(c *gin.Context) (string, bool) {
	if userID, exists := c.Get(ContextKeyUserID); exists {
		if userIDStr, ok := userID.(string); ok {
			return userIDStr, true
		}
	}
	return "", false
}

// GetToken 从 Gin 上下文中获取 Token
func (h *ContextHelper) GetToken(c *gin.Context) (string, bool) {
	if token, exists := c.Get(ContextKeyToken); exists {
		if tokenStr, ok := token.(string); ok {
			return tokenStr, true
		}
	}
	return "", false
}

// GetUserInfo 从 Gin 上下文中获取用户信息（强类型）
func (h *ContextHelper) GetUserInfo(c *gin.Context) (*core.UserInfo, bool) {
	val, ok := c.Get(ContextKeyUserInfo)
	if !ok || val == nil {
		return nil, false
	}
	if ui, ok := val.(*core.UserInfo); ok {
		return ui, true
	}
	// 兼容场景：如果以值类型或其他形式设置，尝试断言
	if ui, ok := val.(core.UserInfo); ok {
		return &ui, true
	}
	return nil, false
}

// MustGetUserID 从 Gin 上下文中获取用户 ID，如果不存在则 panic
func (h *ContextHelper) MustGetUserID(c *gin.Context) string {
	userID, exists := h.GetUserID(c)
	if !exists {
		panic(ContextKeyUserID + " not found in context")
	}
	return userID
}

// MustGetToken 从 Gin 上下文中获取 Token，如果不存在则 panic
func (h *ContextHelper) MustGetToken(c *gin.Context) string {
	token, exists := h.GetToken(c)
	if !exists {
		panic(ContextKeyToken + " not found in context")
	}
	return token
}

// LogoutFromGinContext 从 Gin 上下文中获取 token 并执行登出
// 需要传入 GSToken 实例来执行实际的登出操作
func (h *ContextHelper) LogoutFromGinContext(c *gin.Context, gs interface {
	LogoutFromContext(ctx context.Context) error
}) error {
	// 从 Gin 上下文获取 token 并设置到标准库 context 中
	token, exists := h.GetToken(c)
	if !exists {
		return fmt.Errorf("token not found in gin context")
	}

	// 创建包含 token 的上下文
	ctx := context.WithValue(c.Request.Context(), ContextKeyToken, token)

	// 执行登出
	return gs.LogoutFromContext(ctx)
}

// 全局通用上下文辅助实例
var Helper = &ContextHelper{}
