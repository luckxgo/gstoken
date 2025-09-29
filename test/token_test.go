package test

import (
	"gstoken"
	"testing"

	"gstoken/config"
	"gstoken/core"
)

// TestBasicTokenGeneration 测试基础Token生成功能
func TestBasicTokenGeneration(t *testing.T) {
	// 使用默认配置创建GSToken实例
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)

	// 测试不同风格的Token生成
	styles := []struct {
		name  string
		style core.TokenStyle
	}{
		{"UUID风格", core.StyleUUID},
		{"UUID简化风格", core.StyleUUIDSimple},
		{"随机32位", core.StyleRandom32},
		{"随机64位", core.StyleRandom64},
		{"随机128位", core.StyleRandom128},
		{"Tik风格", core.StyleTik},
	}

	for _, styleTest := range styles {
		t.Run(styleTest.name, func(t *testing.T) {
			// 创建指定风格的生成器
			generator := gs.GetTokenGenerator()
			generator.SetStyle(styleTest.style)

			token, err := generator.Generate(map[string]interface{}{
				"device": "web",
			})

			if err != nil {
				t.Errorf("生成%s Token失败: %v", styleTest.name, err)
				return
			}

			if token == "" {
				t.Errorf("%s Token不能为空", styleTest.name)
				return
			}

			t.Logf("%s: %s", styleTest.name, token)

			// 验证Token长度
			switch styleTest.style {
			case core.StyleUUID:
				if len(token) != 36 {
					t.Errorf("UUID Token长度应为36，实际为%d", len(token))
				}
			case core.StyleUUIDSimple:
				if len(token) != 32 {
					t.Errorf("UUID简化Token长度应为32，实际为%d", len(token))
				}
			case core.StyleRandom32:
				if len(token) != 32 {
					t.Errorf("随机32位Token长度应为32，实际为%d", len(token))
				}
			case core.StyleRandom64:
				if len(token) != 64 {
					t.Errorf("随机64位Token长度应为64，实际为%d", len(token))
				}
			case core.StyleRandom128:
				if len(token) != 128 {
					t.Errorf("随机128位Token长度应为128，实际为%d", len(token))
				}
			}
		})
	}
}

// TestCustomTokenGeneration 测试自定义Token生成功能
func TestCustomTokenGeneration(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	t.Run("业务前缀Token", func(t *testing.T) {
		businessTokenFunc := func(extra map[string]interface{}) (string, error) {
			businessCode := getStringFromExtra(extra, "business_code", "BIZ")
			timestamp := getTimestamp()
			return businessCode + "_" + timestamp + "_" + generateRandomSuffix(), nil
		}

		generator.RegisterCustomFunc(businessTokenFunc)
		generator.SetStyle(core.StyleCustom)

		token, err := generator.Generate(map[string]interface{}{
			"business_code": "USER",
		})

		if err != nil {
			t.Errorf("生成业务前缀Token失败: %v", err)
			return
		}

		if token == "" {
			t.Error("业务前缀Token不能为空")
			return
		}

		t.Logf("业务前缀Token: %s", token)

		// 验证Token格式
		if len(token) < 10 {
			t.Error("业务前缀Token长度过短")
		}
	})

	t.Run("编码风格Token", func(t *testing.T) {
		encodedTokenFunc := func(extra map[string]interface{}) (string, error) {
			timestamp := getTimestamp()
			encoded := encodeString(timestamp)
			return "enc_" + encoded, nil
		}

		generator.RegisterCustomFunc(encodedTokenFunc)

		token, err := generator.Generate(map[string]interface{}{})

		if err != nil {
			t.Errorf("生成编码风格Token失败: %v", err)
			return
		}

		if token == "" {
			t.Error("编码风格Token不能为空")
			return
		}

		t.Logf("编码风格Token: %s", token)
	})

	t.Run("参数驱动Token", func(t *testing.T) {
		paramTokenFunc := func(extra map[string]interface{}) (string, error) {
			prefix := getStringFromExtra(extra, "prefix", "CUSTOM")
			suffix := getStringFromExtra(extra, "suffix", "DEFAULT")
			return prefix + "_" + suffix, nil
		}

		generator.RegisterCustomFunc(paramTokenFunc)

		token, err := generator.Generate(map[string]interface{}{
			"prefix": "CUSTOM",
			"suffix": "YZ0123456789",
		})

		if err != nil {
			t.Errorf("生成参数驱动Token失败: %v", err)
			return
		}

		if token != "CUSTOM_YZ0123456789" {
			t.Errorf("参数驱动Token格式错误，期望: CUSTOM_YZ0123456789，实际: %s", token)
		}

		t.Logf("参数驱动Token: %s", token)
	})
}

// TestTokenStyleSwitching 测试Token风格切换
func TestTokenStyleSwitching(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	// 测试从自定义风格切换回内置风格
	customFunc := func(extra map[string]interface{}) (string, error) {
		return "CUSTOM_TOKEN", nil
	}

	generator.RegisterCustomFunc(customFunc)
	generator.SetStyle(core.StyleCustom)

	// 生成自定义Token
	customToken, err := generator.Generate(map[string]interface{}{})
	if err != nil {
		t.Errorf("生成自定义Token失败: %v", err)
		return
	}

	if customToken != "CUSTOM_TOKEN" {
		t.Errorf("自定义Token不匹配，期望: CUSTOM_TOKEN，实际: %s", customToken)
	}

	// 切换回UUID风格
	generator.SetStyle(core.StyleUUID)

	uuidToken, err := generator.Generate(map[string]interface{}{})
	if err != nil {
		t.Errorf("切换回UUID风格后生成Token失败: %v", err)
		return
	}

	if len(uuidToken) != 36 {
		t.Errorf("UUID Token长度错误，期望: 36，实际: %d", len(uuidToken))
	}

	t.Logf("自定义Token: %s", customToken)
	t.Logf("UUID Token: %s", uuidToken)
}

// 辅助函数
func getStringFromExtra(extra map[string]interface{}, key, defaultValue string) string {
	if val, ok := extra[key].(string); ok {
		return val
	}
	return defaultValue
}

func getTimestamp() string {
	return "1759134792028207000"
}

func generateRandomSuffix() string {
	return "7000"
}

func encodeString(input string) string {
	return "79eb5559"
}
