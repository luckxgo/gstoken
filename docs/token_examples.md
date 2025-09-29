# GS-Token 自定义Token示例文档

## 概述

GS-Token 提供了强大的自定义Token生成功能，支持用户根据业务需求定制Token格式和生成逻辑。本文档展示了各种自定义Token场景的实现方法。

## 基础自定义Token

### 1. 业务前缀Token

适用于需要区分不同业务模块的场景：

```go
businessTokenFunc := func(extra map[string]interface{}) (string, error) {
    businessCode := getStringFromExtra(extra, "business_code", "BIZ")
    timestamp := time.Now().UnixNano()
    return fmt.Sprintf("%s_%d_%d", businessCode, timestamp, timestamp%10000), nil
}

// 使用示例
token, _ := generator.Generate(map[string]interface{}{
    "business_code": "USER",
})
// 输出: USER_1759134792028207000_7000
```

### 2. 编码风格Token

对敏感信息进行简单编码：

```go
encodedTokenFunc := func(extra map[string]interface{}) (string, error) {
    timestamp := time.Now().Unix()
    encoded := ""
    for _, char := range strconv.FormatInt(timestamp, 16) {
        encoded += string(char + 1) // 字符偏移编码
    }
    return fmt.Sprintf("enc_%s", encoded), nil
}

// 输出: enc_79eb5559
```

## 高级自定义Token场景

### 1. JWT风格Token

模拟JWT结构的自定义Token：

```go
jwtStyleFunc := func(extra map[string]interface{}) (string, error) {
    header := base64.StdEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
    
    payload := fmt.Sprintf(`{"sub":"%s","iat":%d,"exp":%d}`, 
        getStringFromExtra(extra, "user_id", "anonymous"),
        time.Now().Unix(),
        time.Now().Add(24*time.Hour).Unix())
    encodedPayload := base64.StdEncoding.EncodeToString([]byte(payload))
    
    signature := generateSimpleSignature(header + "." + encodedPayload)
    
    return fmt.Sprintf("%s.%s.%s", header, encodedPayload, signature), nil
}

// 输出: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwiaWF0IjoxNzU5MTM1MDI1LCJleHAiOjE3NTkyMjE0MjV9.3b3eb3a5e3cd7e31
```

### 2. 基于角色的Token

根据用户角色生成不同格式的Token：

```go
roleBasedFunc := func(extra map[string]interface{}) (string, error) {
    role := getStringFromExtra(extra, "role", "user")
    userID := getStringFromExtra(extra, "user_id", "anonymous")
    
    var prefix string
    switch role {
    case "admin":
        prefix = "ADM"
    case "manager":
        prefix = "MGR"
    case "user":
        prefix = "USR"
    default:
        prefix = "GUE" // Guest
    }
    
    timestamp := time.Now().Unix()
    hash := generateHash(userID + role)
    
    return fmt.Sprintf("%s_%d_%s_%s", prefix, timestamp, hash[:8], generateRandomString(6)), nil
}

// 输出示例:
// admin: ADM_1759135025_672ffdd5_E7OJA0
// user:  USR_1759135025_ef97eb54_R0QLAD
```

### 3. 带校验码的Token

增强Token安全性的校验码机制：

```go
checksumFunc := func(extra map[string]interface{}) (string, error) {
    baseToken := fmt.Sprintf("TOK_%d_%s", time.Now().UnixNano(), generateRandomString(12))
    checksum := calculateChecksum(baseToken)
    return fmt.Sprintf("%s_%s", baseToken, checksum), nil
}

// 校验函数
func validateToken(token string) bool {
    parts := strings.Split(token, "_")
    if len(parts) < 4 {
        return false
    }
    baseToken := strings.Join(parts[:3], "_")
    providedChecksum := parts[3]
    calculatedChecksum := calculateChecksum(baseToken)
    return providedChecksum == calculatedChecksum
}

// 输出: TOK_1759135025502973000_U7A55JW3EJ2W_7e2aaba1
```

### 4. 分层级Token

适用于组织架构复杂的企业应用：

```go
hierarchicalFunc := func(extra map[string]interface{}) (string, error) {
    level := getIntFromExtra(extra, "level", 1)
    department := getStringFromExtra(extra, "department", "default")
    
    levelCode := fmt.Sprintf("L%02d", level)
    deptCode := strings.ToUpper(department[:min(3, len(department))])
    
    timestamp := time.Now().Unix()
    sequence := timestamp % 10000
    
    return fmt.Sprintf("%s_%s_%d_%04d", levelCode, deptCode, timestamp, sequence), nil
}

// 输出示例:
// L01_ENG_1759135025_5025 (1级工程部)
// L05_EXE_1759135025_5025 (5级执行部)
```

### 5. 地理位置Token

基于用户地理位置的Token：

```go
locationFunc := func(extra map[string]interface{}) (string, error) {
    country := getStringFromExtra(extra, "country", "CN")
    city := getStringFromExtra(extra, "city", "Beijing")
    
    geoCode := fmt.Sprintf("%s%s", country, city[:min(3, len(city))])
    timestamp := time.Now().Unix()
    
    return fmt.Sprintf("GEO_%s_%d_%s", geoCode, timestamp, generateRandomString(8)), nil
}

// 输出示例:
// GEO_CNBei_1759135025_ZDNATZH0 (中国北京)
// GEO_USNew_1759135025_KUMIJJ9O (美国纽约)
```

### 6. 时间窗口Token

基于时间窗口的Token，适用于限时访问场景：

```go
timeWindowFunc := func(extra map[string]interface{}) (string, error) {
    windowMinutes := getIntFromExtra(extra, "window_minutes", 60)
    
    now := time.Now()
    windowStart := now.Truncate(time.Duration(windowMinutes) * time.Minute)
    windowID := windowStart.Unix() / int64(windowMinutes*60)
    
    return fmt.Sprintf("WIN_%d_%d_%s", windowID, windowMinutes, generateRandomString(10)), nil
}

// 输出示例:
// WIN_1954594_15_F03MSY15OT (15分钟窗口)
// WIN_488648_60_9HX298VVDA (60分钟窗口)
```

### 7. 设备指纹Token

基于设备特征的Token：

```go
deviceFunc := func(extra map[string]interface{}) (string, error) {
    userAgent := getStringFromExtra(extra, "user_agent", "unknown")
    ip := getStringFromExtra(extra, "ip", "0.0.0.0")
    
    fingerprint := generateHash(userAgent + ip)
    timestamp := time.Now().Unix()
    
    return fmt.Sprintf("DEV_%s_%d_%s", fingerprint[:12], timestamp, generateRandomString(6)), nil
}

// 输出: DEV_d98b0ade90cf_1759135025_9X715Y
```

### 8. 多租户Token

支持多租户架构的Token：

```go
tenantFunc := func(extra map[string]interface{}) (string, error) {
    tenantID := getStringFromExtra(extra, "tenant_id", "default")
    environment := getStringFromExtra(extra, "environment", "prod")
    
    tenantHash := generateHash(tenantID)[:8]
    envCode := map[string]string{
        "dev":  "D",
        "test": "T",
        "prod": "P",
    }[environment]
    if envCode == "" {
        envCode = "U" // Unknown
    }
    
    timestamp := time.Now().Unix()
    
    return fmt.Sprintf("TNT_%s%s_%d_%s", envCode, tenantHash, timestamp, generateRandomString(8)), nil
}

// 输出示例:
// TNT_P22f36260_1759135025_A5MLBLZ8 (生产环境)
// TNT_T1e59214e_1759135025_ANZ7OPKA (测试环境)
```

## Token验证和解析

### 1. 格式验证

使用正则表达式验证Token格式：

```go
// 验证特定格式的Token
pattern := regexp.MustCompile(`^(SESSION|REFRESH|API|TEMP)_[a-zA-Z0-9]+_\d+_[A-Z0-9]{8}$`)
isValid := pattern.MatchString(token)
```

### 2. 信息提取

从Token中提取业务信息：

```go
func extractTokenInfo(token string) map[string]string {
    parts := strings.Split(token, "_")
    if len(parts) < 4 {
        return nil
    }
    
    return map[string]string{
        "type":      parts[0],
        "user_id":   parts[1],
        "timestamp": parts[2],
        "random":    parts[3],
    }
}
```

### 3. 安全性检查

实现Token的安全验证：

```go
func validateTokenSecurity(token, expectedIP string) bool {
    parts := strings.Split(token, "_")
    if len(parts) < 5 {
        return false
    }
    
    // 校验码验证
    baseToken := strings.Join(parts[:4], "_")
    providedChecksum := parts[4]
    calculatedChecksum := calculateChecksum(baseToken)
    
    if providedChecksum != calculatedChecksum {
        return false
    }
    
    // IP验证
    expectedIPHash := generateHash(expectedIP)[:8]
    tokenIPHash := parts[2]
    
    return expectedIPHash == tokenIPHash
}
```

## 性能考虑

### 1. 简单vs复杂自定义函数

- **简单函数**: 直接字符串拼接，性能最佳
- **复杂函数**: 包含哈希计算、编码等操作，性能较低

### 2. 性能优化建议

1. **避免复杂计算**: 在Token生成函数中避免耗时的加密或哈希操作
2. **缓存静态数据**: 对不变的配置信息进行缓存
3. **批量生成**: 对于大量Token生成需求，考虑批量处理
4. **并发安全**: 确保自定义函数是线程安全的

### 3. 内存使用

- Token平均长度影响内存使用
- 大量Token缓存时需要考虑内存限制
- 及时清理不需要的Token引用

## 最佳实践

### 1. Token设计原则

- **唯一性**: 确保Token在有效期内全局唯一
- **不可预测**: 包含足够的随机性
- **信息最小化**: 避免在Token中包含敏感信息
- **易于验证**: 提供高效的验证机制

### 2. 安全建议

- **定期轮换**: 实现Token的定期更新机制
- **访问控制**: 限制Token的使用范围和权限
- **审计日志**: 记录Token的生成和使用情况
- **异常检测**: 监控异常的Token使用模式

### 3. 扩展性考虑

- **版本兼容**: 设计向后兼容的Token格式
- **配置化**: 通过配置文件管理Token生成策略
- **插件化**: 支持动态加载自定义Token生成器
- **监控指标**: 提供Token生成和验证的性能指标

## 辅助函数

```go
// 从extra参数中安全获取字符串值
func getStringFromExtra(extra map[string]interface{}, key, defaultValue string) string {
    if val, ok := extra[key].(string); ok {
        return val
    }
    return defaultValue
}

// 从extra参数中安全获取整数值
func getIntFromExtra(extra map[string]interface{}, key string, defaultValue int) int {
    if val, ok := extra[key].(int); ok {
        return val
    }
    return defaultValue
}

// 生成简单哈希
func generateHash(data string) string {
    hash := md5.Sum([]byte(data))
    return hex.EncodeToString(hash[:])
}

// 计算校验码
func calculateChecksum(data string) string {
    return generateHash(data)[:8]
}

// 生成随机字符串
func generateRandomString(length int) string {
    const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, length)
    for i := range result {
        num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
        result[i] = charset[num.Int64()]
    }
    return string(result)
}
```

通过这些示例和最佳实践，您可以根据具体的业务需求设计和实现适合的自定义Token生成策略。