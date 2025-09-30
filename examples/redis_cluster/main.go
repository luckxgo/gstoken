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
	fmt.Println("=== GSToken Redis 集群示例 ===")

	// 1) 构建 Redis 集群配置
	cfg := config.NewBuilder().
		WithRedisCluster([]string{
			"localhost:7001",
			"localhost:7002",
			"localhost:7003",
		}, ""). // 如果有密码，填入此处
		WithRedisClientName("gstoken-client").
		WithRedisRetries(5, 10*time.Millisecond, 500*time.Millisecond).
		WithRedisTimeouts(5*time.Second, 3*time.Second, 3*time.Second, 4*time.Second).
		WithRedisPool(50, 10, 5*time.Minute).
		Build()

	// 设置键前缀以便区分
	cfg.KeyPrefix = "gstoken_cluster_demo"

	// 2) 初始化 GSToken
	gs := gstoken.New(cfg)
	fmt.Println("GSToken 初始化完成，使用 Redis 集群")

	ctx := context.Background()

	// 3) 简单登录两个用户
	for i := 1; i <= 2; i++ {
		_, err := gs.Login(ctx, &core.LoginRequest{
			UserID: fmt.Sprintf("cluster_user_%d", i),
			Device: "web",
			IP:     "127.0.0.1",
			Extra:  map[string]interface{}{"role": "user"},
		})
		if err != nil {
			log.Printf("登录失败: %v", err)
		}
	}

	// 4) 验证 Keys 使用 SCAN 能返回匹配键
	keys, err := gs.GetStorage().Keys(ctx, cfg.KeyPrefix+":*")
	if err != nil {
		log.Fatalf("Keys 扫描失败: %v", err)
	}
	fmt.Println("匹配键数量:", len(keys))
	for i, k := range keys {
		if i >= 10 {
			fmt.Println("...省略其余键")
			break
		}
		fmt.Println(k)
	}

	fmt.Println("=== Redis 集群示例完成 ===")
}
