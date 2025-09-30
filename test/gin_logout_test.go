package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
	"github.com/luckxgo/gstoken/web"
)

func TestGinHelperLogout(t *testing.T) {
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

	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	t.Run("从Gin上下文登出", func(t *testing.T) {
		// 创建 Gin 路由
		r := gin.New()

		// 创建登出路由
		r.POST("/logout", func(c *gin.Context) {
			// 模拟中间件设置的 token
			c.Set(web.ContextKeyToken, loginResp.Token)

			// 使用 GinHelper 从上下文登出
			err := web.Helper.LogoutFromGinContext(c, gs)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
		})

		// 创建测试请求
		req, _ := http.NewRequest("POST", "/logout", nil)
		w := httptest.NewRecorder()

		// 执行请求
		r.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 200, 实际: %d", w.Code)
		}

		// 验证 token 已失效
		isLogin := gs.IsLogin(context.Background(), loginResp.Token)
		if isLogin {
			t.Error("期望 token 已失效，但验证仍然通过")
		}
	})

	// 重新登录用于下一个测试
	loginResp, err = gs.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("重新登录失败: %v", err)
	}

	t.Run("Gin上下文中没有token", func(t *testing.T) {
		// 创建 Gin 路由
		r := gin.New()

		// 创建登出路由（不设置 token）
		r.POST("/logout", func(c *gin.Context) {
			// 不设置 token，模拟未认证的情况

			// 使用 GinHelper 从上下文登出
			err := web.Helper.LogoutFromGinContext(c, gs)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
		})

		// 创建测试请求
		req, _ := http.NewRequest("POST", "/logout", nil)
		w := httptest.NewRecorder()

		// 执行请求
		r.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusBadRequest {
			t.Errorf("期望状态码 400, 实际: %d", w.Code)
		}

		// 验证错误信息
		expectedError := "token not found in gin context"
		if !contains(w.Body.String(), expectedError) {
			t.Errorf("期望包含错误信息: %s, 实际响应: %s", expectedError, w.Body.String())
		}
	})
}

// 辅助函数：检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}