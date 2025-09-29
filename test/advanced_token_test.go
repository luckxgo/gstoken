package test

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"gstoken"
	"regexp"
	"strings"
	"testing"
	"time"

	"gstoken/config"
	"gstoken/core"
)

// TestJWTStyleToken 测试JWT风格Token
func TestJWTStyleToken(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

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

	generator.RegisterCustomFunc(jwtStyleFunc)
	generator.SetStyle(core.StyleCustom)

	token, err := generator.Generate(map[string]interface{}{
		"user_id": "user123",
	})

	if err != nil {
		t.Errorf("生成JWT风格Token失败: %v", err)
		return
	}

	// 验证JWT格式
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("JWT Token格式错误，应包含3个部分，实际: %d", len(parts))
		return
	}

	t.Logf("JWT风格Token: %s", token)
}

// TestRoleBasedToken 测试基于角色的Token
func TestRoleBasedToken(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

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
			prefix = "GUE"
		}

		timestamp := time.Now().Unix()
		hash := generateHash(userID + role)

		return fmt.Sprintf("%s_%d_%s_%s", prefix, timestamp, hash[:8], generateRandomString(6)), nil
	}

	generator.RegisterCustomFunc(roleBasedFunc)
	generator.SetStyle(core.StyleCustom)

	roles := []string{"admin", "manager", "user", "guest"}

	for _, role := range roles {
		t.Run(fmt.Sprintf("角色_%s", role), func(t *testing.T) {
			token, err := generator.Generate(map[string]interface{}{
				"role":    role,
				"user_id": "testuser",
			})

			if err != nil {
				t.Errorf("生成%s角色Token失败: %v", role, err)
				return
			}

			// 验证Token前缀
			expectedPrefix := map[string]string{
				"admin":   "ADM",
				"manager": "MGR",
				"user":    "USR",
				"guest":   "GUE",
			}[role]

			if !strings.HasPrefix(token, expectedPrefix) {
				t.Errorf("%s角色Token前缀错误，期望: %s，实际Token: %s", role, expectedPrefix, token)
			}

			t.Logf("%s角色Token: %s", role, token)
		})
	}
}

// TestChecksumToken 测试带校验码的Token
func TestChecksumToken(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	checksumFunc := func(extra map[string]interface{}) (string, error) {
		baseToken := fmt.Sprintf("TOK_%d_%s", time.Now().UnixNano(), generateRandomString(12))
		checksum := calculateChecksum(baseToken)
		return fmt.Sprintf("%s_%s", baseToken, checksum), nil
	}

	generator.RegisterCustomFunc(checksumFunc)
	generator.SetStyle(core.StyleCustom)

	token, err := generator.Generate(map[string]interface{}{})

	if err != nil {
		t.Errorf("生成校验码Token失败: %v", err)
		return
	}

	// 验证校验码
	parts := strings.Split(token, "_")
	if len(parts) < 4 {
		t.Errorf("校验码Token格式错误，部分数量: %d", len(parts))
		return
	}

	baseToken := strings.Join(parts[:3], "_")
	providedChecksum := parts[3]
	calculatedChecksum := calculateChecksum(baseToken)

	if providedChecksum != calculatedChecksum {
		t.Errorf("校验码验证失败，提供: %s，计算: %s", providedChecksum, calculatedChecksum)
	}

	t.Logf("校验码Token: %s", token)
	t.Logf("校验结果: 通过")
}

// TestTokenValidation 测试Token验证功能
func TestTokenValidation(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	// 创建带验证的自定义Token生成器
	validationFunc := func(extra map[string]interface{}) (string, error) {
		tokenType := getStringFromExtra(extra, "type", "session")
		userID := getStringFromExtra(extra, "user_id", "anonymous")

		timestamp := time.Now().Unix()

		return fmt.Sprintf("%s_%s_%d_%s",
			strings.ToUpper(tokenType), userID, timestamp, generateRandomString(8)), nil
	}

	generator.RegisterCustomFunc(validationFunc)
	generator.SetStyle(core.StyleCustom)

	tokenTypes := []string{"session", "refresh", "api", "temp"}
	var tokens []string

	// 生成不同类型的Token
	for _, tokenType := range tokenTypes {
		token, err := generator.Generate(map[string]interface{}{
			"type":    tokenType,
			"user_id": "user123",
		})

		if err != nil {
			t.Errorf("生成%s Token失败: %v", tokenType, err)
			continue
		}

		tokens = append(tokens, token)
		t.Logf("%s Token: %s", tokenType, token)
	}

	// 验证Token格式
	pattern := regexp.MustCompile(`^(SESSION|REFRESH|API|TEMP)_[a-zA-Z0-9]+_\d+_[A-Z0-9]{8}$`)

	for i, token := range tokens {
		isValid := pattern.MatchString(token)
		if !isValid {
			t.Errorf("Token格式验证失败: %s", token)
		} else {
			t.Logf("Token格式验证通过: %s", tokenTypes[i])
		}
	}
}

// TestHierarchicalToken 测试分层级Token
func TestHierarchicalToken(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	hierarchicalFunc := func(extra map[string]interface{}) (string, error) {
		level := getIntFromExtra(extra, "level", 1)
		department := getStringFromExtra(extra, "department", "default")

		levelCode := fmt.Sprintf("L%02d", level)
		deptCode := strings.ToUpper(department[:min(3, len(department))])

		timestamp := time.Now().Unix()
		sequence := timestamp % 10000

		return fmt.Sprintf("%s_%s_%d_%04d", levelCode, deptCode, timestamp, sequence), nil
	}

	generator.RegisterCustomFunc(hierarchicalFunc)
	generator.SetStyle(core.StyleCustom)

	testCases := []map[string]interface{}{
		{"level": 1, "department": "engineering"},
		{"level": 2, "department": "marketing"},
		{"level": 3, "department": "sales"},
		{"level": 5, "department": "executive"},
	}

	for _, testCase := range testCases {
		token, err := generator.Generate(testCase)
		if err != nil {
			t.Errorf("生成层级Token失败: %v", err)
			continue
		}

		// 验证Token格式
		parts := strings.Split(token, "_")
		if len(parts) != 4 {
			t.Errorf("层级Token格式错误: %s", token)
			continue
		}

		expectedLevel := fmt.Sprintf("L%02d", testCase["level"])
		if parts[0] != expectedLevel {
			t.Errorf("层级代码错误，期望: %s，实际: %s", expectedLevel, parts[0])
		}

		t.Logf("层级%d-%s Token: %s", testCase["level"], testCase["department"], token)
	}
}

// 辅助函数
func generateSimpleSignature(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func generateHash(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func calculateChecksum(data string) string {
	return generateHash(data)[:8]
}

func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	timestamp := time.Now().UnixNano()

	for i := range result {
		result[i] = charset[(timestamp+int64(i))%int64(len(charset))]
	}
	return string(result)
}

func getIntFromExtra(extra map[string]interface{}, key string, defaultValue int) int {
	if val, ok := extra[key].(int); ok {
		return val
	}
	return defaultValue
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
