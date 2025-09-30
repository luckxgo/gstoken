package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
)

func main() {
	fmt.Println("=== GSToken 自定义Token生成示例 ===")

	// 1. 创建基础配置
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)

	ctx := context.Background()

	// 2. 演示内置Token风格
	fmt.Println("\n--- 内置Token风格演示 ---")

	tokenStyles := []struct {
		style core.TokenStyle
		name  string
	}{
		{core.StyleUUID, "UUID风格"},
		{core.StyleUUIDSimple, "UUID简化风格"},
		{core.StyleRandom32, "32位随机字符串"},
		{core.StyleRandom64, "64位随机字符串"},
		{core.StyleRandom128, "128位随机字符串"},
		{core.StyleTik, "Tik风格"},
	}

	generator := gs.GetTokenGenerator()

	for _, style := range tokenStyles {
		err := generator.SetStyle(style.style)
		if err != nil {
			log.Printf("设置Token风格失败: %v", err)
			continue
		}

		token, err := generator.Generate(map[string]interface{}{
			core.TokenExtraKeyUserID: "demo_user",
			core.TokenExtraKeyDevice: "web",
		})
		if err != nil {
			log.Printf("生成Token失败: %v", err)
			continue
		}

		fmt.Printf("%s: %s\n", style.name, token)
	}

	// 3. 自定义Token生成函数
	fmt.Println("\n--- 自定义Token生成函数 ---")

	// 业务前缀风格
	businessPrefixFunc := func(extra map[string]interface{}) (string, error) {
		userID := extra[core.TokenExtraKeyUserID].(string)
		device := extra[core.TokenExtraKeyDevice].(string)
		timestamp := time.Now().Unix()

		// 格式: BIZ_设备_用户ID_时间戳
		return fmt.Sprintf("BIZ_%s_%s_%d",
			strings.ToUpper(device), userID, timestamp), nil
	}

	// 注册并使用业务前缀风格
	err := generator.RegisterCustomFunc(businessPrefixFunc)
	if err != nil {
		log.Fatalf("注册自定义Token函数失败: %v", err)
	}

	err = generator.SetStyle(core.StyleCustom)
	if err != nil {
		log.Fatalf("设置自定义Token风格失败: %v", err)
	}

	token, err := generator.Generate(map[string]interface{}{
		core.TokenExtraKeyUserID: "user123",
		core.TokenExtraKeyDevice: "mobile",
	})
	if err != nil {
		log.Fatalf("生成自定义Token失败: %v", err)
	}

	fmt.Printf("业务前缀风格: %s\n", token)

	// 4. 多种自定义Token风格
	fmt.Println("\n--- 多种自定义Token风格 ---")

	customStyles := []struct {
		name string
		fn   core.CustomTokenFunc
	}{
		{
			"JWT风格前缀",
			func(extra map[string]interface{}) (string, error) {
				userID := extra[core.TokenExtraKeyUserID].(string)
				timestamp := time.Now().Unix()
				return fmt.Sprintf("jwt_%s_%d_signature", userID, timestamp), nil
			},
		},
		{
			"部门编码风格",
			func(extra map[string]interface{}) (string, error) {
				userID := extra[core.TokenExtraKeyUserID].(string)
				dept := "IT" // 可以从extra中获取部门信息
				if deptValue, ok := extra["department"]; ok {
					dept = deptValue.(string)
				}
				timestamp := time.Now().Format("20060102150405")
				return fmt.Sprintf("%s_%s_%s", dept, userID, timestamp), nil
			},
		},
		{
			"版本化Token",
			func(extra map[string]interface{}) (string, error) {
				userID := extra[core.TokenExtraKeyUserID].(string)
				version := "v1"
				if versionValue, ok := extra["version"]; ok {
					version = versionValue.(string)
				}
				timestamp := time.Now().Unix()
				return fmt.Sprintf("%s_%s_%d", version, userID, timestamp), nil
			},
		},
	}

	for _, style := range customStyles {
		err := generator.RegisterCustomFunc(style.fn)
		if err != nil {
			log.Printf("注册 %s 失败: %v", style.name, err)
			continue
		}

		token, err := generator.Generate(map[string]interface{}{
			core.TokenExtraKeyUserID: "user456",
			core.TokenExtraKeyDevice: "web",
			"department":             "SALES",
			"version":                "v2",
		})
		if err != nil {
			log.Printf("生成 %s Token失败: %v", style.name, err)
			continue
		}

		fmt.Printf("%s: %s\n", style.name, token)
	}

	// 5. 在实际登录中使用自定义Token
	fmt.Println("\n--- 实际登录中使用自定义Token ---")

	// 设置一个实用的自定义Token生成器
	practicalFunc := func(extra map[string]interface{}) (string, error) {
		// 安全地获取用户ID
		userID := "unknown"
		if uid, ok := extra[core.TokenExtraKeyUserID]; ok && uid != nil {
			userID = uid.(string)
		}
		
		// 安全地获取设备信息
		device := "unknown"
		if dev, ok := extra[core.TokenExtraKeyDevice]; ok && dev != nil {
			device = dev.(string)
		}

		// 获取用户类型
		userType := "USER"
		if typeValue, ok := extra["user_type"]; ok && typeValue != nil {
			userType = strings.ToUpper(typeValue.(string))
		}

		// 生成时间戳
		timestamp := time.Now().Format("20060102150405")

		// 生成随机后缀
		suffix := fmt.Sprintf("%d", time.Now().UnixNano()%10000)

		return fmt.Sprintf("%s_%s_%s_%s_%s",
			userType, device, userID, timestamp, suffix), nil
	}

	err = generator.RegisterCustomFunc(practicalFunc)
	if err != nil {
		log.Fatalf("注册实用Token函数失败: %v", err)
	}

	// 创建新的GSToken实例使用自定义Token
	customCfg := config.DefaultConfig()
	customCfg.TokenStyle = core.StyleCustom
	customGs := gstoken.New(customCfg)

	// 注册自定义函数到新实例
	customGenerator := customGs.GetTokenGenerator()
	err = customGenerator.RegisterCustomFunc(practicalFunc)
	if err != nil {
		log.Fatalf("为新实例注册自定义Token函数失败: %v", err)
	}

	err = customGenerator.SetStyle(core.StyleCustom)
	if err != nil {
		log.Fatalf("为新实例设置自定义Token风格失败: %v", err)
	}

	// 执行登录
	loginReq := &core.LoginRequest{
		UserID: "admin001",
		Device: "web",
		IP:     "192.168.1.100",
		Extra: map[string]interface{}{
			"username":  "管理员",
			"user_type": "admin",
			"role":      "administrator",
		},
	}

	loginResp, err := customGs.Login(ctx, loginReq)
	if err != nil {
		log.Fatalf("使用自定义Token登录失败: %v", err)
	}

	fmt.Printf("自定义Token登录成功: %s\n", loginResp.Token)

	// 验证自定义Token
	userInfo, err := customGs.GetAuthEngine().Verify(ctx, loginResp.Token)
	if err != nil {
		log.Fatalf("验证自定义Token失败: %v", err)
	}

	fmt.Printf("自定义Token验证成功: 用户=%s, 角色=%v\n",
		userInfo.Username, userInfo.Extra["role"])

	// 6. Token解析示例（如果实现了Parse方法）
	fmt.Println("\n--- Token信息解析 ---")

	// 注意：这里的Parse方法可能需要根据实际实现来调整
	tokenInfo, err := customGenerator.Parse(loginResp.Token)
	if err != nil {
		log.Printf("Token解析失败: %v", err)
	} else {
		fmt.Printf("Token信息: 用户ID=%s, 过期时间=%v\n",
			tokenInfo.UserID, tokenInfo.ExpireTime)
	}

	// 7. 清理
	fmt.Println("\n--- 清理会话 ---")
	err = customGs.Logout(ctx, loginResp.Token)
	if err != nil {
		log.Printf("登出失败: %v", err)
	} else {
		fmt.Println("登出成功")
	}

	fmt.Println("\n=== 自定义Token生成示例完成 ===")
	fmt.Println("\n自定义Token要点:")
	fmt.Println("1. 实现 CustomTokenFunc 函数类型")
	fmt.Println("2. 使用 RegisterCustomFunc 注册自定义函数")
	fmt.Println("3. 设置 TokenStyle 为 StyleCustom")
	fmt.Println("4. 可以在Token中包含业务信息，如部门、用户类型等")
	fmt.Println("5. 确保Token的唯一性和安全性")
}
