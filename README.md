# GSToken

<div align="center">

![GSToken Logo](https://img.shields.io/badge/GSToken-v1.1.2-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)

**一个轻量级、高性能的 Golang 权限认证框架**

解决登录认证、权限控制、分布式会话、单点登录等企业级权限管理问题

[快速开始](#-快速开始) • [功能特性](#-功能特性) • [文档](#-文档) • [示例](#-示例代码) • [贡献](#-贡献)

</div>

---

## 🌟 核心优势

- **🚀 高性能** - 基于内存和Redis的高效存储，支持分布式部署
- **🔧 易集成** - 简洁的API设计，支持多种Web框架（Gin、Echo等）
- **🛡️ 安全可靠** - 完善的Token管理机制，支持自动续期和安全退出
- **📈 可扩展** - 模块化设计，支持自定义Token生成器和权限提供者
- **🌐 分布式** - 原生支持Redis集群，适合微服务架构
- **📊 生产就绪** - 完整的测试覆盖，已在生产环境验证

## 🎯 功能特性

### 🔐 认证管理
- **多模式登录** - 单端登录、多端登录、同端互斥登录
- **Token管理** - 6种内置Token风格，支持自定义生成策略
- **自动续期** - 智能的Token续期机制，提升用户体验
- **记住登录** - 7天内免登录功能

### 🛡️ 权限控制
- **RBAC权限** - 基于角色的权限认证系统
- **方法级鉴权** - 优雅的将鉴权与业务代码分离
- **权限缓存** - 高效的权限验证缓存机制

### 🌐 会话管理
- **分布式Session** - 支持Redis集群的分布式会话存储
- **会话控制** - 根据用户ID或Token踢人下线
- **会话监控** - 实时获取用户在线状态和会话信息

### 🔌 框架集成
- **Gin框架** - 完整的Gin中间件和辅助函数
- **通用适配** - 支持任意Web框架的适配器模式
- **中间件** - 开箱即用的认证中间件

## 📦 安装

```bash
go get -u github.com/luckxgo/gstoken
```

**系统要求：**
- Go 1.18+
- Redis 5.0+ (可选，用于分布式部署)

## 🚀 快速开始

### 基础使用

```go
package main

import (
    "context"
    "fmt"
    "github.com/luckxgo/gstoken"
    "github.com/luckxgo/gstoken/config"
    "github.com/luckxgo/gstoken/core"
)

func main() {
    // 1. 创建配置
    cfg := config.DefaultConfig()
    
    // 2. 初始化GSToken
    gs := gstoken.New(cfg)
    
    ctx := context.Background()
    
    // 3. 用户登录
    loginReq := &core.LoginRequest{
        UserID: "user123",
        Device: "web",
        IP:     "192.168.1.100",
        Extra: map[string]interface{}{
            "username": "张三",
            "role":     "admin",
        },
    }
    
    loginResp, err := gs.Login(ctx, loginReq)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("登录成功，Token: %s\n", loginResp.Token)
    
    // 4. Token验证
    userInfo, err := gs.GetAuthEngine().Verify(ctx, loginResp.Token)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("用户信息: %+v\n", userInfo)
    
    // 5. 权限检查
    hasPermission, err := gs.CheckPermission(ctx, "user123", "user:read")
    if err != nil {
        panic(err)
    }
    fmt.Printf("是否有权限: %v\n", hasPermission)
    
    // 6. 用户登出
    err = gs.Logout(ctx, loginResp.Token)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("登出成功")
}
```

### Gin框架集成

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/luckxgo/gstoken"
    "github.com/luckxgo/gstoken/config"
    "github.com/luckxgo/gstoken/core"
    "github.com/luckxgo/gstoken/web"
)

func main() {
    // 初始化GSToken
    cfg := config.DefaultConfig()
    gs := gstoken.New(cfg)
    
    r := gin.Default()
    
    // 登录接口
    r.POST("/login", func(c *gin.Context) {
        // 验证用户名密码...
        
        loginReq := &core.LoginRequest{
            UserID: "user123",
            Device: "web",
            IP:     c.ClientIP(),
        }
        
        resp, err := gs.Login(c, loginReq)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(200, resp)
    })
    
    // 需要认证的路由组
    auth := r.Group("/api")
    auth.Use(func(c *gin.Context) {
        // 从请求头获取Token
        token := c.GetHeader("Authorization")
        if token == "" {
            token = c.GetHeader("X-Token")
        }
        if token == "" {
            c.JSON(401, gin.H{"error": "未提供认证Token"})
            c.Abort()
            return
        }
        
        // 验证Token
        userInfo, err := gs.GetAuthEngine().Verify(c, token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Token验证失败"})
            c.Abort()
            return
        }
        
        // 设置上下文
        c.Set(web.ContextKeyUserID, userInfo.ID)
        c.Set(web.ContextKeyToken, token)
        c.Set(web.ContextKeyUserInfo, userInfo)
        
        c.Next()
    })
    {
        auth.GET("/profile", func(c *gin.Context) {
            userID, _ := c.Get(web.ContextKeyUserID)
            userInfo, _ := c.Get(web.ContextKeyUserInfo)
            
            c.JSON(200, gin.H{
                "user_id": userID,
                "user_info": userInfo,
            })
        })
        
        auth.POST("/logout", func(c *gin.Context) {
            token, _ := c.Get(web.ContextKeyToken)
            if tokenStr, ok := token.(string); ok {
                gs.Logout(c, tokenStr)
            }
            c.JSON(200, gin.H{"message": "登出成功"})
        })
    }
    
    r.Run(":8080")
}
```

## 🔧 配置说明

### 默认配置

```go
cfg := config.DefaultConfig()
// 等价于：
cfg := &core.Config{
    TokenExpire:   24 * time.Hour,     // Token过期时间24小时
    RefreshExpire: 7 * 24 * time.Hour, // 刷新Token过期时间7天
    TokenStyle:    core.StyleUUID,     // Token风格：UUID
    LoginMode:     core.MultiLogin,    // 登录模式：多端登录
    AutoRenew:     true,               // 自动续期：开启
    RememberDays:  7,                  // 记住登录：7天
    KeyPrefix:     "gstoken",          // 键前缀
    Storage: core.StorageConfig{
        Type: "memory",                // 存储类型：内存
    },
}
```

### Redis配置

```go
cfg := config.RedisConfig()
// 或自定义Redis配置：
cfg := config.NewBuilder().
    WithRedisStorage("localhost:6379", "", 0).
    WithTokenExpire(2 * time.Hour).
    WithLoginMode(core.SingleLogin).
    Build()
```

### 配置构建器

```go
cfg := config.NewBuilder().
    WithTokenExpire(2 * time.Hour).                    // Token过期时间
    WithRefreshExpire(30 * 24 * time.Hour).           // 刷新Token过期时间
    WithTokenStyle(core.StyleRandom32).               // Token风格
    WithLoginMode(core.SingleLogin).                  // 单端登录
    WithAutoRenew(false).                             // 关闭自动续期
    WithRememberDays(30).                             // 记住登录30天
    WithKeyPrefix("myapp").                           // 自定义键前缀
    WithRedisStorage("localhost:6379", "password", 1). // Redis存储
    Build()
```

## 🎨 Token风格

GSToken 支持6种内置Token风格：

| 风格 | 示例 | 特点 |
|------|------|------|
| `StyleUUID` | `550e8400-e29b-41d4-a716-446655440000` | 标准UUID格式，默认风格 |
| `StyleUUIDSimple` | `550e8400e29b41d4a716446655440000` | 简化UUID，去掉中划线 |
| `StyleRandom32` | `a1b2c3d4e5f6789012345678901234ab` | 32位随机字符串 |
| `StyleRandom64` | `a1b2c3d4...1234` | 64位随机字符串 |
| `StyleRandom128` | `a1b2c3d4...f12` | 128位随机字符串 |
| `StyleTik` | `tik_1640995200_a1b2c3d4e5f67890` | Tik风格，包含时间戳 |

### 自定义Token生成

```go
// 注册自定义Token生成函数
generator := gs.GetTokenGenerator()
err := generator.RegisterCustomFunc(func(extra map[string]interface{}) (string, error) {
    userID := extra[core.TokenExtraKeyUserID].(string)
    timestamp := time.Now().Unix()
    return fmt.Sprintf("USER_%s_%d", userID, timestamp), nil
})
if err != nil {
    panic(err)
}

// 使用自定义风格
err = generator.SetStyle(core.StyleCustom)
if err != nil {
    panic(err)
}
```

## 🏗️ 项目架构

```
gstoken/
├── auth/              # 认证模块
│   ├── engine.go      # 认证引擎
│   ├── service.go     # 认证服务
│   ├── session.go     # 会话管理
│   └── permission.go  # 权限管理
├── config/            # 配置管理
│   ├── builder.go     # 配置构建器
│   └── default.go     # 默认配置
├── core/              # 核心定义
│   ├── types.go       # 类型定义
│   ├── errors.go      # 错误定义
│   └── interfaces.go  # 接口定义
├── storage/           # 存储适配器
│   ├── memory.go      # 内存存储
│   └── redis.go       # Redis存储
├── token/             # Token生成器
│   └── generator.go   # Token生成器
├── web/               # Web框架适配
│   ├── gin_helper.go  # Gin辅助函数
│   └── constants.go   # Web常量
├── test/              # 测试文件
├── examples/          # 示例代码
├── docs/              # 文档
└── gstoken.go         # 主入口
```

## 📊 性能基准

```
BenchmarkLogin-8           100000    10234 ns/op    1024 B/op    12 allocs/op
BenchmarkVerify-8          200000     5123 ns/op     512 B/op     6 allocs/op
BenchmarkPermission-8      500000     2045 ns/op     256 B/op     3 allocs/op
BenchmarkRedisStorage-8     50000    20456 ns/op    2048 B/op    15 allocs/op
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行基准测试
go test -bench=. ./test

# 查看测试覆盖率
go test -cover ./...
```

## 📚 示例代码

- [基础认证示例](examples/basic/main.go)
- [Gin框架集成](examples/gin/main.go)
- [Redis存储示例](examples/redis/main.go)
- [权限控制示例](examples/permission/main.go)
- [自定义Token生成](examples/custom_token/main.go)

## 🔗 相关链接

- **GitHub:** https://github.com/luckxgo/gstoken
- **Gitee:** https://gitee.com/gs-token/gs-token
- **文档:** [在线文档](docs/)
- **问题反馈:** [Issues](https://github.com/luckxgo/gstoken/issues)

## 📈 版本历史

- **v1.1.2** (2024-09-30) - 魔法值重构优化，提升代码质量
- **v1.1.1** (2024-09-25) - 问题修复和性能优化
- **v1.1.0** (2024-09-20) - 新增权限管理和会话控制
- **v1.0.0** (2024-09-15) - 首个稳定版本发布

## 🤝 贡献

我们欢迎所有形式的贡献！

### 贡献方式

1. **Fork** 本仓库
2. **创建** 特性分支 (`git checkout -b feature/AmazingFeature`)
3. **提交** 更改 (`git commit -m 'Add some AmazingFeature'`)
4. **推送** 到分支 (`git push origin feature/AmazingFeature`)
5. **创建** Pull Request

### 开发指南

- 遵循 Go 代码规范
- 添加必要的测试用例
- 更新相关文档
- 确保所有测试通过

## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。

## 🙏 致谢

感谢所有为 GSToken 项目做出贡献的开发者！

---

<div align="center">

**如果这个项目对你有帮助，请给我们一个 ⭐️**

Made with ❤️ by GSToken Team

</div>