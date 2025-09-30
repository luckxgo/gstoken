package test

import (
	"context"
	"testing"
	"time"

	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
)

func TestLogoutFromContext(t *testing.T) {
	// 创建测试用的角色提供者
	roleProvider := &TestUserRoleProvider{}

	// 创建配置
	cfg := config.NewBuilder().
		WithMemoryStorage().
		WithUserRoleProvider(roleProvider).
		WithTokenExpire(time.Hour).
		WithRefreshExpire(time.Hour * 24).
		Build()

	// 创建 GSToken 实例
	gs := gstoken.New(cfg)

	// 用户登录
	loginReq := &core.LoginRequest{
		UserID: "test-user-123",
		Extra:  map[string]interface{}{"name": "测试用户", "role": "admin"},
	}

	loginResp, err := gs.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("用户登录失败: %v", err)
	}

	// 测试 1: 正常情况 - 从上下文登出
	t.Run("正常登出", func(t *testing.T) {
		// 创建包含 token 的上下文
		ctx := context.WithValue(context.Background(), "token", loginResp.Token)

		// 从上下文登出
		err := gs.LogoutFromContext(ctx)
		if err != nil {
			t.Errorf("从上下文登出失败: %v", err)
		}

		// 验证 token 已失效
		isLogin := gs.IsLogin(context.Background(), loginResp.Token)
		if isLogin {
			t.Error("期望 token 已失效，但验证仍然通过")
		}
	})

	// 重新登录用于后续测试
	loginResp, err = gs.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("重新登录失败: %v", err)
	}

	// 测试 2: 上下文中没有 token
	t.Run("上下文中没有token", func(t *testing.T) {
		ctx := context.Background()

		err := gs.LogoutFromContext(ctx)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}

		expectedMsg := "token not found in context"
		if err.Error() != expectedMsg {
			t.Errorf("期望错误信息: %s, 实际: %s", expectedMsg, err.Error())
		}
	})

	// 测试 3: 上下文中的 token 不是字符串类型
	t.Run("token类型错误", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", 12345)

		err := gs.LogoutFromContext(ctx)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}

		expectedMsg := "token in context is not a string"
		if err.Error() != expectedMsg {
			t.Errorf("期望错误信息: %s, 实际: %s", expectedMsg, err.Error())
		}
	})

	// 测试 4: 上下文中的 token 为空字符串
	t.Run("token为空字符串", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "")

		err := gs.LogoutFromContext(ctx)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}

		expectedMsg := "token in context is empty"
		if err.Error() != expectedMsg {
			t.Errorf("期望错误信息: %s, 实际: %s", expectedMsg, err.Error())
		}
	})

	// 测试 5: 无效的 token
	t.Run("无效token", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "invalid-token")

		err := gs.LogoutFromContext(ctx)
		// 这里应该不会报错，因为登出操作通常不验证 token 的有效性
		// 只是简单地从存储中删除相关数据
		if err != nil {
			t.Logf("无效 token 登出结果: %v", err)
		}
	})
}
