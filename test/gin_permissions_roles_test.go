package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/luckxgo/gstoken/core"
	"github.com/luckxgo/gstoken/web"
)

// 本用例针对 Gin 框架下权限与角色中间件的行为进行验证。
// 复用已有的适配器：web.NewGSTokenWebAdapter + web.NewGinAuthMiddleware
// 并依赖 test/web_integration_test.go 中的 setupTestGSToken 初始化角色/权限提供者。

// helper: 统一发起请求
func doReq(r *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set(web.HeaderAuthorization, web.BearerPrefix+token)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestGinPermissionMiddlewares(t *testing.T) {
	gs, _ := setupTestGSToken()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	gsAdapter := web.NewGSTokenWebAdapter(gs)
	auth := web.NewGinAuthMiddleware(gsAdapter, nil)

	// 路由：分别验证 RequirePermission / RequireAnyPermission / RequireAllPermissions
	// 使用现有权限：user:read、user:update、settings:read
	r.GET("/perm/read", auth.RequirePermission("user:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/perm/any", auth.RequireAnyPermission("user:delete", "settings:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/perm/all", auth.RequireAllPermissions("user:read", "settings:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// 登录两个用户：普通用户 user1（角色 user），管理员 admin1（角色 admin,user）
	respUser, err := gs.Login(context.Background(), &core.LoginRequest{UserID: "user1"})
	if err != nil {
		t.Fatalf("login user1 failed: %v", err)
	}
	respAdmin, err := gs.Login(context.Background(), &core.LoginRequest{UserID: "admin1"})
	if err != nil {
		t.Fatalf("login admin1 failed: %v", err)
	}

	// 1) RequirePermission(user:read)：user1 有 => 200
	w := doReq(r, "GET", "/perm/read", respUser.Token)
	if w.Code != http.StatusOK {
		t.Fatalf("RequirePermission(user:read) expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// 2) RequireAnyPermission(user:delete 或 settings:read)：
	// 管理员具备 user:delete 与 settings:read => 200
	w = doReq(r, "GET", "/perm/any", respAdmin.Token)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireAnyPermission expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// 3) RequireAllPermissions(user:read, settings:read)：
	// user1 与 admin1 都具备，任选一个验证 => 200
	w = doReq(r, "GET", "/perm/all", respUser.Token)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireAllPermissions expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// 4) 未携带 token => 401
	w = doReq(r, "GET", "/perm/read", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("RequirePermission expected 401 without token, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestGinRoleMiddlewares(t *testing.T) {
	gs, _ := setupTestGSToken()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	gsAdapter := web.NewGSTokenWebAdapter(gs)
	auth := web.NewGinAuthMiddleware(gsAdapter, nil)

	// 路由：分别验证 RequireRole / RequireAnyRole / RequireAllRoles
	r.GET("/role/admin", auth.RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/role/any", auth.RequireAnyRole("admin", "user"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/role/all", auth.RequireAllRoles("admin", "user"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// 登录两个用户：普通用户 user1（角色 user），管理员 admin1（角色 admin,user）
	respUser, err := gs.Login(context.Background(), &core.LoginRequest{UserID: "user1"})
	if err != nil {
		t.Fatalf("login user1 failed: %v", err)
	}
	respAdmin, err := gs.Login(context.Background(), &core.LoginRequest{UserID: "admin1"})
	if err != nil {
		t.Fatalf("login admin1 failed: %v", err)
	}

	// 1) RequireRole(admin)：普通用户 => 403；管理员 => 200
	w := doReq(r, "GET", "/role/admin", respUser.Token)
	if w.Code != http.StatusForbidden {
		t.Fatalf("RequireRole(admin) expected 403 for user1, got %d body=%s", w.Code, w.Body.String())
	}
	w = doReq(r, "GET", "/role/admin", respAdmin.Token)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireRole(admin) expected 200 for admin1, got %d body=%s", w.Code, w.Body.String())
	}

	// 2) RequireAnyRole(admin,user)：管理员与普通用户都满足任一 => 200（验证普通用户）
	w = doReq(r, "GET", "/role/any", respUser.Token)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireAnyRole expected 200 for user1, got %d body=%s", w.Code, w.Body.String())
	}

	// 3) RequireAllRoles(admin,user)：管理员（具备两者）=> 200；普通用户不具备 => 403
	w = doReq(r, "GET", "/role/all", respAdmin.Token)
	if w.Code != http.StatusOK {
		t.Fatalf("RequireAllRoles expected 200 for admin1, got %d body=%s", w.Code, w.Body.String())
	}
	w = doReq(r, "GET", "/role/all", respUser.Token)
	if w.Code != http.StatusForbidden {
		t.Fatalf("RequireAllRoles expected 403 for user1, got %d body=%s", w.Code, w.Body.String())
	}

	// 4) 未携带 token => 401
	w = doReq(r, "GET", "/role/admin", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("RequireRole expected 401 without token, got %d body=%s", w.Code, w.Body.String())
	}
}

//// 已移除自定义辅助函数，直接使用 gs.Login 生成 token，避免依赖不存在的类型或方法。
