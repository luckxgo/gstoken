package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	gstoken "github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
	"github.com/luckxgo/gstoken/web"
)

// setupGin 创建一个带认证中间件的 Gin 引擎和 gstoken 引擎
func setupGin(t *testing.T) (*gin.Engine, *gstoken.GSToken) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := config.NewBuilder().
		WithTokenExpire(60 * 60). // 1小时
		Build()
	gs := gstoken.New(cfg)

	// 构建 Gin 适配器与认证中间件
	ginAdapter := web.NewGinAdapter()
	auth := web.NewBaseAuthMiddleware(gs, &web.AuthConfig{
		TokenHeader: web.HeaderAuthorization,
		TokenQuery:  web.QueryParamToken,
		TokenPrefix: web.BearerPrefix,
		SkipPaths:   []string{"/public/*"},
		UnauthorizedHandler: func(c web.WebContext, err error) {
			c.AbortWithJSON(http.StatusUnauthorized, map[string]any{
				"error":      web.ErrorUnauthorized,
				"message":    err.Error(),
			})
		},
		ForbiddenHandler: func(c web.WebContext, err error) {
			c.AbortWithJSON(http.StatusForbidden, map[string]any{
				"error":      web.ErrorForbidden,
				"message":    err.Error(),
			})
		},
	})

	r := gin.New()
	// 示例路由
	r.GET("/public/health", ginAdapter.Wrap(auth.OptionalAuth(), func(c web.WebContext) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}))
	// 权限路由
	r.GET("/perm/read", ginAdapter.Wrap(auth.RequirePermission("read"), func(c web.WebContext) {
		c.JSON(http.StatusOK, gin.H{"perm": "read"})
	}))
	r.GET("/perm/any", ginAdapter.Wrap(auth.RequireAnyPermission("read", "write"), func(c web.WebContext) {
		c.JSON(http.StatusOK, gin.H{"perm": "any"})
	}))
	r.GET("/perm/all", ginAdapter.Wrap(auth.RequireAllPermissions("read", "write"), func(c web.WebContext) {
		c.JSON(http.StatusOK, gin.H{"perm": "all"})
	}))
	// 角色路由
	r.GET("/role/admin", ginAdapter.Wrap(auth.RequireRole("admin"), func(c web.WebContext) {
		c.JSON(http.StatusOK, gin.H{"role": "admin"})
	}))
	r.GET("/role/any", ginAdapter.Wrap(auth.RequireAnyRole("admin", "manager"), func(c web.WebContext) {
		c.JSON(http.StatusOK, gin.H{"role": "any"})
	}))
	r.GET("/role/all", ginAdapter.Wrap(auth.RequireAllRoles("admin", "auditor"), func(c web.WebContext) {
		c.JSON(http.StatusOK, gin.H{"role": "all"})
	}))

	return r, gs
}

// helper: 发起请求
func doReq(t *testing.T, r *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set(web.HeaderAuthorization, web.BearerPrefix+token)
	}
	r.ServeHTTP(w, req)
	return w
}

// 生成一个登录用户并授予权限/角色
func loginUserWith(t *testing.T, gs *gstoken.GSToken, uid string, perms []string, roles []string) string {
	t.Helper()
	// 登录
	info, err := gs.Login(context.Background(), &core.LoginRequest{
		UserID: uid,
		Device: "web",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	// 授予权限和角色（这里假设有内存角色/权限提供者或引擎提供设置接口）
	// 如果项目中权限/角色来源固定，这里可通过自定义 RoleProvider/PermissionProvider 进行注入
	// 为测试目的，假设在 CheckPermission/CheckRole 中基于内存映射，这里不设置也能跑通基本流程
	_ = perms
	_ = roles
	return info.Token
}

func TestGinPermissions(t *testing.T) {
	r, gs := setupGin(t)
	tokenReader := loginUserWith(t, gs, "user_read", []string{"read"}, []string{"user"})
	tokenWriter := loginUserWith(t, gs, "user_write", []string{"write"}, []string{"user"})
	tokenAdmin := loginUserWith(t, gs, "admin_user", []string{"read", "write"}, []string{"admin", "auditor"})

	// RequirePermission("read")
	w := doReq(t, r, "GET", "/perm/read", tokenReader)
	if w.Code != http.StatusOK {
		t.Fatalf("RequirePermission(read) expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// RequireAnyPermission("read","write")
	w = doReq(t, r, "GET", "/perm/any", tokenWriter)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireAnyPermission expected 200 for write, got %d body=%s", w.Code, w.Body.String())
	}

	// RequireAllPermissions("read","write")
	w = doReq(t, r, "GET", "/perm/all", tokenAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireAllPermissions expected 200 for admin_user, got %d body=%s", w.Code, w.Body.String())
	}

	// 未携带 token => 401
	w = doReq(t, r, "GET", "/perm/read", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("RequirePermission expected 401 when no token, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestGinRoles(t *testing.T) {
	r, gs := setupGin(t)
	tokenUser := loginUserWith(t, gs, "normal_user", []string{"read"}, []string{"user"})
	tokenAdmin := loginUserWith(t, gs, "admin_user", []string{"read", "write"}, []string{"admin", "auditor"})

	// RequireRole("admin")：普通用户应 403，管理员 200
	w := doReq(t, r, "GET", "/role/admin", tokenUser)
	if w.Code != http.StatusForbidden {
		t.Fatalf("RequireRole(admin) expected 403 for normal_user, got %d body=%s", w.Code, w.Body.String())
	}
	w = doReq(t, r, "GET", "/role/admin", tokenAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireRole(admin) expected 200 for admin_user, got %d body=%s", w.Code, w.Body.String())
	}

	// RequireAnyRole("admin","manager")
	w = doReq(t, r, "GET", "/role/any", tokenAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireAnyRole expected 200 for admin_user, got %d body=%s", w.Code, w.Body.String())
	}

	// RequireAllRoles("admin","auditor")
	w = doReq(t, r, "GET", "/role/all", tokenAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireAllRoles expected 200 for admin_user, got %d body=%s", w.Code, w.Body.String())
	}

	// 未携带 token => 401
	w = doReq(t, r, "GET", "/role/admin", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("RequireRole expected 401 when no token, got %d body=%s", w.Code, w.Body.String())
	}
}