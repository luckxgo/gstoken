package web

import "github.com/gin-gonic/gin"

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
