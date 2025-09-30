# GS-Token Web 框架集成

## 概述

GS-Token Web 模块提供了灵活的 Web 框架集成方案，支持 Gin 框架的中间件和方法式鉴权。

## 常量定义

为了避免魔法值，所有的键名和配置都统一定义在 `web/constants.go` 文件中：

### 文件结构
```
web/
├── constants.go    # 统一的常量定义文件
├── adapter.go      # 通用适配器实现
├── gin_adapter.go  # Gin框架适配器
├── decorator.go    # 方法式鉴权装饰器
└── README.md       # 使用文档
```

### 常量分类

**上下文键常量**
- `ContextKeyUserID` = "user_id" - 用户ID在上下文中的键名
- `ContextKeyToken` = "token" - Token在上下文中的键名  
- `ContextKeyUserInfo` = "user_info" - 用户信息在上下文中的键名

**HTTP头常量**
- `HeaderAuthorization` = "Authorization" - 授权头名称
- `HeaderXToken` = "X-Token" - 自定义Token头名称
- `BearerPrefix` = "Bearer " - Bearer Token前缀

**查询参数常量**
- `QueryParamToken` = "token" - Token查询参数名称

**错误响应常量**
- `ErrorUnauthorized` = "unauthorized" - 未授权错误类型
- `ErrorForbidden` = "forbidden" - 禁止访问错误类型
- `ErrorMessage` = "message" - 错误消息字段名

## 使用示例

### 1. 基本中间件使用

```go
package main

import (
    "gstoken"
    "gstoken/config"
    "gstoken/web"
    "github.com/gin-gonic/gin"
)

func main() {
    // 创建 GSToken 实例
    gs := gstoken.New(config.NewBuilder().WithMemoryStorage().Build())
    
    // 创建 Gin 应用
    r := gin.Default()
    
    // 创建认证中间件
    gsAdapter := web.NewGSTokenWebAdapter(gs)
    authMiddleware := web.NewGinAuthMiddleware(gsAdapter, nil)
    
    // 公开路由
    r.GET("/public", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "public endpoint"})
    })
    
    // 需要认证的路由组
    auth := r.Group("/api")
    auth.Use(authMiddleware.RequireAuth())
    {
        auth.GET("/profile", func(c *gin.Context) {
            // 使用常量获取用户信息
            userID, _ := web.Helper.GetUserID(c)
            c.JSON(200, gin.H{"user_id": userID})
        })
    }
    
    // 需要管理员角色的路由
    admin := r.Group("/admin")
    admin.Use(authMiddleware.RequireRole("admin"))
    {
        admin.GET("/users", func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "admin only"})
        })
    }
    
    r.Run(":8080")
}
```

### 2. 自定义配置使用常量

```go
// 自定义认证配置
config := &web.AuthConfig{
    TokenHeader: web.HeaderXToken,        // 使用 X-Token 头
    TokenQuery:  web.QueryParamToken,     // 支持查询参数
    TokenPrefix: "",                      // 不使用前缀
    // SkipPaths 支持精确匹配、前缀与通配符：
    // - "/health" 精确匹配
    // - "/public/*" 前缀匹配，以 /public/ 开头
    // - "/static/*.js"、"/api/user??" 通配符匹配
    SkipPaths:   []string{"/health", "/public/*"},
    UnauthorizedHandler: func(c web.WebContext, err error) {
        c.AbortWithJSON(401, map[string]interface{}{
            "error":           web.ErrorUnauthorized,
            web.ErrorMessage:  err.Error(),
            "timestamp":       time.Now().Unix(),
        })
    },
}
```

### 3. 获取用户信息

```go
func profileHandler(c *gin.Context) {
    // 使用辅助函数获取用户信息
    userID, exists := web.Helper.GetUserID(c)
    if !exists {
        c.JSON(400, gin.H{"error": "user not found"})
        return
    }
    
    token, _ := web.Helper.GetToken(c)
    userInfo, _ := web.Helper.GetUserInfo(c)
    
    c.JSON(200, gin.H{
        "user_id":   userID,
        "token":     token,
        "user_info": userInfo,
    })
}
```

### 4. 方法式鉴权装饰器

```go
// 创建装饰器
decorator := web.NewAuthDecorator(gsAdapter, nil)

// 业务函数
func getUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
    // 业务逻辑
    return &UserProfile{ID: userID}, nil
}

// 装饰后的函数
decoratedFunc := decorator.RequireAuth(getUserProfile)

// 调用装饰后的函数
authCtx := web.NewAuthContext(ctx, "user123", "token123", userInfo)
result := decoratedFunc.(func(context.Context, string) (*UserProfile, error))(authCtx, "user123")
```

## 最佳实践

1. **使用常量**: 始终使用预定义的常量，避免硬编码字符串
2. **错误处理**: 自定义错误处理函数，提供一致的错误响应格式
3. **配置管理**: 根据环境使用不同的认证配置
4. **测试覆盖**: 为所有认证场景编写测试用例

## 扩展其他框架

由于使用了适配器模式，可以轻松扩展支持其他 Web 框架：

1. 实现 `WebContext` 接口
2. 创建框架特定的中间件适配器
3. 使用相同的常量和配置系统

这样可以确保在不同框架间保持一致的 API 和行为。