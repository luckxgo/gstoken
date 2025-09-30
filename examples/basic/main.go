package main

import (
	"context"
	"fmt"
	"log"

	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
)

func main() {
	fmt.Println("=== GSToken 基础认证示例 ===")

	// 1. 创建配置
	cfg := config.DefaultConfig()
	fmt.Printf("使用配置: Token过期时间=%v, 登录模式=%v\n", cfg.TokenExpire, cfg.LoginMode)

	// 2. 初始化GSToken
	gs := gstoken.New(cfg)
	fmt.Println("GSToken 初始化完成")

	ctx := context.Background()

	// 3. 用户登录
	fmt.Println("\n--- 用户登录 ---")
	loginReq := &core.LoginRequest{
		UserID: "user123",
		Device: "web",
		IP:     "192.168.1.100",
		Extra: map[string]interface{}{
			"username": "张三",
			"role":     "admin",
			"email":    "zhangsan@example.com",
		},
	}

	loginResp, err := gs.Login(ctx, loginReq)
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}

	fmt.Printf("登录成功!\n")
	fmt.Printf("Token: %s\n", loginResp.Token)
	fmt.Printf("过期时间: %v\n", loginResp.ExpireTime)
	if loginResp.RefreshToken != "" {
		fmt.Printf("刷新Token: %s\n", loginResp.RefreshToken)
	}

	// 4. Token验证
	fmt.Println("\n--- Token验证 ---")
	userInfo, err := gs.GetAuthEngine().Verify(ctx, loginResp.Token)
	if err != nil {
		log.Fatalf("Token验证失败: %v", err)
	}

	fmt.Printf("验证成功!\n")
	fmt.Printf("用户ID: %s\n", userInfo.ID)
	fmt.Printf("用户名: %s\n", userInfo.Username)
	fmt.Printf("角色: %v\n", userInfo.Roles)
	fmt.Printf("额外信息: %+v\n", userInfo.Extra)

	// 5. 检查登录状态
	fmt.Println("\n--- 检查登录状态 ---")
	isLogin := gs.IsLogin(ctx, loginResp.Token)
	fmt.Printf("是否已登录: %v\n", isLogin)

	// 6. 获取登录信息
	fmt.Println("\n--- 获取登录信息 ---")
	loginInfo, err := gs.GetLoginInfo(ctx, loginResp.Token)
	if err != nil {
		log.Printf("获取登录信息失败: %v", err)
	} else {
		fmt.Printf("登录时间: %v\n", loginInfo.LoginTime)
		fmt.Printf("最后访问: %v\n", loginInfo.LastAccess)
		fmt.Printf("设备: %s\n", loginInfo.Device)
		fmt.Printf("IP: %s\n", loginInfo.IP)
	}

	// 7. 用户登出
	fmt.Println("\n--- 用户登出 ---")
	err = gs.Logout(ctx, loginResp.Token)
	if err != nil {
		log.Fatalf("登出失败: %v", err)
	}

	fmt.Println("登出成功!")

	// 8. 验证登出后的状态
	fmt.Println("\n--- 验证登出后状态 ---")
	isLogin = gs.IsLogin(ctx, loginResp.Token)
	fmt.Printf("登出后是否还在登录: %v\n", isLogin)

	fmt.Println("\n=== 基础认证示例完成 ===")
}
