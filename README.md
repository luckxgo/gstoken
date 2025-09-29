# GSToken

GSToken 是一个轻量级的 Golang 权限认证框架，主要解决：登录认证、权限认证、单点登录、分布式Session会话等一系列权限相关问题。

## 🚀 功能特性

- **多模式登录认证** - 单端登录、多端登录、同端互斥登录、七天内免登录
- **RBAC权限控制** - 基于角色的权限认证，可自定义权限规则
- **分布式Session管理** - 支持Redis集群的分布式会话存储
- **Token风格定制** - 内置6种Token风格，支持自定义Token生成策略
- **踢人下线功能** - 根据账号ID踢人下线、根据Token值踢人下线
- **方法式鉴权** - 优雅的将鉴权与业务代码分离
- **单点登录(SSO)** - 跨应用的统一登录认证
- **多账号体系认证** - 一个系统多套账号分开鉴权

## 📦 安装

```bash
go get -u gs-token
```

## 🎯 快速开始

### 基础使用

```go
package main

import (
    "gs-token"
    "gs-token/config"
)

func main() {
    // 使用默认配置
    cfg := config.DefaultConfig()
    gs := gstoken.New(cfg)
    
    // 或使用Redis存储
    // cfg := config.RedisConfig()
    // gs := gstoken.New(cfg)
}
```

### Token风格

GS-Token 支持6种内置Token风格：

```go
// UUID风格 (默认)
// 示例: 550e8400-e29b-41d4-a716-446655440000

// UUID简化风格 (去掉中划线)
// 示例: 550e8400e29b41d4a716446655440000

// 随机32位字符串
// 示例: a1b2c3d4e5f6789012345678901234ab

// 随机64位字符串
// 示例: a1b2c3d4e5f6789012345678901234ab567890abcdef1234567890abcdef1234

// 随机128位字符串
// 示例: a1b2c3d4e5f6789012345678901234ab567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12

// Tik风格
// 示例: tik_1640995200_a1b2c3d4e5f6789012345678901234ab

// 自定义风格 - 支持用户自定义生成逻辑
// 示例: USER_1640995200_ABCD (业务前缀风格)
```

### 自定义Token生成

```go
// 注册自定义Token生成函数
generator := gs.GetTokenGenerator()

customFunc := func(extra map[string]interface{}) (string, error) {
    // 自定义生成逻辑
    timestamp := time.Now().Unix()
    businessCode := extra["business_code"].(string)
    return fmt.Sprintf("%s_%d", businessCode, timestamp), nil
}

// 注册并使用自定义函数
generator.RegisterCustomFunc(customFunc)
generator.SetStyle(core.StyleCustom)

token, err := generator.Generate(map[string]interface{}{
    "business_code": "USER",
})
```

## 🏗️ 项目结构

```
gs-token/
├── internal/           # 内部模块
│   ├── core/          # 核心接口和类型定义
│   ├── token/         # Token生成器
│   ├── storage/       # 存储适配器
│   ├── auth/          # 认证服务
│   ├── permission/    # 权限管理
│   ├── session/       # 会话管理
│   ├── middleware/    # 中间件
│   └── sso/           # 单点登录
├── config/            # 配置管理
├── examples/          # 示例代码
├── docs/              # 文档
└── gstoken.go         # 主入口文件
```

## 📚 文档

- [架构设计文档](docs/architecture.md)
- [API文档](docs/api.md)
- [配置说明](docs/config.md)
- [示例代码](examples/)

## 🔧 配置

### 默认配置

```go
config := &core.Config{
    TokenExpire:   24 * time.Hour,     // Token过期时间
    RefreshExpire: 7 * 24 * time.Hour, // 刷新Token过期时间
    TokenStyle:    core.StyleUUID,     // Token风格
    LoginMode:     core.MultiLogin,    // 登录模式
    AutoRenew:     true,               // 自动续期
    RememberDays:  7,                  // 记住登录天数
    Storage: core.StorageConfig{
        Type: "memory",                // 存储类型
    },
}
```

### Redis配置

```go
config.Storage.Type = "redis"
config.Redis = core.RedisConfig{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
    PoolSize: 10,
}
```

## 🚦 开发状态

- [x] 项目初始化和基础架构搭建
- [ ] 实现核心认证引擎和基础接口定义
- [ ] 开发Token生成器模块，支持多种Token风格
- [ ] 构建会话管理模块，实现分布式Session存储
- [ ] 实现登录认证模块，支持多种登录模式
- [ ] 开发权限认证模块，实现RBAC权限控制
- [ ] 构建踢人下线功能和会话控制机制
- [ ] 实现方法式鉴权和中间件集成
- [ ] 开发单点登录和多账号体系支持

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！