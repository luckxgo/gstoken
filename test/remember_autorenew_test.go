package test

import (
	"context"
	"testing"
	"time"

	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
)

// TestAutoRenewEnabled 验证开启自动续期时，Verify 会更新 LastAccess 并延长有效期
func TestAutoRenewEnabled(t *testing.T) {
	// 短期过期便于验证续期
	cfg := config.NewBuilder().
		WithTokenExpire(2 * time.Second).
		WithRefreshExpire(0).
		WithAutoRenew(true).
		Build()

	gs := gstoken.New(cfg)
	ctx := context.Background()

	resp, err := gs.Login(ctx, &core.LoginRequest{
		UserID: "user_auto_renew_on",
		Device: "web",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// 初次验证
	if _, err := gs.GetAuthEngine().Verify(ctx, resp.Token); err != nil {
		t.Fatalf("verify failed: %v", err)
	}

	// 等待一段时间接近过期
	time.Sleep(1500 * time.Millisecond)

	// 再次验证触发续期
	if _, err := gs.GetAuthEngine().Verify(ctx, resp.Token); err != nil {
		t.Fatalf("verify after sleep failed: %v", err)
	}

	// 再等待超过原过期时间，若成功续期应仍有效
	time.Sleep(1200 * time.Millisecond)

	if _, err := gs.GetAuthEngine().Verify(ctx, resp.Token); err != nil {
		t.Fatalf("verify after renew failed (should still be valid): %v", err)
	}
}

// TestAutoRenewDisabled 验证关闭自动续期时，不更新 LastAccess，按原过期逻辑失效
func TestAutoRenewDisabled(t *testing.T) {
	cfg := config.NewBuilder().
		WithTokenExpire(1 * time.Second).
		WithRefreshExpire(0).
		WithAutoRenew(false).
		Build()

	gs := gstoken.New(cfg)
	ctx := context.Background()

	resp, err := gs.Login(ctx, &core.LoginRequest{
		UserID: "user_auto_renew_off",
		Device: "web",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// 初次验证
	if _, err := gs.GetAuthEngine().Verify(ctx, resp.Token); err != nil {
		t.Fatalf("verify failed: %v", err)
	}

	// 等待超过过期时间
	time.Sleep(1200 * time.Millisecond)

	// 应判定过期
	if _, err := gs.GetAuthEngine().Verify(ctx, resp.Token); err == nil {
		t.Fatalf("expected token expired, but verify succeeded")
	}
}

// TestRememberDaysIssueRefreshToken 当 RefreshExpire=0 且 RememberDays>0 时，应发放刷新Token并可刷新
func TestRememberDaysIssueRefreshToken(t *testing.T) {
	cfg := config.NewBuilder().
		WithTokenExpire(1 * time.Second).
		WithRefreshExpire(0). // 不设置刷新过期
		Build()
	// 手工设置 RememberDays
	cfg.RememberDays = 1

	gs := gstoken.New(cfg)
	ctx := context.Background()

	resp, err := gs.Login(ctx, &core.LoginRequest{
		UserID: "user_remember",
		Device: "web",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if resp.RefreshToken == "" {
		t.Fatalf("expected refresh token issued when RememberDays>0")
	}

	// 立即刷新，应获得新 access 与新的 refresh
	newResp, err := gs.RefreshToken(ctx, resp.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if newResp.Token == "" || newResp.RefreshToken == "" {
		t.Fatalf("expected new tokens on refresh")
	}
}
