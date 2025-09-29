package test

import (
	"context"
	"encoding/json"
	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
	"github.com/luckxgo/gstoken/web"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestUserRoleProvider 测试用的用户角色提供者
type TestUserRoleProvider struct {
	users map[string]*core.UserInfo
	roles map[string][]string
	perms map[string][]string
}

func NewTestUserRoleProvider() *TestUserRoleProvider {
	return &TestUserRoleProvider{
		users: make(map[string]*core.UserInfo),
		roles: make(map[string][]string),
		perms: make(map[string][]string),
	}
}

func (p *TestUserRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
	if roleNames, exists := p.roles[userID]; exists {
		roles := make([]core.Role, len(roleNames))
		for i, roleName := range roleNames {
			// 获取角色权限
			permissions, _ := p.GetRolePermissions(ctx, roleName)
			roles[i] = core.Role{
				ID:          roleName,
				Name:        roleName,
				Permissions: permissions,
			}
		}
		return roles, nil
	}
	return []core.Role{}, nil
}

func (p *TestUserRoleProvider) GetRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	if perms, exists := p.perms[roleID]; exists {
		return perms, nil
	}
	return []string{}, nil
}

func (p *TestUserRoleProvider) AddUser(userID string, roles []string) {
	p.users[userID] = &core.UserInfo{
		ID:    userID,
		Roles: roles,
	}
	p.roles[userID] = roles
}

func (p *TestUserRoleProvider) AddRolePermissions(roleID string, permissions []string) {
	p.perms[roleID] = permissions
}

// 设置测试环境
func setupTestGSToken() (*gstoken.GSToken, *TestUserRoleProvider) {
	// 创建测试用的用户角色提供者
	roleProvider := NewTestUserRoleProvider()

	// 添加测试数据
	roleProvider.AddUser("user1", []string{"user"})
	roleProvider.AddUser("admin1", []string{"admin", "user"})
	roleProvider.AddRolePermissions("user", []string{"user:read", "user:update", "settings:read"})
	roleProvider.AddRolePermissions("admin", []string{"user:read", "user:update", "user:delete", "admin:read", "settings:read"})

	// 创建配置
	cfg := config.NewBuilder().
		WithMemoryStorage().
		WithUserRoleProvider(roleProvider).
		Build()

	// 创建 GSToken 实例
	gs := gstoken.New(cfg)

	return gs, roleProvider
}

// 测试 Gin 中间件集成
func TestGinMiddlewareIntegration(t *testing.T) {
	gs, _ := setupTestGSToken()

	// 创建 Gin 应用
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 创建认证中间件
	gsAdapter := web.NewGSTokenWebAdapter(gs)
	authMiddleware := web.NewGinAuthMiddleware(gsAdapter, nil)

	// 设置路由
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public"})
	})

	auth := r.Group("/auth")
	auth.Use(authMiddleware.RequireAuth())
	{
		auth.GET("/profile", func(c *gin.Context) {
			userID, _ := web.Gin.GetUserID(c)
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})
	}

	admin := r.Group("/admin")
	admin.Use(authMiddleware.RequireRole("admin"))
	{
		admin.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin users"})
		})
	}

	r.GET("/settings", authMiddleware.RequirePermission("user:update"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "settings"})
	})

	// 测试用例
	tests := []struct {
		name           string
		method         string
		path           string
		token          string
		expectedStatus int
		setupToken     func() string
	}{
		{
			name:           "公开接口无需认证",
			method:         "GET",
			path:           "/public",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "认证接口无Token返回401",
			method:         "GET",
			path:           "/auth/profile",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "认证接口有效Token返回200",
			method:         "GET",
			path:           "/auth/profile",
			expectedStatus: http.StatusOK,
			setupToken: func() string {
				resp, _ := gs.Login(context.Background(), &core.LoginRequest{
					UserID: "user1",
				})
				return resp.Token
			},
		},
		{
			name:           "管理员接口普通用户返回403",
			method:         "GET",
			path:           "/admin/users",
			expectedStatus: http.StatusForbidden,
			setupToken: func() string {
				resp, _ := gs.Login(context.Background(), &core.LoginRequest{
					UserID: "user1",
				})
				return resp.Token
			},
		},
		{
			name:           "管理员接口管理员用户返回200",
			method:         "GET",
			path:           "/admin/users",
			expectedStatus: http.StatusOK,
			setupToken: func() string {
				resp, _ := gs.Login(context.Background(), &core.LoginRequest{
					UserID: "admin1",
				})
				return resp.Token
			},
		},
		{
			name:           "权限接口有权限用户返回200",
			method:         "GET",
			path:           "/settings",
			expectedStatus: http.StatusOK,
			setupToken: func() string {
				resp, _ := gs.Login(context.Background(), &core.LoginRequest{
					UserID: "user1",
				})
				return resp.Token
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			if tt.setupToken != nil {
				token := tt.setupToken()
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("期望状态码 %d，实际得到 %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// 测试方法式鉴权装饰器
func TestMethodAuthDecorator(t *testing.T) {
	gs, _ := setupTestGSToken()

	// 创建装饰器
	gsAdapter := web.NewGSTokenWebAdapter(gs)
	decorator := web.NewAuthDecorator(gsAdapter, nil)

	// 测试业务方法
	getUserInfo := func(ctx context.Context, token string, userID string) (*core.UserInfo, error) {
		authCtx := ctx.(*web.AuthContext)
		return &core.UserInfo{
			ID:       authCtx.UserID,
			Username: "test_user",
		}, nil
	}

	deleteUser := func(ctx context.Context, token string, userID string) error {
		return nil
	}

	// 装饰方法
	decoratedGetUserInfo := decorator.RequireAuth(getUserInfo).(func(context.Context, string, string) (*core.UserInfo, error))
	decoratedDeleteUser := decorator.RequireRole("admin")(deleteUser).(func(context.Context, string, string) error)

	// 登录获取Token
	loginResp, err := gs.Login(context.Background(), &core.LoginRequest{
		UserID: "user1",
	})
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	adminLoginResp, err := gs.Login(context.Background(), &core.LoginRequest{
		UserID: "admin1",
	})
	if err != nil {
		t.Fatalf("管理员登录失败: %v", err)
	}

	// 测试认证装饰器
	t.Run("认证装饰器-有效Token", func(t *testing.T) {
		userInfo, err := decoratedGetUserInfo(context.Background(), loginResp.Token, "user1")
		if err != nil {
			t.Errorf("期望成功，但得到错误: %v", err)
		}
		if userInfo.ID != "user1" {
			t.Errorf("期望用户ID为user1，实际得到 %s", userInfo.ID)
		}
	})

	t.Run("认证装饰器-无效Token", func(t *testing.T) {
		_, err := decoratedGetUserInfo(context.Background(), "invalid_token", "user1")
		if err == nil {
			t.Error("期望得到错误，但调用成功")
		}
	})

	t.Run("角色装饰器-普通用户", func(t *testing.T) {
		err := decoratedDeleteUser(context.Background(), loginResp.Token, "user2")
		if err == nil {
			t.Error("期望得到权限错误，但调用成功")
		}
	})

	t.Run("角色装饰器-管理员用户", func(t *testing.T) {
		err := decoratedDeleteUser(context.Background(), adminLoginResp.Token, "user2")
		if err != nil {
			t.Errorf("期望成功，但得到错误: %v", err)
		}
	})
}

// 测试自定义配置
func TestCustomAuthConfig(t *testing.T) {
	gs, _ := setupTestGSToken()

	// 自定义配置
	config := &web.AuthConfig{
		TokenHeader: "X-Auth-Token",
		TokenQuery:  "access_token",
		TokenPrefix: "",
		SkipPaths:   []string{"/skip"},
		UnauthorizedHandler: func(c web.WebContext, err error) {
			c.AbortWithJSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    "CUSTOM_UNAUTHORIZED",
				"message": "自定义未授权错误",
			})
		},
	}

	// 创建 Gin 应用
	gin.SetMode(gin.TestMode)
	r := gin.New()

	gsAdapter := web.NewGSTokenWebAdapter(gs)
	authMiddleware := web.NewGinAuthMiddleware(gsAdapter, config)

	r.GET("/skip", authMiddleware.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "skipped"})
	})

	r.GET("/auth", authMiddleware.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "authenticated"})
	})

	// 测试跳过路径
	t.Run("跳过认证路径", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/skip", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 200，实际得到 %d", w.Code)
		}
	})

	// 测试自定义Token头
	t.Run("自定义Token头", func(t *testing.T) {
		loginResp, _ := gs.Login(context.Background(), &core.LoginRequest{
			UserID: "user1",
		})

		req := httptest.NewRequest("GET", "/auth", nil)
		req.Header.Set("X-Auth-Token", loginResp.Token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 200，实际得到 %d", w.Code)
		}
	})

	// 测试查询参数Token
	t.Run("查询参数Token", func(t *testing.T) {
		loginResp, _ := gs.Login(context.Background(), &core.LoginRequest{
			UserID: "user1",
		})

		req := httptest.NewRequest("GET", "/auth?access_token="+loginResp.Token, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 200，实际得到 %d", w.Code)
		}
	})

	// 测试自定义错误处理
	t.Run("自定义错误处理", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("期望状态码 401，实际得到 %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["code"] != "CUSTOM_UNAUTHORIZED" {
			t.Errorf("期望自定义错误码，实际得到 %v", response["code"])
		}
	})
}

// 测试可选认证
func TestOptionalAuth(t *testing.T) {
	gs, _ := setupTestGSToken()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	gsAdapter := web.NewGSTokenWebAdapter(gs)
	authMiddleware := web.NewGinAuthMiddleware(gsAdapter, nil)

	r.GET("/content", authMiddleware.OptionalAuth(), func(c *gin.Context) {
		userID, isLoggedIn := web.Gin.GetUserID(c)

		if isLoggedIn {
			c.JSON(http.StatusOK, gin.H{
				"message": "已登录用户",
				"user_id": userID,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "游客用户",
			})
		}
	})

	// 测试未登录访问
	t.Run("游客访问", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/content", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 200，实际得到 %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["message"] != "游客用户" {
			t.Errorf("期望游客消息，实际得到 %v", response["message"])
		}
	})

	// 测试已登录访问
	t.Run("已登录用户访问", func(t *testing.T) {
		loginResp, _ := gs.Login(context.Background(), &core.LoginRequest{
			UserID: "user1",
		})

		req := httptest.NewRequest("GET", "/content", nil)
		req.Header.Set("Authorization", "Bearer "+loginResp.Token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 200，实际得到 %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["message"] != "已登录用户" {
			t.Errorf("期望已登录消息，实际得到 %v", response["message"])
		}

		if response["user_id"] != "user1" {
			t.Errorf("期望用户ID为user1，实际得到 %v", response["user_id"])
		}
	})
}
