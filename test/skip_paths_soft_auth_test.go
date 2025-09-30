package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	gstoken "github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
	"github.com/luckxgo/gstoken/web"
)

// TestSkipPathsSoftAuth 验证当路径命中 SkipPaths 时：
// - 如果请求携带有效 Token，则执行软认证，写入用户信息到上下文，并继续放行
// - 如果请求未携带 Token，则直接放行，不写入用户信息
func TestSkipPathsSoftAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 初始化 GSToken（内存存储，自动续期关闭，便于测试）
	cfg := config.NewBuilder().
		WithMemoryStorage().
		WithAutoRenew(false).
		Build()
	gs := gstoken.New(cfg)
	gsAdapter := web.NewGSTokenWebAdapter(gs)

	// 准备用户并登录，获取 Token
	loginResp, err := gs.Login(context.Background(), &core.LoginRequest{
		UserID: "soft_user",
		Device: "web",
	})
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	token := loginResp.Token

	// 配置 SkipPaths 命中 /public/*，并启用 Bearer 前缀
	authConfig := &web.AuthConfig{
		TokenHeader: web.HeaderAuthorization,
		TokenQuery:  web.QueryParamToken,
		TokenPrefix: web.BearerPrefix,
		SkipPaths:   []string{"/public/*"},
		UnauthorizedHandler: func(c web.WebContext, e error) {
			// 测试场景中不应进入未授权
			c.AbortWithJSON(http.StatusUnauthorized, gin.H{"error": e.Error()})
		},
		ForbiddenHandler: func(c web.WebContext, e error) {
			c.AbortWithJSON(http.StatusForbidden, gin.H{"error": e.Error()})
		},
	}

	authMiddleware := web.NewGinAuthMiddleware(gsAdapter, authConfig)

	// 路由
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/public/info", authMiddleware.RequireAuth(), func(c *gin.Context) {
		// 期望：若带 Token，则应写入用户信息
		userID, hasUID := web.Helper.GetUserID(c)
		_, hasToken := web.Helper.GetToken(c)
		userInfo, hasUserInfo := web.Helper.GetUserInfo(c)

		c.JSON(http.StatusOK, gin.H{
			"has_uid":       hasUID,
			"user_id":       userID,
			"has_token":     hasToken,
			"has_user_info": hasUserInfo,
			"user_info_id": func() string {
				if userInfo != nil {
					return userInfo.ID
				}
				return ""
			}(),
		})
	})

	// 1) 命中 SkipPaths 且携带有效 Token => 软认证生效
	req1 := httptest.NewRequest("GET", "/public/info", nil)
	req1.Header.Set(web.HeaderAuthorization, web.BearerPrefix+token)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d", w1.Code)
	}
	// 简单检查响应内容包含 user_id 与 has_user_info=true
	body1 := w1.Body.String()
	if !(strings.Contains(body1, `"has_uid":true`) && (strings.Contains(body1, `"user_id":"soft_user"`) || strings.Contains(body1, `"user_info_id":"soft_user"`))) {
		t.Fatalf("软认证未写入用户信息: %s", body1)
	}

	// 2) 命中 SkipPaths 且不携带 Token => 放行但不写用户信息
	req2 := httptest.NewRequest("GET", "/public/info", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d", w2.Code)
	}
	body2 := w2.Body.String()
	if strings.Contains(body2, `"has_uid":true`) || strings.Contains(body2, `"has_user_info":true`) {
		t.Fatalf("未携带 token 不应写入用户信息: %s", body2)
	}
}
