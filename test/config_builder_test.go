package test

import (
	"context"
	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
	"testing"
	"time"
)

func TestConfigBuilder(t *testing.T) {
	ctx := context.Background()

	t.Run("使用配置构建器设置角色提供者", func(t *testing.T) {
		// 创建配置，使用构建器模式
		cfg := config.NewBuilder().
			WithTokenExpire(24 * time.Hour).
			WithLoginMode(core.MultiLogin).
			WithMemoryStorage().
			Build()

		// 创建GSToken实例
		gs := gstoken.New(cfg)

		// 创建角色提供者并设置
		roleProvider := newTestUserRoleProvider(gs.GetStorage())
		gs.GetPermissionService().SetUserRoleProvider(roleProvider)

		// 设置测试数据
		adminRole := &core.Role{
			ID:          "admin",
			Name:        "管理员",
			Permissions: []string{"user:read", "user:write", "user:delete"},
		}

		err := setRoleInStorage(ctx, gs.GetStorage(), adminRole)
		if err != nil {
			t.Fatalf("设置角色失败: %v", err)
		}

		err = setUserRoleInStorage(ctx, gs.GetStorage(), "test_user", []string{"admin"})
		if err != nil {
			t.Fatalf("分配角色失败: %v", err)
		}

		// 测试权限检查
		hasPermission, err := gs.CheckPermission(ctx, "test_user", "user:delete")
		if err != nil {
			t.Fatalf("权限检查失败: %v", err)
		}

		if !hasPermission {
			t.Error("用户应该有删除权限")
		}

		// 测试角色检查
		hasRole, err := gs.CheckRole(ctx, "test_user", "admin")
		if err != nil {
			t.Fatalf("角色检查失败: %v", err)
		}

		if !hasRole {
			t.Error("用户应该有管理员角色")
		}

		t.Log("配置构建器测试通过")
	})

	t.Run("链式配置测试", func(t *testing.T) {
		// 测试链式配置的各种选项
		cfg := config.NewBuilder().
			WithTokenExpire(12*time.Hour).
			WithRefreshExpire(7*24*time.Hour).
			WithTokenStyle(core.StyleUUIDSimple).
			WithLoginMode(core.SingleLogin).
			WithAutoRenew(true).
			WithRememberDays(30).
			WithRedisStorage("localhost:6379", "", 0).
			Build()

		// 验证配置是否正确设置
		if cfg.TokenExpire != 12*time.Hour {
			t.Errorf("TokenExpire配置错误: 期望 %v, 实际 %v", 12*time.Hour, cfg.TokenExpire)
		}

		if cfg.RefreshExpire != 7*24*time.Hour {
			t.Errorf("RefreshExpire配置错误: 期望 %v, 实际 %v", 7*24*time.Hour, cfg.RefreshExpire)
		}

		if cfg.TokenStyle != core.StyleUUIDSimple {
			t.Errorf("TokenStyle配置错误: 期望 %v, 实际 %v", core.StyleUUIDSimple, cfg.TokenStyle)
		}

		if cfg.LoginMode != core.SingleLogin {
			t.Errorf("LoginMode配置错误: 期望 %v, 实际 %v", core.SingleLogin, cfg.LoginMode)
		}

		if !cfg.AutoRenew {
			t.Error("AutoRenew应该为true")
		}

		if cfg.RememberDays != 30 {
			t.Errorf("RememberDays配置错误: 期望 30, 实际 %d", cfg.RememberDays)
		}

		if cfg.Storage.Type != "redis" {
			t.Errorf("Storage类型配置错误: 期望 redis, 实际 %s", cfg.Storage.Type)
		}

		t.Log("链式配置测试通过")
	})
}

// 自定义角色提供者用于测试
type TestRoleProvider struct {
	roles map[string][]core.Role
}

func NewTestRoleProvider() *TestRoleProvider {
	return &TestRoleProvider{
		roles: make(map[string][]core.Role),
	}
}

func (p *TestRoleProvider) SetUserRoles(userID string, roles []core.Role) {
	p.roles[userID] = roles
}

func (p *TestRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
	if roles, exists := p.roles[userID]; exists {
		return roles, nil
	}
	return []core.Role{}, nil
}

func TestCustomRoleProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("自定义角色提供者测试", func(t *testing.T) {
		// 创建自定义角色提供者
		roleProvider := NewTestRoleProvider()

		// 设置测试角色数据
		roleProvider.SetUserRoles("admin_user", []core.Role{
			{
				ID:          "admin",
				Name:        "管理员",
				Permissions: []string{"user:create", "user:read", "user:update", "user:delete"},
			},
		})

		roleProvider.SetUserRoles("normal_user", []core.Role{
			{
				ID:          "user",
				Name:        "普通用户",
				Permissions: []string{"user:read"},
			},
		})

		// 使用配置构建器设置角色提供者
		cfg := config.NewBuilder().
			WithUserRoleProvider(roleProvider).
			WithTokenExpire(24 * time.Hour).
			WithLoginMode(core.MultiLogin).
			Build()

		// 创建GSToken实例，角色提供者会自动配置
		gs := gstoken.New(cfg)

		// 测试管理员权限
		hasAdminPermission, err := gs.CheckPermission(ctx, "admin_user", "user:delete")
		if err != nil {
			t.Fatalf("检查管理员权限失败: %v", err)
		}
		if !hasAdminPermission {
			t.Error("管理员应该有删除权限")
		}

		// 测试普通用户权限
		hasUserPermission, err := gs.CheckPermission(ctx, "normal_user", "user:delete")
		if err != nil {
			t.Fatalf("检查普通用户权限失败: %v", err)
		}
		if hasUserPermission {
			t.Error("普通用户不应该有删除权限")
		}

		// 测试角色检查
		isAdmin, err := gs.CheckRole(ctx, "admin_user", "admin")
		if err != nil {
			t.Fatalf("检查管理员角色失败: %v", err)
		}
		if !isAdmin {
			t.Error("admin_user应该有管理员角色")
		}

		isUser, err := gs.CheckRole(ctx, "normal_user", "user")
		if err != nil {
			t.Fatalf("检查用户角色失败: %v", err)
		}
		if !isUser {
			t.Error("normal_user应该有用户角色")
		}

		t.Log("自定义角色提供者测试通过")
	})
}
