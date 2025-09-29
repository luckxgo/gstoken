package token

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"gstoken/core"

	"github.com/google/uuid"
)

// Generator Token生成器实现
type Generator struct {
	style      core.TokenStyle
	customFunc core.CustomTokenFunc
}

// NewGenerator 创建Token生成器
func NewGenerator(style core.TokenStyle) *Generator {
	return &Generator{
		style: style,
	}
}

// Generate 生成Token
func (g *Generator) Generate(extra map[string]interface{}) (string, error) {
	switch g.style {
	case core.StyleUUID:
		return g.generateUUID()
	case core.StyleUUIDSimple:
		return g.generateUUIDSimple()
	case core.StyleRandom32:
		return g.generateRandom(32)
	case core.StyleRandom64:
		return g.generateRandom(64)
	case core.StyleRandom128:
		return g.generateRandom(128)
	case core.StyleTik:
		return g.generateTik()
	case core.StyleCustom:
		return g.generateCustom(extra)
	default:
		return g.generateUUID()
	}
}

// Parse 解析Token
func (g *Generator) Parse(token string) (*core.TokenInfo, error) {
	// 这里简化实现，实际应该根据Token风格进行不同的解析
	return &core.TokenInfo{
		UserID:     "", // 需要从存储中获取
		ExpireTime: time.Now().Add(24 * time.Hour),
		Extra:      make(map[string]interface{}),
	}, nil
}

// Refresh 刷新Token
func (g *Generator) Refresh(token string) (string, error) {
	// 解析原Token获取用户信息
	tokenInfo, err := g.Parse(token)
	if err != nil {
		return "", err
	}

	// 生成新Token
	return g.Generate(tokenInfo.Extra)
}

// generateUUID 生成UUID风格Token
func (g *Generator) generateUUID() (string, error) {
	id := uuid.New()
	return id.String(), nil
}

// generateUUIDSimple 生成简化UUID风格Token (去掉中划线)
func (g *Generator) generateUUIDSimple() (string, error) {
	id := uuid.New()
	return strings.ReplaceAll(id.String(), "-", ""), nil
}

// generateRandom 生成随机字符串Token
func (g *Generator) generateRandom(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateTik 生成tik风格Token
func (g *Generator) generateTik() (string, error) {
	// tik风格: tik_ + 时间戳 + _ + 随机字符串
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	randomStr := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("tik_%d_%s", timestamp, randomStr), nil
}

// generateCustom 使用自定义函数生成Token
func (g *Generator) generateCustom(extra map[string]interface{}) (string, error) {
	if g.customFunc == nil {
		return "", fmt.Errorf("自定义Token生成函数未注册")
	}
	return g.customFunc(extra)
}

// RegisterCustomFunc 注册自定义Token生成函数
func (g *Generator) RegisterCustomFunc(fn core.CustomTokenFunc) error {
	if fn == nil {
		return fmt.Errorf("自定义函数不能为空")
	}
	g.customFunc = fn
	return nil
}

// SetStyle 设置Token风格
func (g *Generator) SetStyle(style core.TokenStyle) error {
	g.style = style
	return nil
}
