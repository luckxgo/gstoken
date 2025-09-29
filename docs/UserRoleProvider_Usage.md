# UserRoleProvider 使用指南

## 概述

GSToken 框架将用户角色获取逻辑交给业务系统实现，通过 `UserRoleProvider` 接口让用户可以根据自己的业务需求来定义角色获取策略。

## 接口定义

```go
// UserRoleProvider 用户角色提供者接口（由用户实现）
type UserRoleProvider interface {
    GetUserRoles(ctx context.Context, userID string) ([]Role, error)
}
```

## 使用步骤

### 1. 实现 UserRoleProvider 接口

```go
package main

import (
    "context"
    "gstoken/core"
)

// 自定义用户角色提供者
type MyUserRoleProvider struct {
    // 可以注入数据库连接、缓存等依赖
    db Database
}

func (p *MyUserRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
    // 从数据库、LDAP、或其他系统获取用户角色
    roles, err := p.db.GetUserRoles(userID)
    if err != nil {
        return nil, err
    }
    
    // 转换为框架需要的角色格式
    var result []core.Role
    for _, role := range roles {
        result = append(result, core.Role{
            ID:          role.ID,
            Name:        role.Name,
            Permissions: role.Permissions,
        })
    }
    
    return result, nil
}
```

### 2. 设置角色提供者

```go
// 创建GSToken实例
gs := gstoken.New(config)

// 获取权限服务
permissionService := gs.GetPermissionService()

// 创建并设置自定义角色提供者
roleProvider := &MyUserRoleProvider{db: myDatabase}
permissionService.SetUserRoleProvider(roleProvider)

// 现在可以使用权限检查功能
hasPermission, err := gs.CheckPermission(ctx, userID, "user:read")
hasRole, err := gs.CheckRole(ctx, userID, "admin")
```

## 示例实现

框架提供了一个基于存储的示例实现 `ExampleUserRoleProvider`，可以作为参考：

```go
// 使用示例实现（基于框架存储）
roleProvider := auth.NewExampleUserRoleProvider(gs.GetStorage())
permissionService.SetUserRoleProvider(roleProvider)
```

## 常见实现场景

### 1. 基于数据库的实现

```go
type DatabaseRoleProvider struct {
    db *sql.DB
}

func (p *DatabaseRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
    query := `
        SELECT r.id, r.name, r.permissions 
        FROM roles r 
        JOIN user_roles ur ON r.id = ur.role_id 
        WHERE ur.user_id = ?
    `
    
    rows, err := p.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var roles []core.Role
    for rows.Next() {
        var role core.Role
        var permissionsJSON string
        
        err := rows.Scan(&role.ID, &role.Name, &permissionsJSON)
        if err != nil {
            return nil, err
        }
        
        // 解析权限JSON
        json.Unmarshal([]byte(permissionsJSON), &role.Permissions)
        roles = append(roles, role)
    }
    
    return roles, nil
}
```

### 2. 基于LDAP的实现

```go
type LDAPRoleProvider struct {
    ldapClient *ldap.Conn
}

func (p *LDAPRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
    // LDAP查询用户组
    searchRequest := ldap.NewSearchRequest(
        "ou=groups,dc=company,dc=com",
        ldap.ScopeWholeSubtree,
        ldap.NeverDerefAliases,
        0, 0, false,
        fmt.Sprintf("(member=uid=%s,ou=users,dc=company,dc=com)", userID),
        []string{"cn", "description"},
        nil,
    )
    
    sr, err := p.ldapClient.Search(searchRequest)
    if err != nil {
        return nil, err
    }
    
    var roles []core.Role
    for _, entry := range sr.Entries {
        role := core.Role{
            ID:   entry.GetAttributeValue("cn"),
            Name: entry.GetAttributeValue("description"),
            // 根据LDAP组映射权限
            Permissions: p.mapGroupToPermissions(entry.GetAttributeValue("cn")),
        }
        roles = append(roles, role)
    }
    
    return roles, nil
}
```

### 3. 基于缓存的实现

```go
type CachedRoleProvider struct {
    baseProvider core.UserRoleProvider
    cache        Cache
    cacheTTL     time.Duration
}

func (p *CachedRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
    cacheKey := fmt.Sprintf("user_roles:%s", userID)
    
    // 尝试从缓存获取
    if cached, err := p.cache.Get(cacheKey); err == nil {
        var roles []core.Role
        if json.Unmarshal(cached, &roles) == nil {
            return roles, nil
        }
    }
    
    // 缓存未命中，从基础提供者获取
    roles, err := p.baseProvider.GetUserRoles(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    if data, err := json.Marshal(roles); err == nil {
        p.cache.Set(cacheKey, data, p.cacheTTL)
    }
    
    return roles, nil
}
```

## 注意事项

1. **性能考虑**：角色获取可能会被频繁调用，建议实现缓存机制
2. **错误处理**：妥善处理数据库连接失败、网络超时等异常情况
3. **安全性**：确保角色数据的完整性和准确性
4. **并发安全**：实现应该是线程安全的
5. **上下文支持**：充分利用 context 进行超时控制和取消操作

## 测试建议

```go
func TestMyRoleProvider(t *testing.T) {
    provider := &MyUserRoleProvider{db: testDB}
    
    roles, err := provider.GetUserRoles(context.Background(), "test_user")
    assert.NoError(t, err)
    assert.NotEmpty(t, roles)
    
    // 验证角色内容
    assert.Equal(t, "admin", roles[0].ID)
    assert.Contains(t, roles[0].Permissions, "user:read")
}
```

通过这种设计，GSToken 框架专注于权限验证逻辑，而角色管理完全由业务系统负责，实现了职责分离和高度的灵活性。