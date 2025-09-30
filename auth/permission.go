package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/luckxgo/gstoken/core"
)

// PermissionService 权限服务默认实现
type PermissionService struct {
	storage          core.Storage
	keyService       *core.KeyService
	userRoleProvider core.UserRoleProvider
}

// NewPermissionService 创建新的权限服务
func NewPermissionService(storage core.Storage, keyService *core.KeyService) core.PermissionService {
	return &PermissionService{
		storage:    storage,
		keyService: keyService,
	}
}

// SetUserRoleProvider 设置用户角色提供者
func (p *PermissionService) SetUserRoleProvider(provider core.UserRoleProvider) {
	p.userRoleProvider = provider
}

// CheckPermission 检查用户权限
func (p *PermissionService) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	if userID == "" {
		return false, errors.New(core.ErrMsgUserIDEmpty)
	}

	if permission == "" {
		return false, errors.New(core.ErrMsgPermissionEmpty)
	}

	if p.userRoleProvider == nil {
		return false, errors.New(core.ErrMsgUserRoleProviderEmpty)
	}

	// 获取用户角色
	roles, err := p.userRoleProvider.GetUserRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", core.ErrMsgGetUserRoles, err)
	}

	// 检查角色权限
	for _, role := range roles {
		for _, perm := range role.Permissions {
			if perm == permission || perm == core.PermissionWildcard {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckRole 检查用户是否拥有指定角色
func (p *PermissionService) CheckRole(ctx context.Context, userID, roleID string) (bool, error) {
	if userID == "" {
		return false, errors.New(core.ErrMsgUserIDEmpty)
	}

	if roleID == "" {
		return false, errors.New(core.ErrMsgRoleIDEmpty)
	}

	if p.userRoleProvider == nil {
		return false, errors.New(core.ErrMsgUserRoleProviderEmpty)
	}

	// 获取用户角色
	roles, err := p.userRoleProvider.GetUserRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", core.ErrMsgGetUserRoles, err)
	}

	// 检查是否拥有指定角色
	for _, role := range roles {
		if role.ID == roleID {
			return true, nil
		}
	}

	return false, nil
}
