package test

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
	"github.com/redis/go-redis/v9"
)

// TestRedisIntegration 测试使用 Redis 存储的完整认证流程
func TestRedisIntegration(t *testing.T) {
	// 从环境读取 Redis 配置，提供更稳健的探测
	addr := os.Getenv("GSTOKEN_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("GSTOKEN_REDIS_PASSWORD")
	db := 2
	if v := os.Getenv("GSTOKEN_REDIS_DB"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			db = parsed
		}
	}

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db, // 默认使用 DB 2 避免与其他数据冲突
	})

	// 使用短时上下文进行连通性与写权限探测
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	// 探测连接
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis 连接失败，跳过测试: %v (addr=%s, db=%d)", err, addr, db)
	}

	// 探测写权限
	probeKey := "__gstoken_test_probe"
	if err := client.Set(ctx, probeKey, "ok", time.Second).Err(); err != nil {
		t.Skipf("Redis 写入探测失败，跳过测试: %v (addr=%s, db=%d)", err, addr, db)
	}
	// 清理探测键
	_ = client.Del(ctx, probeKey).Err()

	// 清理测试数据
	defer func() {
		_ = client.FlushDB(context.Background())
		_ = client.Close()
	}()

	// 创建测试用的角色提供者
	roleProvider := &redisRoleProvider{client: client}

	// 创建配置
	cfg := config.NewBuilder().
		WithRedisStorage("localhost:6379", "", 2).
		WithUserRoleProvider(roleProvider).
		WithTokenExpire(time.Hour).
		WithRefreshExpire(time.Hour * 24).
		Build()

	// 创建认证引擎
	engine := gstoken.New(cfg)

	t.Run("Redis存储-完整认证流程", func(t *testing.T) {
		testUserID := "redis_test_user"
		testDevice := "web"

		// 设置测试用户的角色数据
		err := roleProvider.setupTestData(ctx, testUserID)
		if err != nil {
			t.Fatalf("设置测试数据失败: %v", err)
		}

		// 1. 用户登录
		t.Log("=== 步骤1: 用户登录 ===")
		loginReq := &core.LoginRequest{
			UserID: testUserID,
			Device: testDevice,
		}
		loginResult, err := engine.Login(ctx, loginReq)
		if err != nil {
			t.Fatalf("登录失败: %v", err)
		}

		token := loginResult.Token
		t.Logf("登录成功: Token=%s", token)
		t.Logf("过期时间: %v", loginResult.ExpireTime)

		// 2. 验证 Token
		t.Log("\n=== 步骤2: 验证 Token ===")
		loginInfo2, err := engine.GetLoginInfo(ctx, token)
		if err != nil {
			t.Fatalf("Token验证失败: %v", err)
		}
		t.Logf("Token验证成功: UserID=%s", loginInfo2.UserID)

		if loginInfo2.UserID != testUserID {
			t.Errorf("用户ID不匹配: 期望=%s, 实际=%s", testUserID, loginInfo2.UserID)
		}

		// 3. 检查角色权限
		t.Log("\n=== 步骤3: 检查角色权限 ===")
		hasAdminRole, err := engine.CheckRole(ctx, testUserID, "admin")
		if err != nil {
			t.Fatalf("检查管理员角色失败: %v", err)
		}
		if !hasAdminRole {
			t.Error("用户应该具有管理员角色")
		}
		t.Log("✅ 管理员角色检查通过")

		// 4. 检查权限
		t.Log("\n=== 步骤4: 检查权限 ===")
		hasReadPerm, err := engine.CheckPermission(ctx, testUserID, "user:read")
		if err != nil {
			t.Fatalf("检查读取权限失败: %v", err)
		}
		if !hasReadPerm {
			t.Error("用户应该具有读取权限")
		}
		t.Log("✅ 读取权限检查通过")

		// 5. 验证存储数据一致性
		t.Log("\n=== 步骤5: 验证存储数据一致性 ===")

		// 直接从 Redis 获取登录信息
		loginKey := "gstoken:login:" + token
		loginData, err := client.Get(ctx, loginKey).Result()
		if err != nil {
			t.Fatalf("获取登录信息失败: %v", err)
		}

		// 验证数据格式
		t.Logf("✅ 登录信息数据长度: %d", len(loginData))
		t.Logf("登录信息内容预览: %.100s...", loginData)

		// 验证可以正确解析
		var loginInfo core.LoginInfo
		if err := json.Unmarshal([]byte(loginData), &loginInfo); err != nil {
			t.Fatalf("解析登录信息失败: %v", err)
		}

		if loginInfo.UserID != testUserID {
			t.Errorf("登录信息用户ID不匹配: 期望=%s, 实际=%s", testUserID, loginInfo.UserID)
		}
		if loginInfo.Device != testDevice {
			t.Errorf("登录信息设备不匹配: 期望=%s, 实际=%s", testDevice, loginInfo.Device)
		}
		t.Log("✅ 登录信息数据一致性验证通过")

		// 6. 登出测试
		t.Log("\n=== 步骤6: 用户登出 ===")
		err = engine.Logout(ctx, token)
		if err != nil {
			t.Fatalf("登出失败: %v", err)
		}
		t.Log("✅ 登出成功")

		// 7. 验证 Token 已失效
		t.Log("\n=== 步骤7: 验证 Token 已失效 ===")
		_, err = engine.GetLoginInfo(ctx, token)
		if err == nil {
			t.Error("Token应该已经失效")
		}
		t.Log("✅ Token已成功失效")

		t.Log("\n🎉 Redis存储完整认证流程测试通过!")
	})

	t.Run("Redis存储-数据格式验证", func(t *testing.T) {
		t.Log("=== Redis 存储数据格式验证 ===")

		// 测试不同类型的数据存储
		testCases := []struct {
			name string
			key  string
			data interface{}
		}{
			{"字符串", "test:string", "hello world"},
			{"数字", "test:number", 12345},
			{"对象", "test:object", map[string]interface{}{
				"id":    "test123",
				"name":  "测试用户",
				"roles": []string{"admin", "user"},
			}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 序列化并存储
				jsonData, err := json.Marshal(tc.data)
				if err != nil {
					t.Fatalf("序列化失败: %v", err)
				}

				err = client.Set(ctx, tc.key, jsonData, time.Minute).Err()
				if err != nil {
					t.Fatalf("存储失败: %v", err)
				}

				// 获取并验证
				retrievedData, err := client.Get(ctx, tc.key).Result()
				if err != nil {
					t.Fatalf("获取失败: %v", err)
				}

				// 验证数据一致性
				if retrievedData != string(jsonData) {
					t.Errorf("数据不匹配")
				}

				t.Logf("✅ %s 存储和获取一致性验证通过", tc.name)
			})
		}
	})
}

// redisRoleProvider Redis 角色提供者实现
type redisRoleProvider struct {
	client *redis.Client
}

func (p *redisRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
	key := "gstoken:user_roles:" + userID
	data, err := p.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []core.Role{}, nil
		}
		return nil, err
	}

	var roleIDs []string
	if err := json.Unmarshal([]byte(data), &roleIDs); err != nil {
		return nil, err
	}

	// 获取角色详细信息
	var roles []core.Role
	for _, roleID := range roleIDs {
		role, err := p.GetRole(ctx, roleID)
		if err != nil {
			continue
		}
		if role != nil {
			roles = append(roles, *role)
		}
	}
	return roles, nil
}

func (p *redisRoleProvider) GetRole(ctx context.Context, roleID string) (*core.Role, error) {
	key := "gstoken:role:" + roleID
	data, err := p.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var role core.Role
	if err := json.Unmarshal([]byte(data), &role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (p *redisRoleProvider) setupTestData(ctx context.Context, userID string) error {
	// 设置用户角色
	userRoles := []string{"admin", "user"}
	userRolesJSON, _ := json.Marshal(userRoles)
	err := p.client.Set(ctx, "gstoken:user_roles:"+userID, userRolesJSON, time.Hour).Err()
	if err != nil {
		return err
	}

	// 设置管理员角色
	adminRole := &core.Role{
		ID:          "admin",
		Name:        "管理员",
		Permissions: []string{"user:read", "user:write", "user:delete", "system:admin"},
	}
	adminRoleJSON, _ := json.Marshal(adminRole)
	err = p.client.Set(ctx, "gstoken:role:admin", adminRoleJSON, time.Hour).Err()
	if err != nil {
		return err
	}

	// 设置用户角色
	userRole := &core.Role{
		ID:          "user",
		Name:        "普通用户",
		Permissions: []string{"user:read", "profile:edit"},
	}
	userRoleJSON, _ := json.Marshal(userRole)
	return p.client.Set(ctx, "gstoken:role:user", userRoleJSON, time.Hour).Err()
}
