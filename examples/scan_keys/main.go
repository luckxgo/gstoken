package main

import (
	"context"
	"fmt"
	"time"

	"github.com/luckxgo/gstoken"
	"github.com/luckxgo/gstoken/config"
	"github.com/luckxgo/gstoken/core"
)

func main() {
	ctx := context.Background()

	// 使用单机或集群均可，按需切换
	cfg := config.NewBuilder().
		WithRedisStorage("localhost:6379", "", 0).
		WithTokenExpire(10 * time.Minute).
		Build()
	cfg.KeyPrefix = "gstoken_demo"

	gs := gstoken.New(cfg)

	// 简单登录一两个用户，生成若干键
	for i := 1; i <= 2; i++ {
		_, _ = gs.Login(ctx, &core.LoginRequest{
			UserID: fmt.Sprintf("u%03d", i),
			Device: "web",
			IP:     "127.0.0.1",
			Extra:  map[string]interface{}{"username": fmt.Sprintf("user%d", i)},
		})
	}

	// 验证 Keys（SCAN）能返回匹配键
	st := gs.GetStorage()
	keys, err := st.Keys(ctx, cfg.KeyPrefix+":*")
	if err != nil {
		fmt.Println("Keys error:", err)
		return
	}
	fmt.Println("Matched keys count:", len(keys))
	for _, k := range keys {
		fmt.Println(k)
	}
}
