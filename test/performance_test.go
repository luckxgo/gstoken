package test

import (
	"gstoken"
	"sync"
	"testing"
	"time"

	"gstoken/config"
	"gstoken/internal/core"
)

// BenchmarkTokenGeneration 基准测试Token生成性能
func BenchmarkTokenGeneration(b *testing.B) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate(map[string]interface{}{
			"test_id": i,
		})
		if err != nil {
			b.Errorf("生成Token失败: %v", err)
		}
	}
}

// BenchmarkConcurrentTokenGeneration 并发Token生成基准测试
func BenchmarkConcurrentTokenGeneration(b *testing.B) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, err := generator.Generate(map[string]interface{}{
				"test_id": i,
			})
			if err != nil {
				b.Errorf("生成Token失败: %v", err)
			}
			i++
		}
	})
}

// BenchmarkTokenStyles 不同Token风格性能基准测试
func BenchmarkTokenStyles(b *testing.B) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)

	styles := []struct {
		name  string
		style core.TokenStyle
	}{
		{"UUID", core.StyleUUID},
		{"UUID简化", core.StyleUUIDSimple},
		{"随机32位", core.StyleRandom32},
		{"随机64位", core.StyleRandom64},
		{"随机128位", core.StyleRandom128},
		{"Tik风格", core.StyleTik},
	}

	for _, styleTest := range styles {
		b.Run(styleTest.name, func(b *testing.B) {
			generator := gs.GetTokenGenerator()
			generator.SetStyle(styleTest.style)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := generator.Generate(map[string]interface{}{})
				if err != nil {
					b.Errorf("生成%s Token失败: %v", styleTest.name, err)
				}
			}
		})
	}
}

// BenchmarkCustomTokens 自定义Token性能基准测试
func BenchmarkCustomTokens(b *testing.B) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	b.Run("简单自定义", func(b *testing.B) {
		simpleFunc := func(extra map[string]interface{}) (string, error) {
			return "SIMPLE_TOKEN", nil
		}

		generator.RegisterCustomFunc(simpleFunc)
		generator.SetStyle(core.StyleCustom)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := generator.Generate(map[string]interface{}{})
			if err != nil {
				b.Errorf("生成简单自定义Token失败: %v", err)
			}
		}
	})

	b.Run("复杂自定义", func(b *testing.B) {
		complexFunc := func(extra map[string]interface{}) (string, error) {
			timestamp := time.Now().Unix()
			userID := "user123"

			hash := 0
			for _, c := range userID {
				hash = hash*31 + int(c)
			}

			return generateComplexToken(timestamp, hash), nil
		}

		generator.RegisterCustomFunc(complexFunc)
		generator.SetStyle(core.StyleCustom)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := generator.Generate(map[string]interface{}{})
			if err != nil {
				b.Errorf("生成复杂自定义Token失败: %v", err)
			}
		}
	})
}

// TestConcurrentSafety 测试并发安全性
func TestConcurrentSafety(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	const goroutines = 100
	const tokensPerGoroutine = 1000

	var wg sync.WaitGroup
	tokens := make(chan string, goroutines*tokensPerGoroutine)
	errors := make(chan error, goroutines*tokensPerGoroutine)

	// 启动多个goroutine并发生成Token
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < tokensPerGoroutine; j++ {
				token, err := generator.Generate(map[string]interface{}{
					"goroutine_id": id,
					"token_id":     j,
				})

				if err != nil {
					errors <- err
					return
				}

				tokens <- token
			}
		}(i)
	}

	wg.Wait()
	close(tokens)
	close(errors)

	// 检查错误
	errorCount := 0
	for err := range errors {
		t.Errorf("并发生成Token出错: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("并发测试中出现%d个错误", errorCount)
		return
	}

	// 检查Token数量
	tokenCount := 0
	tokenSet := make(map[string]bool)

	for token := range tokens {
		tokenCount++
		if tokenSet[token] {
			t.Errorf("发现重复Token: %s", token)
		}
		tokenSet[token] = true
	}

	expectedCount := goroutines * tokensPerGoroutine
	if tokenCount != expectedCount {
		t.Errorf("Token数量不匹配，期望: %d，实际: %d", expectedCount, tokenCount)
	}

	t.Logf("并发安全性测试通过: %d个goroutine生成了%d个唯一Token", goroutines, tokenCount)
}

// TestMemoryUsage 测试内存使用情况
func TestMemoryUsage(t *testing.T) {
	cfg := config.DefaultConfig()
	gs := gstoken.New(cfg)
	generator := gs.GetTokenGenerator()

	const tokenCount = 10000
	tokens := make([]string, 0, tokenCount)

	start := time.Now()

	for i := 0; i < tokenCount; i++ {
		token, err := generator.Generate(map[string]interface{}{
			"index": i,
		})
		if err != nil {
			t.Errorf("生成Token失败: %v", err)
			return
		}
		tokens = append(tokens, token)
	}

	duration := time.Since(start)

	// 计算统计信息
	totalLength := 0
	for _, token := range tokens {
		totalLength += len(token)
	}

	avgLength := float64(totalLength) / float64(len(tokens))
	qps := float64(len(tokens)) / duration.Seconds()

	t.Logf("内存使用测试结果:")
	t.Logf("  生成Token数量: %d", len(tokens))
	t.Logf("  总耗时: %v", duration)
	t.Logf("  平均QPS: %.2f", qps)
	t.Logf("  Token平均长度: %.2f字符", avgLength)
	t.Logf("  总字符数: %d", totalLength)
}

// 辅助函数
func generateComplexToken(timestamp int64, hash int) string {
	return "COMPLEX_" + string(rune(timestamp%1000)) + "_" + string(rune(hash%1000))
}
