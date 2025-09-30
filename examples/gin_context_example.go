package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luckxgo/gstoken/web"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 演示双重上下文支持的中间件
	r.Use(func(c *gin.Context) {
		// 创建 GinContext 适配器
		ginCtx := web.NewGinContext(c)

		// 设置一些值到上下文
		ginCtx.Set("user_id", "12345")
		ginCtx.Set("role", "admin")
		ginCtx.Set("permissions", []string{"read", "write", "delete"})

		c.Next()
	})

	// 路由处理器：演示从两种上下文获取值
	r.GET("/demo", func(c *gin.Context) {
		ginCtx := web.NewGinContext(c)

		fmt.Println("=== 演示双重上下文支持 ===")

		// 方式1: 通过 GinContext 获取值
		if userID, exists := ginCtx.Get("user_id"); exists {
			fmt.Printf("通过 GinContext 获取 user_id: %v\n", userID)
		}

		// 方式2: 通过标准库 context 获取值
		stdCtx := ginCtx.GetContext()
		if userID := stdCtx.Value("user_id"); userID != nil {
			fmt.Printf("通过 context.Context 获取 user_id: %v\n", userID)
		}

		// 演示在标准库函数中使用
		processWithStandardContext(stdCtx)

		c.JSON(http.StatusOK, gin.H{
			"message": "双重上下文支持演示完成",
			"user_id": stdCtx.Value("user_id"),
			"role":    stdCtx.Value("role"),
		})
	})

	fmt.Println("服务器启动在 :8080")
	fmt.Println("访问 http://localhost:8080/demo 查看演示")
	r.Run(":8080")
}

// 模拟一个需要标准库 context 的函数
func processWithStandardContext(ctx context.Context) {
	fmt.Println("\n--- 在标准库函数中使用 context ---")

	if userID := ctx.Value("user_id"); userID != nil {
		fmt.Printf("标准库函数获取到 user_id: %v\n", userID)
	}

	if role := ctx.Value("role"); role != nil {
		fmt.Printf("标准库函数获取到 role: %v\n", role)
	}

	if permissions := ctx.Value("permissions"); permissions != nil {
		fmt.Printf("标准库函数获取到 permissions: %v\n", permissions)
	}
}
