# GS-Token 架构设计文档

## 1. 系统架构概览

### 1.1 整体架构图
```
┌─────────────────────────────────────────────────────────────┐
│                    GS-Token 权限认证框架                      │
├─────────────────────────────────────────────────────────────┤
│  应用层 (Application Layer)                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ HTTP中间件   │ │ 方法式鉴权   │ │ SSO客户端   │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
├─────────────────────────────────────────────────────────────┤
│  服务层 (Service Layer)                                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ 认证服务     │ │ 权限服务     │ │ 会话服务     │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
├─────────────────────────────────────────────────────────────┤
│  核心层 (Core Layer)                                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ Token管理器  │ │ 权限引擎     │ │ 会话管理器   │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
├─────────────────────────────────────────────────────────────┤
│  存储层 (Storage Layer)                                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ Redis缓存   │ │ 数据库存储   │ │ 内存缓存     │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 模块依赖关系
```
Application Layer
    ↓ depends on
Service Layer
    ↓ depends on  
Core Layer
    ↓ depends on
Storage Layer
```

## 2. 核心模块设计

### 2.1 Core 模块 (核心引擎)
```go
// 核心接口定义
type AuthEngine interface {
    Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
    Logout(ctx context.Context, token string) error
    Verify(ctx context.Context, token string) (*UserInfo, error)
    CheckPermission(ctx context.Context, userID string, permission string) (bool, error)
}

// 配置管理
type Config struct {
    TokenExpire     time.Duration
    RefreshExpire   time.Duration
    TokenStyle      TokenStyle
    Storage         StorageConfig
    Redis           RedisConfig
    Database        DatabaseConfig
}
```

### 2.2 Token 模块 (Token管理)
```go
// Token生成器接口
type TokenGenerator interface {
    Generate(userID string, extra map[string]interface{}) (string, error)
    Parse(token string) (*TokenInfo, error)
    Refresh(token string) (string, error)
}

// 支持的Token风格
type TokenStyle int
const (
    StyleUUID TokenStyle = iota    // UUID风格 (默认风格)
    StyleUUIDSimple                // UUID风格, 只不过去掉了中划线
    StyleRandom32                  // 随机32位字符串
    StyleRandom64                  // 随机64位字符串
    StyleRandom128                 // 随机128位字符串
    StyleTik                       // tik风格
)
```

### 2.3 Auth 模块 (登录认证)
```go
// 登录模式
type LoginMode int
const (
    SingleLogin    LoginMode = iota  // 单端登录
    MultiLogin                       // 多端登录
    MutexLogin                       // 同端互斥登录
)

// 认证服务接口
type AuthService interface {
    Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
    Logout(ctx context.Context, token string) error
    LogoutByUserID(ctx context.Context, userID string) error
    GetLoginInfo(ctx context.Context, token string) (*LoginInfo, error)
}
```

### 2.4 Permission 模块 (权限管理)
```go
// RBAC权限模型
type Role struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Permissions []string `json:"permissions"`
}

type User struct {
    ID    string   `json:"id"`
    Roles []string `json:"roles"`
}

// 权限服务接口
type PermissionService interface {
    CheckPermission(ctx context.Context, userID, permission string) (bool, error)
    GetUserRoles(ctx context.Context, userID string) ([]Role, error)
    AssignRole(ctx context.Context, userID, roleID string) error
    RevokeRole(ctx context.Context, userID, roleID string) error
}
```

### 2.5 Session 模块 (会话管理)
```go
// 会话信息
type Session struct {
    ID         string                 `json:"id"`
    UserID     string                 `json:"user_id"`
    Token      string                 `json:"token"`
    Device     string                 `json:"device"`
    IP         string                 `json:"ip"`
    LoginTime  time.Time             `json:"login_time"`
    LastAccess time.Time             `json:"last_access"`
    Extra      map[string]interface{} `json:"extra"`
}

// 会话服务接口
type SessionService interface {
    CreateSession(ctx context.Context, session *Session) error
    GetSession(ctx context.Context, token string) (*Session, error)
    UpdateSession(ctx context.Context, session *Session) error
    DeleteSession(ctx context.Context, token string) error
    KickOut(ctx context.Context, userID string) error
    KickOutByToken(ctx context.Context, token string) error
}
```

## 3. 存储层设计

### 3.1 存储适配器接口
```go
// 统一存储接口
type Storage interface {
    Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
    Get(ctx context.Context, key string) (interface{}, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Keys(ctx context.Context, pattern string) ([]string, error)
}

// Redis存储实现
type RedisStorage struct {
    client *redis.Client
}

// 内存存储实现
type MemoryStorage struct {
    data sync.Map
}
```

### 3.2 数据存储策略
- **热数据**: 存储在Redis中，如活跃会话、Token信息
- **温数据**: 存储在内存缓存中，如权限配置、角色信息
- **冷数据**: 存储在数据库中，如用户信息、历史记录

## 4. 中间件集成

### 4.1 HTTP中间件
```go
// Gin中间件
func GinAuthMiddleware(auth *AuthEngine) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        if token == "" {
            c.JSON(401, gin.H{"error": "未授权"})
            c.Abort()
            return
        }
        
        userInfo, err := auth.Verify(c, token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Token无效"})
            c.Abort()
            return
        }
        
        c.Set("user", userInfo)
        c.Next()
    }
}
```

### 4.2 方法式鉴权
```go
// 权限注解
type RequirePermission struct {
    Permission string
}

// 鉴权装饰器
func WithAuth(auth *AuthEngine, permission string) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // 权限验证逻辑
            if !checkPermission(r, permission) {
                http.Error(w, "权限不足", http.StatusForbidden)
                return
            }
            next(w, r)
        }
    }
}
```

## 5. 单点登录(SSO)设计

### 5.1 SSO架构
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   应用A     │    │   SSO服务   │    │   应用B     │
│             │    │             │    │             │
│ ┌─────────┐ │    │ ┌─────────┐ │    │ ┌─────────┐ │
│ │客户端SDK│ │◄──►│ │认证中心 │ │◄──►│ │客户端SDK│ │
│ └─────────┘ │    │ └─────────┘ │    │ └─────────┘ │
└─────────────┘    └─────────────┘    └─────────────┘
```

### 5.2 SSO流程
1. 用户访问应用A，重定向到SSO认证中心
2. 用户在认证中心完成登录
3. 认证中心生成票据(Ticket)并重定向回应用A
4. 应用A使用票据向认证中心验证用户身份
5. 用户访问应用B时，检测到已登录状态，自动完成认证

## 6. 多账号体系设计

### 6.1 账号体系隔离
```go
// 账号体系配置
type AccountSystem struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Config   Config `json:"config"`
    Storage  Storage `json:"-"`
}

// 多账号管理器
type MultiAccountManager struct {
    systems map[string]*AccountSystem
    default string
}
```

### 6.2 跨体系权限映射
```go
// 权限映射配置
type PermissionMapping struct {
    SourceSystem string `json:"source_system"`
    TargetSystem string `json:"target_system"`
    Mappings     map[string]string `json:"mappings"`
}
```

## 7. 性能优化策略

### 7.1 缓存策略
- **L1缓存**: 进程内存缓存，存储热点数据
- **L2缓存**: Redis缓存，存储会话和权限数据
- **缓存更新**: 采用Write-Through策略保证数据一致性

### 7.2 并发控制
- 使用读写锁保护共享资源
- 连接池管理Redis连接
- 异步处理非关键路径操作

### 7.3 监控指标
- Token生成/验证QPS
- 权限检查响应时间
- 缓存命中率
- 错误率统计

## 8. 安全设计

### 8.1 Token安全
- Token加密存储
- 定期轮换密钥
- 防重放攻击机制

### 8.2 权限安全
- 最小权限原则
- 权限继承控制
- 敏感操作审计

### 8.3 会话安全
- 会话超时机制
- 异地登录检测
- 可疑行为监控

## 9. 扩展性设计

### 9.1 插件机制
```go
// 插件接口
type Plugin interface {
    Name() string
    Init(config map[string]interface{}) error
    Execute(ctx context.Context, data interface{}) (interface{}, error)
}
```

### 9.2 事件系统
```go
// 事件类型
type EventType string
const (
    EventLogin  EventType = "login"
    EventLogout EventType = "logout"
    EventKickOut EventType = "kickout"
)

// 事件监听器
type EventListener interface {
    OnEvent(ctx context.Context, event *Event) error
}
```

## 10. 部署架构

### 10.1 单机部署
```
┌─────────────────────────────────────┐
│           应用服务器                 │
│  ┌─────────────┐ ┌─────────────┐   │
│  │ Web应用     │ │ GS-Token    │   │
│  └─────────────┘ └─────────────┘   │
│  ┌─────────────┐ ┌─────────────┐   │
│  │ Redis       │ │ MySQL       │   │
│  └─────────────┘ └─────────────┘   │
└─────────────────────────────────────┘
```

### 10.2 分布式部署
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Web服务1   │    │  Web服务2   │    │  Web服务N   │
│ ┌─────────┐ │    │ ┌─────────┐ │    │ ┌─────────┐ │
│ │GS-Token │ │    │ │GS-Token │ │    │ │GS-Token │ │
│ └─────────┘ │    │ └─────────┘ │    │ └─────────┘ │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
              ┌─────────────────────────┐
              │     Redis集群           │
              │ ┌─────┐ ┌─────┐ ┌─────┐ │
              │ │节点1│ │节点2│ │节点3│ │
              │ └─────┘ └─────┘ └─────┘ │
              └─────────────────────────┘
                           │
              ┌─────────────────────────┐
              │     数据库集群           │
              │ ┌─────┐ ┌─────┐ ┌─────┐ │
              │ │主库 │ │从库1│ │从库2│ │
              │ └─────┘ └─────┘ └─────┘ │
              └─────────────────────────┘
```

这个架构设计文档涵盖了 GS-Token 框架的所有核心组件和设计考虑，为后续的开发实施提供了详细的技术指导。