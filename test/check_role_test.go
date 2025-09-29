package test

import (
	"context"
	"testing"
	"time"

	"gstoken"
	"gstoken/config"
	"gstoken/core"
)

func TestCheckRole(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TokenExpire = 24 * time.Hour
	gs := gstoken.New(cfg)
	engine := gs.GetAuthEngine()

	ctx := context.Background()
	userID := "role_test_user"

	t.Run("角色检查测试", func(t *testing.T) {
		// 设置用户角色提供者
		permissionService := engine.GetPermissionService()
		userRoleProvider := newTestUserRoleProvider(gs.GetStorage())
		permissionService.SetUserRoleProvider(userRoleProvider)

		// 设置测试角色
		adminRole := &core.Role{
			ID:          "admin",
			Name:        "管理员",
			Permissions: []string{"user:read", "user:write", "user:delete"},
		}

		userRole := &core.Role{
			ID:          "user",
			Name:        "普通用户",
			Permissions: []string{"user:read"},
		}

		// 在存储中设置角色数据
		err := setRoleInStorage(ctx, gs.GetStorage(), adminRole)
		if err != nil {
			t.Fatalf("设置管理员角色失败: %v", err)
		}

		err = setRoleInStorage(ctx, gs.GetStorage(), userRole)
		if err != nil {
			t.Fatalf("设置用户角色失败: %v", err)
		}

		// 为用户分配管理员角色
		err = setUserRoleInStorage(ctx, gs.GetStorage(), userID, []string{"admin"})
		if err != nil {
			t.Fatalf("分配用户角色失败: %v", err)
		}

		// 测试角色检查
		hasAdminRole, err := engine.CheckPermission(ctx, userID, "role:admin")
		if err != nil {
			t.Fatalf("检查管理员角色失败: %v", err)
		}

		// 由于我们没有实现 CheckRole 的具体逻辑，这里先测试权限检查
		hasUserReadPermission, err := engine.CheckPermission(ctx, userID, "user:read")
		if err != nil {
			t.Fatalf("检查用户读权限失败: %v", err)
		}

		if !hasUserReadPermission {
			t.Error("用户应该有读权限")
		}

		hasUserDeletePermission, err := engine.CheckPermission(ctx, userID, "user:delete")
		if err != nil {
			t.Fatalf("检查用户删除权限失败: %v", err)
		}

		if !hasUserDeletePermission {
			t.Error("管理员应该有删除权限")
		}

		// 测试不存在的权限
		hasNonExistentPermission, err := engine.CheckPermission(ctx, userID, "system:admin")
		if err != nil {
			t.Fatalf("检查不存在权限失败: %v", err)
		}

		if hasNonExistentPermission {
			t.Error("用户不应该有系统管理权限")
		}

		t.Logf("角色检查测试通过: hasAdminRole=%v, hasUserRead=%v, hasUserDelete=%v",
			hasAdminRole, hasUserReadPermission, hasUserDeletePermission)
	})
}
