package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/core"
	"github.com/luckxgo/gstoken/web"
)

func main() {
	// 创建 GSToken 实例
	gs := gstoken.New(&core.Config{
		TokenExpire:   2 * time.Hour,
		RefreshExpire: 7 * 24 * time.Hour,
		TokenStyle:    core.StyleUUID,
		LoginMode:     core.MultiLogin,
		AutoRenew:     true,
		RememberDays:  7,
		KeyPrefix:     "gstoken",
	})

	// 创建 GSToken Web 适配器
	gsAdapter := web.NewGSTokenWebAdapter(gs)

	// 创建 Gin 认证中间件
	authMiddleware := web.NewGinAuthMiddleware(gsAdapter, web.DefaultAuthConfig())

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 路由
	router := gin.Default()

	// 公共路由 - 无需认证
	public := router.Group("/api/public")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "healthy",
				"message": "GSToken Gin Example is running",
			})
		})

		public.POST("/login", func(c *gin.Context) {
			var loginReq struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.ShouldBindJSON(&loginReq); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid request body",
				})
				return
			}

			// 模拟用户验证
			if loginReq.Username == "testuser" && loginReq.Password == "123456" {
				// 执行登录
				resp, err := gs.Login(context.Background(), &core.LoginRequest{
					UserID: "user-123",
					Device: "web",
					IP:     c.ClientIP(),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Failed to generate token",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"token":         resp.Token,
					"refresh_token": resp.RefreshToken,
					"expire_time":   resp.ExpireTime,
					"user_info":     resp.UserInfo,
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid credentials",
				})
			}
		})
	}

	// 认证路由 - 需要认证
	auth := router.Group("/api/auth")
	auth.Use(authMiddleware.RequireAuth())
	{
		auth.GET("/profile", func(c *gin.Context) {
			// 使用 web.Helper 获取用户信息
			userID := web.Helper.MustGetUserID(c)
			token, _ := web.Helper.GetToken(c)
			userInfo, _ := web.Helper.GetUserInfo(c)

			c.JSON(http.StatusOK, gin.H{
				"user_id":   userID,
				"token":     token,
				"user_info": userInfo,
				"message":   "This is a protected endpoint",
			})
		})

		auth.POST("/logout", func(c *gin.Context) {
			token, hasToken := web.Helper.GetToken(c)
			if hasToken && token != "" {
				gs.Logout(context.Background(), token)
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Logged out successfully",
			})
		})
	}

	// 可选认证路由 - 认证可选
	optional := router.Group("/api/optional")
	optional.Use(authMiddleware.OptionalAuth())
	{
		optional.GET("/content", func(c *gin.Context) {
			userID, hasUserID := web.Helper.GetUserID(c)
			token, _ := web.Helper.GetToken(c)

			response := gin.H{
				"message": "This endpoint works with or without authentication",
			}

			if hasUserID && userID != "" {
				response["authenticated"] = true
				response["user_id"] = userID
				response["token"] = token
			} else {
				response["authenticated"] = false
				response["message"] = "You are accessing this endpoint anonymously"
			}

			c.JSON(http.StatusOK, response)
		})
	}

	// 权限控制路由 - 需要特定权限
	permission := router.Group("/api/permission")
	permission.Use(authMiddleware.RequirePermission("admin:access"))
	{
		permission.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "You have admin access",
			})
		})
	}

	// 角色控制路由 - 需要特定角色
	role := router.Group("/api/role")
	role.Use(authMiddleware.RequireRole("admin"))
	{
		role.GET("/admin-only", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "This is admin-only content",
			})
		})
	}

	fmt.Println("=== GSToken Gin Example ===")
	fmt.Println("Server starting on http://localhost:8081")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /api/public/health")
	fmt.Println("  POST /api/public/login")
	fmt.Println("  GET  /api/auth/profile (requires authentication)")
	fmt.Println("  POST /api/auth/logout (requires authentication)")
	fmt.Println("  GET  /api/optional/content (optional authentication)")
	fmt.Println("  GET  /api/permission/admin (requires 'admin:access' permission)")
	fmt.Println("  GET  /api/role/admin-only (requires 'admin' role)")

	// 启动服务器
	if err := router.Run(":8081"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
