package web

import (
	"context"

	"github.com/luckxgo/gstoken/core"
)

// GSTokenWebAdapter GSToken 的 Web 适配器实现
type GSTokenWebAdapter struct {
	gsToken GSTokenInterface
}

// GSTokenInterface GSToken 接口定义
type GSTokenInterface interface {
	// GetAuthEngine 获取认证引擎
	GetAuthEngine() core.AuthEngine

	// CheckPermission 检查权限
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)

	// CheckRole 检查角色
	CheckRole(ctx context.Context, userID, role string) (bool, error)

	// GetLoginInfo 获取登录信息
	GetLoginInfo(ctx context.Context, token string) (*core.LoginInfo, error)
}

// NewGSTokenWebAdapter 创建 GSToken Web 适配器
func NewGSTokenWebAdapter(gsToken GSTokenInterface) *GSTokenWebAdapter {
	return &GSTokenWebAdapter{
		gsToken: gsToken,
	}
}

// Verify 验证 Token
func (a *GSTokenWebAdapter) Verify(ctx context.Context, token string) (*core.UserInfo, error) {
	return a.gsToken.GetAuthEngine().Verify(ctx, token)
}

// CheckPermission 检查权限
func (a *GSTokenWebAdapter) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	return a.gsToken.CheckPermission(ctx, userID, permission)
}

// CheckRole 检查角色
func (a *GSTokenWebAdapter) CheckRole(ctx context.Context, userID, role string) (bool, error) {
	return a.gsToken.CheckRole(ctx, userID, role)
}

// GetLoginInfo 获取登录信息
func (a *GSTokenWebAdapter) GetLoginInfo(ctx context.Context, token string) (*core.LoginInfo, error) {
	return a.gsToken.GetLoginInfo(ctx, token)
}
