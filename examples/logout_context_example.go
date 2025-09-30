package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
	"github.com/luckxgo/gstoken/web"
)

// 示例用户角色提供者
type ExampleUserRoleProvider struct{}

func (p *ExampleUserRoleProvider) GetUserInfo(ctx context.Context, userID string) (*core.UserInfo, error) {
	return &core.UserInfo{
		ID:       userID,
		Username: "user_" + userID,
		Roles:    []string{"user"},
		Extra:    map[string]interface{}{"department": "IT"},
	}, nil
}

func (p *ExampleUserRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
	return []core.Role{
		{ID: "user", Name: "普通用户", Permissions: []string{"read"}},
		{ID: "member", Name: "会员", Permissions: []string{"read", "write"}},
	}, nil
}

func (p *ExampleUserRoleProvider) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return []string{"read", "write"}, nil
}

func (p *ExampleUserRoleProvider) CheckRole(ctx context.Context, userID, roleID string) (bool, error) {
	return roleID == "user" || roleID == "member", nil
}

func (p *ExampleUserRoleProvider) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	return permission == "read" || permission == "write", nil
}

func main() {
	// 创建配置
	cfg := config.NewBuilder().
		WithMemoryStorage().
		WithUserRoleProvider(&ExampleUserRoleProvider{}).
		WithTokenExpire(time.Hour).
		WithRefreshExpire(time.Hour * 24).
		Build()

	// 创建 GSToken 实例
	gs := gstoken.New(cfg)

	// 创建 Web 适配器
	gsAdapter := web.NewGSTokenWebAdapter(gs)
	authMiddleware := web.NewGinAuthMiddleware(gsAdapter, nil)

	// 创建 Gin 路由
	r := gin.Default()

	// 登录接口
	r.POST("/login", func(c *gin.Context) {
		userID := c.PostForm("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		loginReq := &core.LoginRequest{
			UserID: userID,
			Extra:  map[string]interface{}{"login_time": time.Now()},
		}

		loginResp, err := gs.Login(c.Request.Context(), loginReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "登录成功",
			"token":   loginResp.Token,
		})
	})

	// 需要认证的路由组
	auth := r.Group("/api")
	auth.Use(authMiddleware.RequireAuth())

	// 用户信息接口
	auth.GET("/profile", func(c *gin.Context) {
		userID, _ := web.Helper.GetUserID(c)
		userInfo, _ := web.Helper.GetUserInfo(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id":   userID,
			"user_info": userInfo,
		})
	})

	// 方式1: 传统登出方式 - 需要显式传入 token
	auth.POST("/logout", func(c *gin.Context) {
		token, exists := web.Helper.GetToken(c)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token not found"})
			return
		}

		err := gs.Logout(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
	})

	// 方式2: 新的上下文登出方式 - 自动从上下文获取 token
	auth.POST("/logout-auto", func(c *gin.Context) {
		// 使用 LogoutFromContext，自动从上下文获取 token
		token, _ := web.Helper.GetToken(c)
		ctx := context.WithValue(c.Request.Context(), "token", token)

		err := gs.LogoutFromContext(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "自动登出成功"})
	})

	// 方式3: 使用 GinHelper 的便捷方法
	auth.POST("/logout-helper", func(c *gin.Context) {
		// 使用 GinHelper 的便捷方法，一行代码完成登出
		err := web.Helper.LogoutFromGinContext(c, gs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "便捷登出成功"})
	})

	// 演示不同登出方式的使用
	fmt.Println("=== GSToken 登出方式演示 ===")
	fmt.Println()
	fmt.Println("启动服务器在 :8080")
	fmt.Println()
	fmt.Println("使用步骤：")
	fmt.Println("1. 登录获取 token:")
	fmt.Println("   curl -X POST http://localhost:8080/login -d 'user_id=123'")
	fmt.Println()
	fmt.Println("2. 使用 token 访问受保护的接口:")
	fmt.Println("   curl -H 'Authorization: Bearer YOUR_TOKEN' http://localhost:8080/api/profile")
	fmt.Println()
	fmt.Println("3. 选择一种登出方式:")
	fmt.Println("   方式1 - 传统登出:")
	fmt.Println("   curl -X POST -H 'Authorization: Bearer YOUR_TOKEN' http://localhost:8080/api/logout")
	fmt.Println()
	fmt.Println("   方式2 - 上下文自动登出:")
	fmt.Println("   curl -X POST -H 'Authorization: Bearer YOUR_TOKEN' http://localhost:8080/api/logout-auto")
	fmt.Println()
	fmt.Println("   方式3 - GinHelper 便捷登出:")
	fmt.Println("   curl -X POST -H 'Authorization: Bearer YOUR_TOKEN' http://localhost:8080/api/logout-helper")
	fmt.Println()

	// 启动服务器
	r.Run(":8080")
}