# GSToken 示例代码

本目录包含了 GSToken 的各种使用示例，帮助您快速上手和理解不同的功能特性。

## 示例列表

### 1. 基础认证示例 (`basic/main.go`)
演示 GSToken 的基本功能：
- 用户登录和登出
- Token 验证
- 获取登录信息
- 检查登录状态

**运行方式：**
```bash
cd examples/basic
go run main.go
```

### 2. Gin框架集成 (`gin/main.go`)
展示如何在 Gin Web 框架中集成 GSToken：
- 认证中间件
- 登录接口
- 受保护的API接口
- CORS 支持
- 错误处理

**运行方式：**
```bash
cd examples/gin
go run main.go
```

然后访问：
- `http://localhost:8080/api/public/health` - 健康检查
- `POST http://localhost:8080/api/public/login` - 用户登录
- `GET http://localhost:8080/api/auth/profile` - 获取用户信息（需要Token）

### 3. Redis存储示例 (`redis/main.go`)
演示使用 Redis 作为存储后端：
- Redis 配置
- 多用户登录
- 会话持久化
- Token 刷新
- 批量登出

**运行前准备：**
```bash
# 启动 Redis 服务器
redis-server

# 或使用 Docker
docker run -d -p 6379:6379 redis:alpine
```

**运行方式：**
```bash
cd examples/redis
go run main.go
```

### 3.1 Redis 集群示例 (`redis_cluster/main.go`)
演示使用 Redis Cluster（需先准备本地或远程集群）：
- 集群连接参数
- 键扫描（采用 SCAN，遍历主分片）

**运行前准备：**
请确保已有集群环境，示例默认使用 `localhost:7001,7002,7003`，可按需修改。

**运行方式：**
```bash
go run examples/redis_cluster/main.go
```

### 4. 权限控制示例 (`permission/main.go`)
展示完整的 RBAC 权限控制系统：
- 用户角色管理
- 权限检查
- 角色检查
- 业务场景权限验证
- 中间件权限控制

**运行方式：**
```bash
cd examples/permission
go run main.go
```

### 5. 自定义Token生成 (`custom_token/main.go`)
演示各种 Token 生成策略：
- 内置 Token 风格
- 自定义 Token 生成函数
- 业务相关的 Token 格式
- Token 解析和验证

**运行方式：**
```bash
cd examples/custom_token
go run main.go
```

## 通用依赖

所有示例都需要以下依赖：

```bash
go mod tidy
```

如果您在项目根目录，可以直接运行：
```bash
# 基础示例
go run examples/basic/main.go

# Gin 示例
go run examples/gin/main.go

# Redis 示例（需要先启动 Redis）
go run examples/redis/main.go

# Redis 集群示例（需要准备 Redis Cluster）
go run examples/redis_cluster/main.go

# 权限控制示例
go run examples/permission/main.go

# 自定义Token示例
go run examples/custom_token/main.go
```

## 示例说明

### 配置选项
每个示例都展示了不同的配置选项：
- **存储类型**：内存存储 vs Redis 存储
- **登录模式**：单端登录 vs 多端登录
- **Token 过期**：访问Token和刷新Token的过期时间
- **自动续期**：是否开启Token自动续期

### 错误处理
所有示例都包含了完整的错误处理，展示了：
- 登录失败处理
- Token 验证失败处理
- 权限不足处理
- 存储连接失败处理

### 最佳实践
示例中体现的最佳实践：
- 使用上下文 (Context) 传递请求信息
- 合理的错误处理和日志记录
- 安全的Token传递方式
- 清晰的权限控制逻辑

## 扩展示例

您可以基于这些示例创建更复杂的应用：
- 结合数据库的用户管理系统
- 微服务架构中的统一认证
- 移动应用的Token管理
- 企业级的权限管理系统

## 问题反馈

如果您在运行示例时遇到问题，请：
1. 检查 Go 版本（建议 1.19+）
2. 确保依赖已正确安装
3. 对于 Redis 示例，确保 Redis 服务正在运行
4. 查看错误日志获取详细信息

更多问题请提交 Issue 或查看项目文档。