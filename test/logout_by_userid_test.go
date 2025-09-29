package test

import (
	"context"
	"testing"
	"time"

	"gstoken"
	"gstoken/config"
	"gstoken/core"
)

func TestLogoutByUserID(t *testing.T) {
	// 创建GSToken实例
	gs := gstoken.New(config.DefaultConfig())

	ctx := context.Background()
	userID := "test_user_logout"

	t.Run("根据UserID登出所有会话", func(t *testing.T) {
		// 用户在多个设备上登录
		loginReq1 := &core.LoginRequest{
			UserID: userID,
			Device: "mobile",
			IP:     "192.168.1.100",
			Extra: map[string]interface{}{
				"platform": "iOS",
			},
		}

		loginReq2 := &core.LoginRequest{
			UserID: userID,
			Device: "web",
			IP:     "192.168.1.101",
			Extra: map[string]interface{}{
				"platform": "Chrome",
			},
		}

		loginReq3 := &core.LoginRequest{
			UserID: userID,
			Device: "desktop",
			IP:     "192.168.1.102",
			Extra: map[string]interface{}{
				"platform": "Windows",
			},
		}

		// 执行登录
		resp1, err := gs.Login(ctx, loginReq1)
		if err != nil {
			t.Fatalf("第一次登录失败: %v", err)
		}

		resp2, err := gs.Login(ctx, loginReq2)
		if err != nil {
			t.Fatalf("第二次登录失败: %v", err)
		}

		resp3, err := gs.Login(ctx, loginReq3)
		if err != nil {
			t.Fatalf("第三次登录失败: %v", err)
		}

		t.Logf("用户在3个设备上登录成功:")
		t.Logf("  Mobile Token: %s", resp1.Token)
		t.Logf("  Web Token: %s", resp2.Token)
		t.Logf("  Desktop Token: %s", resp3.Token)

		// 验证所有Token都有效
		_, err = gs.GetLoginInfo(ctx, resp1.Token)
		if err != nil {
			t.Fatalf("Mobile Token验证失败: %v", err)
		}

		_, err = gs.GetLoginInfo(ctx, resp2.Token)
		if err != nil {
			t.Fatalf("Web Token验证失败: %v", err)
		}

		_, err = gs.GetLoginInfo(ctx, resp3.Token)
		if err != nil {
			t.Fatalf("Desktop Token验证失败: %v", err)
		}

		t.Logf("所有Token验证通过")

		// 根据UserID登出所有会话
		err = gs.LogoutByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("根据UserID登出失败: %v", err)
		}

		t.Logf("根据UserID登出成功")

		// 等待一小段时间确保删除操作完成
		time.Sleep(100 * time.Millisecond)

		// 验证所有Token都已失效
		_, err = gs.GetLoginInfo(ctx, resp1.Token)
		if err == nil {
			t.Errorf("Mobile Token应该已失效，但仍然有效")
		}

		_, err = gs.GetLoginInfo(ctx, resp2.Token)
		if err == nil {
			t.Errorf("Web Token应该已失效，但仍然有效")
		}

		_, err = gs.GetLoginInfo(ctx, resp3.Token)
		if err == nil {
			t.Errorf("Desktop Token应该已失效，但仍然有效")
		}

		t.Logf("所有Token已成功失效")
	})

	t.Run("验证其他用户不受影响", func(t *testing.T) {
		// 创建另一个用户的登录
		otherUserID := "other_user"
		otherLoginReq := &core.LoginRequest{
			UserID: otherUserID,
			Device: "mobile",
			IP:     "192.168.1.200",
		}

		otherResp, err := gs.Login(ctx, otherLoginReq)
		if err != nil {
			t.Fatalf("其他用户登录失败: %v", err)
		}

		// 再次根据原用户ID登出
		err = gs.LogoutByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("根据UserID登出失败: %v", err)
		}

		// 验证其他用户的Token仍然有效
		_, err = gs.GetLoginInfo(ctx, otherResp.Token)
		if err != nil {
			t.Errorf("其他用户的Token不应该受影响，但已失效: %v", err)
		}

		t.Logf("其他用户Token未受影响，验证通过")

		// 清理其他用户的会话
		gs.Logout(ctx, otherResp.Token)
	})
}
