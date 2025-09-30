package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
)

// setRoleInStorage 直接在存储中设置角色数据（模拟业务系统创建的角色）
func setRoleInStorage(ctx context.Context, storage core.Storage, role *core.Role) error {
	roleKey := fmt.Sprintf("gstoken:role:%s", role.ID)
	// 直接存储 role 对象，让 storage.Set 内部进行 JSON 序列化
	return storage.Set(ctx, roleKey, role, 0) // 角色信息不过期
}

// setUserRoleInStorage 直接在存储中设置用户角色映射（模拟业务系统分配角色）
func setUserRoleInStorage(ctx context.Context, storage core.Storage, userID string, roleIDs []string) error {
	userRoleKey := fmt.Sprintf("gstoken:user_role:%s", userID)
	// 直接存储 roleIDs，让 storage.Set 内部进行 JSON 序列化
	return storage.Set(ctx, userRoleKey, roleIDs, 0) // 用户角色映射不过期
}

func TestAuthEngine(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TokenExpire = 24 * time.Hour
	cfg.LoginMode = core.MultiLogin

	gs := gstoken.New(cfg)
	engine := gs.GetAuthEngine()

	ctx := context.Background()

	t.Run("用户登录测试", func(t *testing.T) {
		req := &core.LoginRequest{
			UserID: "test_user_001",
			Device: "web",
			IP:     "192.168.1.100",
			Extra: map[string]interface{}{
				"client_version": "1.0.0",
				"platform":       "web",
			},
		}

		resp, err := engine.Login(ctx, req)
		if err != nil {
			t.Fatalf("登录失败: %v", err)
		}

		if resp.Token == "" {
			t.Error("Token不能为空")
		}

		if resp.UserInfo == nil {
			t.Error("用户信息不能为空")
		}

		if resp.UserInfo.ID != req.UserID {
			t.Errorf("用户ID不匹配: 期望 %s, 实际 %s", req.UserID, resp.UserInfo.ID)
		}

		t.Logf("登录成功: Token=%s, ExpireTime=%v", resp.Token, resp.ExpireTime)

		// 验证Token
		userInfo, err := engine.Verify(ctx, resp.Token)
		if err != nil {
			t.Fatalf("Token验证失败: %v", err)
		}

		if userInfo.ID != req.UserID {
			t.Errorf("验证后用户ID不匹配: 期望 %s, 实际 %s", req.UserID, userInfo.ID)
		}

		t.Logf("Token验证成功: UserID=%s", userInfo.ID)

		// 登出
		err = engine.Logout(ctx, resp.Token)
		if err != nil {
			t.Fatalf("登出失败: %v", err)
		}

		// 验证登出后Token无效
		_, err = engine.Verify(ctx, resp.Token)
		if err == nil {
			t.Error("登出后Token应该无效")
		}

		t.Log("登出成功，Token已失效")
	})
}

func TestLoginModes(t *testing.T) {
	ctx := context.Background()

	t.Run("单端登录模式", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.LoginMode = core.SingleLogin
		cfg.TokenExpire = 24 * time.Hour

		gs := gstoken.New(cfg)
		engine := gs.GetAuthEngine()

		userID := "single_login_user"

		// 第一次登录
		req1 := &core.LoginRequest{
			UserID: userID,
			Device: "web",
			IP:     "192.168.1.100",
		}

		resp1, err := engine.Login(ctx, req1)
		if err != nil {
			t.Fatalf("第一次登录失败: %v", err)
		}

		// 第二次登录（应该踢出第一次登录）
		req2 := &core.LoginRequest{
			UserID: userID,
			Device: "mobile",
			IP:     "192.168.1.101",
		}

		resp2, err := engine.Login(ctx, req2)
		if err != nil {
			t.Fatalf("第二次登录失败: %v", err)
		}

		// 验证第一个Token已失效
		_, err = engine.Verify(ctx, resp1.Token)
		if err == nil {
			t.Error("单端登录模式下，第一个Token应该已失效")
		}

		// 验证第二个Token有效
		_, err = engine.Verify(ctx, resp2.Token)
		if err != nil {
			t.Errorf("第二个Token应该有效: %v", err)
		}

		t.Log("单端登录模式测试通过")
	})

	t.Run("多端登录模式", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.LoginMode = core.MultiLogin
		cfg.TokenExpire = 24 * time.Hour

		gs := gstoken.New(cfg)
		engine := gs.GetAuthEngine()

		userID := "multi_login_user"

		// 第一次登录
		req1 := &core.LoginRequest{
			UserID: userID,
			Device: "web",
			IP:     "192.168.1.100",
		}

		resp1, err := engine.Login(ctx, req1)
		if err != nil {
			t.Fatalf("第一次登录失败: %v", err)
		}

		// 第二次登录
		req2 := &core.LoginRequest{
			UserID: userID,
			Device: "mobile",
			IP:     "192.168.1.101",
		}

		resp2, err := engine.Login(ctx, req2)
		if err != nil {
			t.Fatalf("第二次登录失败: %v", err)
		}

		// 验证两个Token都有效
		_, err = engine.Verify(ctx, resp1.Token)
		if err != nil {
			t.Errorf("多端登录模式下，第一个Token应该有效: %v", err)
		}

		_, err = engine.Verify(ctx, resp2.Token)
		if err != nil {
			t.Errorf("多端登录模式下，第二个Token应该有效: %v", err)
		}

		t.Log("多端登录模式测试通过")
	})

	t.Run("同端互斥登录模式", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.LoginMode = core.MutexLogin
		cfg.TokenExpire = 24 * time.Hour

		gs := gstoken.New(cfg)
		engine := gs.GetAuthEngine()

		userID := "mutex_login_user"

		// 在web端第一次登录
		req1 := &core.LoginRequest{
			UserID: userID,
			Device: "web",
			IP:     "192.168.1.100",
		}

		resp1, err := engine.Login(ctx, req1)
		if err != nil {
			t.Fatalf("web端第一次登录失败: %v", err)
		}

		// 在mobile端登录（不同设备，应该都有效）
		req2 := &core.LoginRequest{
			UserID: userID,
			Device: "mobile",
			IP:     "192.168.1.101",
		}

		resp2, err := engine.Login(ctx, req2)
		if err != nil {
			t.Fatalf("mobile端登录失败: %v", err)
		}

		// 在web端第二次登录（同设备，应该踢出第一次登录）
		req3 := &core.LoginRequest{
			UserID: userID,
			Device: "web",
			IP:     "192.168.1.102",
		}

		resp3, err := engine.Login(ctx, req3)
		if err != nil {
			t.Fatalf("web端第二次登录失败: %v", err)
		}

		// 验证web端第一个Token已失效
		_, err = engine.Verify(ctx, resp1.Token)
		if err == nil {
			t.Error("同端互斥模式下，web端第一个Token应该已失效")
		}

		// 验证mobile端Token仍有效
		_, err = engine.Verify(ctx, resp2.Token)
		if err != nil {
			t.Errorf("mobile端Token应该仍有效: %v", err)
		}

		// 验证web端第二个Token有效
		_, err = engine.Verify(ctx, resp3.Token)
		if err != nil {
			t.Errorf("web端第二个Token应该有效: %v", err)
		}

		t.Log("同端互斥登录模式测试通过")
	})
}

func TestPermissionSystem(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	engine := gs.GetAuthEngine()

	ctx := context.Background()
	userID := "permission_test_user"

	t.Run("权限检查测试", func(t *testing.T) {
		// 设置用户角色提供者
		permissionService := engine.GetPermissionService()
		userRoleProvider := newTestUserRoleProvider(gs.GetStorage())
		permissionService.SetUserRoleProvider(userRoleProvider)

		// 创建角色
		adminRole := &core.Role{
			ID:   "admin",
			Name: "管理员",
			Permissions: []string{
				"user:create",
				"user:read",
				"user:update",
				"user:delete",
				"system:config",
			},
		}

		// 直接在存储中设置角色数据（模拟业务系统创建的角色）
		storage := gs.GetStorage()
		err := setRoleInStorage(ctx, storage, adminRole)
		if err != nil {
			t.Fatalf("设置管理员角色失败: %v", err)
		}

		userRole := &core.Role{
			ID:   "user",
			Name: "普通用户",
			Permissions: []string{
				"user:read",
				"profile:update",
			},
		}

		err = setRoleInStorage(ctx, storage, userRole)
		if err != nil {
			t.Fatalf("设置普通用户角色失败: %v", err)
		}

		// 分配角色（模拟业务系统分配角色）
		err = setUserRoleInStorage(ctx, storage, userID, []string{"user"})
		if err != nil {
			t.Fatalf("分配用户角色失败: %v", err)
		}

		// 检查权限
		hasPermission, err := engine.CheckPermission(ctx, userID, "user:read")
		if err != nil {
			t.Fatalf("检查权限失败: %v", err)
		}

		if !hasPermission {
			t.Error("用户应该有 user:read 权限")
		}

		hasPermission, err = engine.CheckPermission(ctx, userID, "user:delete")
		if err != nil {
			t.Fatalf("检查权限失败: %v", err)
		}

		if hasPermission {
			t.Error("用户不应该有 user:delete 权限")
		}

		// 分配管理员角色（模拟业务系统分配角色）
		err = setUserRoleInStorage(ctx, storage, userID, []string{"user", "admin"})
		if err != nil {
			t.Fatalf("分配管理员角色失败: %v", err)
		}

		// 再次检查权限
		hasPermission, err = engine.CheckPermission(ctx, userID, "user:delete")
		if err != nil {
			t.Fatalf("检查权限失败: %v", err)
		}

		if !hasPermission {
			t.Error("分配管理员角色后，用户应该有 user:delete 权限")
		}

		t.Log("权限检查测试通过")
	})
}
