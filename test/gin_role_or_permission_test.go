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

type rp struct {
	users map[string][]string
	perms map[string][]string
}

func newRP() *rp {
	return &rp{
		users: make(map[string][]string),
		perms: make(map[string][]string),
	}
}

func (p *rp) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
	roleNames := p.users[userID]
	roles := make([]core.Role, 0, len(roleNames))
	for _, rn := range roleNames {
		roles = append(roles, core.Role{
			ID:          rn,
			Name:        rn,
			Permissions: p.perms[rn],
		})
	}
	return roles, nil
}

func (p *rp) GetRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	return p.perms[roleID], nil
}

func doReq(r *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set(web.HeaderAuthorization, web.BearerPrefix+token)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestGinRequireRoleOrPermission(t *testing.T) {
	// 构造角色/权限：
	// - userperm 角色具备 settings:read 权限
	// - admin 角色不需要 settings:read，但拥有 admin 身份
	// - guest 角色不具备任意权限
	provider := newRP()
	provider.users["hasPermUser"] = []string{"userperm"}
	provider.users["adminUser"] = []string{"admin"}
	provider.users["guestUser"] = []string{"guest"}

	provider.perms["userperm"] = []string{"settings:read"}
	provider.perms["admin"] = []string{"user:delete"} // 用于区分身份
	provider.perms["guest"] = []string{}

	// 初始化 GSToken
	cfg := config.NewBuilder().
		WithMemoryStorage().
		WithUserRoleProvider(provider).
		Build()
	gs := gstoken.New(cfg)

	// 生成三类用户的 Token
	respHasPerm, err := gs.Login(context.Background(), &core.LoginRequest{UserID: "hasPermUser"})
	if err != nil {
		t.Fatalf("login hasPermUser failed: %v", err)
	}
	respAdmin, err := gs.Login(context.Background(), &core.LoginRequest{UserID: "adminUser"})
	if err != nil {
		t.Fatalf("login adminUser failed: %v", err)
	}
	respGuest, err := gs.Login(context.Background(), &core.LoginRequest{UserID: "guestUser"})
	if err != nil {
		t.Fatalf("login guestUser failed: %v", err)
	}

	// Gin 路由与中间件
	gin.SetMode(gin.TestMode)
	r := gin.New()

	auth := web.NewGinAuthMiddleware(web.NewGSTokenWebAdapter(gs), nil)

	// 任意满足角色或权限即放行：
	// 角色 = admin，权限 = settings:read
	r.GET("/mixed/any", auth.RequireRoleOrPermission([]string{"admin"}, []string{"settings:read"}), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// 1) 具备 settings:read 的用户 => 200
	w := doReq(r, "GET", "/mixed/any", respHasPerm.Token)
	if w.Code != http.StatusOK {
		t.Fatalf("hasPermUser expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// 2) 具备 admin 角色的用户 => 200
	w = doReq(r, "GET", "/mixed/any", respAdmin.Token)
	if w.Code != http.StatusOK {
		t.Fatalf("adminUser expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// 3) guest 用户既无角色也无该权限 => 403
	w = doReq(r, "GET", "/mixed/any", respGuest.Token)
	if w.Code != http.StatusForbidden {
		t.Fatalf("guestUser expected 403, got %d body=%s", w.Code, w.Body.String())
	}

	// 4) 未携带 Token => 401
	w = doReq(r, "GET", "/mixed/any", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("no token expected 401, got %d body=%s", w.Code, w.Body.String())
	}
}
