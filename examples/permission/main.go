package main

import (
	"context"
	"fmt"
	"log"

	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
)

// MockUserRoleProvider 模拟用户角色提供者
type MockUserRoleProvider struct {
	userRoles map[string][]core.Role
}

// GetUserRoles 获取用户角色
func (p *MockUserRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
	roles, exists := p.userRoles[userID]
	if !exists {
		return []core.Role{}, nil
	}
	return roles, nil
}

// 创建模拟的用户角色数据
func createMockUserRoles() *MockUserRoleProvider {
	return &MockUserRoleProvider{
		userRoles: map[string][]core.Role{
			"admin_user": {
				{
					ID:   "admin",
					Name: "管理员",
					Permissions: []string{
						"user:read", "user:write", "user:delete",
						"system:read", "system:write", "system:config",
						"*", // 通配符权限，拥有所有权限
					},
				},
			},
			"manager_user": {
				{
					ID:   "manager",
					Name: "经理",
					Permissions: []string{
						"user:read", "user:write",
						"report:read", "report:write",
						"team:manage",
					},
				},
			},
			"normal_user": {
				{
					ID:   "user",
					Name: "普通用户",
					Permissions: []string{
						"user:read",
						"profile:read", "profile:write",
					},
				},
			},
		},
	}
}

func main() {
	fmt.Println("=== GSToken 权限控制示例 ===")

	// 1. 创建配置并设置用户角色提供者
	roleProvider := createMockUserRoles()
	cfg := config.DefaultConfig()
	cfg.UserRoleProvider = roleProvider

	fmt.Println("配置用户角色提供者完成")

	// 2. 初始化GSToken
	gs := gstoken.New(cfg)
	fmt.Println("GSToken 初始化完成")

	ctx := context.Background()

	// 3. 创建不同角色的用户并登录
	fmt.Println("\n--- 不同角色用户登录 ---")
	
	users := []struct {
		UserID   string
		Username string
		Role     string
	}{
		{"admin_user", "管理员张三", "admin"},
		{"manager_user", "经理李四", "manager"},
		{"normal_user", "普通用户王五", "user"},
	}

	var userTokens = make(map[string]string)
	
	for _, user := range users {
		loginReq := &core.LoginRequest{
			UserID: user.UserID,
			Device: "web",
			IP:     "192.168.1.100",
			Extra: map[string]interface{}{
				"username": user.Username,
				"role":     user.Role,
			},
		}

		loginResp, err := gs.Login(ctx, loginReq)
		if err != nil {
			log.Printf("用户 %s 登录失败: %v", user.Username, err)
			continue
		}

		userTokens[user.UserID] = loginResp.Token
		fmt.Printf("用户 %s 登录成功\n", user.Username)
	}

	// 4. 测试权限检查
	fmt.Println("\n--- 权限检查测试 ---")
	
	permissions := []string{
		"user:read",
		"user:write", 
		"user:delete",
		"system:config",
		"report:read",
		"profile:write",
		"unknown:permission",
	}

	for _, userID := range []string{"admin_user", "manager_user", "normal_user"} {
		fmt.Printf("\n用户 %s 的权限检查:\n", userID)
		
		for _, permission := range permissions {
			hasPermission, err := gs.CheckPermission(ctx, userID, permission)
			if err != nil {
				log.Printf("  检查权限 %s 失败: %v", permission, err)
				continue
			}
			
			status := "❌"
			if hasPermission {
				status = "✅"
			}
			fmt.Printf("  %s %s\n", status, permission)
		}
	}

	// 5. 测试角色检查
	fmt.Println("\n--- 角色检查测试 ---")
	
	roles := []string{"admin", "manager", "user", "guest"}
	
	for _, userID := range []string{"admin_user", "manager_user", "normal_user"} {
		fmt.Printf("\n用户 %s 的角色检查:\n", userID)
		
		for _, roleID := range roles {
			hasRole, err := gs.CheckRole(ctx, userID, roleID)
			if err != nil {
				log.Printf("  检查角色 %s 失败: %v", roleID, err)
				continue
			}
			
			status := "❌"
			if hasRole {
				status = "✅"
			}
			fmt.Printf("  %s %s\n", status, roleID)
		}
	}

	// 6. 模拟业务场景权限检查
	fmt.Println("\n--- 业务场景权限检查 ---")
	
	scenarios := []struct {
		name       string
		userID     string
		permission string
		action     string
	}{
		{"管理员删除用户", "admin_user", "user:delete", "删除用户账号"},
		{"经理查看报表", "manager_user", "report:read", "查看销售报表"},
		{"普通用户修改资料", "normal_user", "profile:write", "修改个人资料"},
		{"普通用户删除用户", "normal_user", "user:delete", "删除其他用户"},
		{"经理配置系统", "manager_user", "system:config", "修改系统配置"},
	}

	for _, scenario := range scenarios {
		hasPermission, err := gs.CheckPermission(ctx, scenario.userID, scenario.permission)
		if err != nil {
			log.Printf("场景 '%s' 权限检查失败: %v", scenario.name, err)
			continue
		}

		if hasPermission {
			fmt.Printf("✅ %s: 允许%s\n", scenario.name, scenario.action)
		} else {
			fmt.Printf("❌ %s: 拒绝%s (权限不足)\n", scenario.name, scenario.action)
		}
	}

	// 7. 动态权限检查（模拟中间件场景）
	fmt.Println("\n--- 动态权限检查（中间件场景）---")
	
	// 模拟HTTP请求的权限检查
	httpRequests := []struct {
		method     string
		path       string
		userID     string
		permission string
	}{
		{"GET", "/api/users", "normal_user", "user:read"},
		{"POST", "/api/users", "manager_user", "user:write"},
		{"DELETE", "/api/users/123", "admin_user", "user:delete"},
		{"GET", "/api/system/config", "normal_user", "system:read"},
		{"PUT", "/api/system/config", "admin_user", "system:config"},
	}

	for _, req := range httpRequests {
		hasPermission, err := gs.CheckPermission(ctx, req.userID, req.permission)
		if err != nil {
			log.Printf("请求 %s %s 权限检查失败: %v", req.method, req.path, err)
			continue
		}

		if hasPermission {
			fmt.Printf("✅ %s %s: 用户 %s 访问允许\n", req.method, req.path, req.userID)
		} else {
			fmt.Printf("❌ %s %s: 用户 %s 访问被拒绝 (需要权限: %s)\n", 
				req.method, req.path, req.userID, req.permission)
		}
	}

	// 8. 清理会话
	fmt.Println("\n--- 清理会话 ---")
	for userID, token := range userTokens {
		err := gs.Logout(ctx, token)
		if err != nil {
			log.Printf("用户 %s 登出失败: %v", userID, err)
		} else {
			fmt.Printf("用户 %s 登出成功\n", userID)
		}
	}

	fmt.Println("\n=== 权限控制示例完成 ===")
	fmt.Println("\n权限控制要点:")
	fmt.Println("1. 实现 UserRoleProvider 接口来提供用户角色信息")
	fmt.Println("2. 在配置中设置 UserRoleProvider")
	fmt.Println("3. 使用 CheckPermission 检查具体权限")
	fmt.Println("4. 使用 CheckRole 检查用户角色")
	fmt.Println("5. 支持通配符权限 '*' 表示拥有所有权限")
}