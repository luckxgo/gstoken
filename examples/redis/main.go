package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
)

func main() {
	fmt.Println("=== GSToken Redis存储示例 ===")

	// 1. 创建Redis配置
	cfg := config.NewBuilder().
		WithRedisStorage("localhost:6379", "", 0). // Redis地址、密码、数据库
		WithTokenExpire(30 * time.Minute).         // Token过期时间30分钟
		WithRefreshExpire(7 * 24 * time.Hour).     // 刷新Token过期时间7天
		WithLoginMode(core.MultiLogin).            // 多端登录模式
		WithAutoRenew(true).                       // 开启自动续期
		Build()

	// 手动设置键前缀
	cfg.KeyPrefix = "gstoken_demo"

	fmt.Printf("Redis配置: %s, DB: %d, 键前缀: %s\n",
		cfg.Redis.Addr, cfg.Redis.DB, cfg.KeyPrefix)

	// 2. 初始化GSToken
	gs := gstoken.New(cfg)
	fmt.Println("GSToken 初始化完成，使用Redis存储")

	ctx := context.Background()

	// 3. 测试多个用户登录
	fmt.Println("\n--- 多用户登录测试 ---")

	users := []struct {
		UserID   string
		Username string
		Role     string
		Device   string
	}{
		{"user001", "张三", "admin", "web"},
		{"user002", "李四", "user", "mobile"},
		{"user003", "王五", "manager", "desktop"},
	}

	var tokens []string
	for _, user := range users {
		loginReq := &core.LoginRequest{
			UserID: user.UserID,
			Device: user.Device,
			IP:     "192.168.1.100",
			Extra: map[string]interface{}{
				"username": user.Username,
				"role":     user.Role,
				"device":   user.Device,
			},
		}

		loginResp, err := gs.Login(ctx, loginReq)
		if err != nil {
			log.Printf("用户 %s 登录失败: %v", user.Username, err)
			continue
		}

		tokens = append(tokens, loginResp.Token)
		fmt.Printf("用户 %s 登录成功, Token: %s\n", user.Username, loginResp.Token[:20]+"...")
	}

	// 4. 验证所有Token
	fmt.Println("\n--- Token验证测试 ---")
	for i, token := range tokens {
		userInfo, err := gs.GetAuthEngine().Verify(ctx, token)
		if err != nil {
			log.Printf("Token %d 验证失败: %v", i+1, err)
			continue
		}
		fmt.Printf("Token %d 验证成功: 用户=%s, 角色=%v\n",
			i+1, userInfo.Username, userInfo.Extra["role"])
	}

	// 5. 测试会话持久化
	fmt.Println("\n--- 会话持久化测试 ---")
	if len(tokens) > 0 {
		testToken := tokens[0]

		// 获取登录信息
		loginInfo, err := gs.GetLoginInfo(ctx, testToken)
		if err != nil {
			log.Printf("获取登录信息失败: %v", err)
		} else {
			fmt.Printf("登录信息: 用户=%s, 设备=%s, 登录时间=%v\n",
				loginInfo.UserID, loginInfo.Device, loginInfo.LoginTime)
		}

		// 模拟重启应用（重新创建GSToken实例）
		fmt.Println("模拟应用重启...")
		gs2 := gstoken.New(cfg)

		// 验证Token是否仍然有效
		userInfo, err := gs2.GetAuthEngine().Verify(ctx, testToken)
		if err != nil {
			log.Printf("重启后Token验证失败: %v", err)
		} else {
			fmt.Printf("重启后Token仍然有效: 用户=%s\n", userInfo.Username)
		}
	}

	// 6. 测试Token刷新
	fmt.Println("\n--- Token刷新测试 ---")
	if len(tokens) > 0 {
		// 创建一个带刷新Token的登录
		loginReq := &core.LoginRequest{
			UserID: "refresh_user",
			Device: "web",
			IP:     "192.168.1.100",
			Extra: map[string]interface{}{
				"username": "刷新测试用户",
				"role":     "user",
			},
		}

		loginResp, err := gs.Login(ctx, loginReq)
		if err != nil {
			log.Printf("刷新测试用户登录失败: %v", err)
		} else if loginResp.RefreshToken != "" {
			fmt.Printf("获得刷新Token: %s\n", loginResp.RefreshToken[:20]+"...")

			// 使用刷新Token获取新的访问Token
			newResp, err := gs.RefreshToken(ctx, loginResp.RefreshToken)
			if err != nil {
				log.Printf("Token刷新失败: %v", err)
			} else {
				fmt.Printf("Token刷新成功, 新Token: %s\n", newResp.Token[:20]+"...")
			}
		}
	}

	// 7. 测试批量登出
	fmt.Println("\n--- 批量登出测试 ---")
	if len(users) > 0 {
		// 根据用户ID登出所有会话
		testUserID := users[0].UserID
		err := gs.LogoutByUserID(ctx, testUserID)
		if err != nil {
			log.Printf("用户 %s 批量登出失败: %v", testUserID, err)
		} else {
			fmt.Printf("用户 %s 的所有会话已登出\n", testUserID)

			// 验证该用户的Token是否已失效
			if len(tokens) > 0 {
				_, err := gs.GetAuthEngine().Verify(ctx, tokens[0])
				if err != nil {
					fmt.Printf("确认: 用户Token已失效\n")
				} else {
					fmt.Printf("警告: 用户Token仍然有效\n")
				}
			}
		}
	}

	// 8. 清理剩余会话
	fmt.Println("\n--- 清理剩余会话 ---")
	for i, token := range tokens[1:] { // 跳过第一个已经登出的
		err := gs.Logout(ctx, token)
		if err != nil {
			log.Printf("Token %d 登出失败: %v", i+2, err)
		} else {
			fmt.Printf("Token %d 登出成功\n", i+2)
		}
	}

	fmt.Println("\n=== Redis存储示例完成 ===")
	fmt.Println("提示: 可以使用 redis-cli 查看存储的数据:")
	fmt.Printf("redis-cli -n %d KEYS \"%s:*\"\n", cfg.Redis.DB, cfg.KeyPrefix)
}
