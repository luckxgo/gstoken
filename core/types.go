package core

import (
	"time"
)

// TokenStyle Token风格枚举
type TokenStyle int

const (
	StyleUUID       TokenStyle = iota // UUID风格 (默认风格)
	StyleUUIDSimple                   // UUID风格, 只不过去掉了中划线
	StyleRandom32                     // 随机32位字符串
	StyleRandom64                     // 随机64位字符串
	StyleRandom128                    // 随机128位字符串
	StyleTik                          // tik风格
	StyleCustom                       // 自定义风格
)

// CustomTokenFunc 自定义Token生成函数类型
type CustomTokenFunc func(extra map[string]interface{}) (string, error)

// LoginMode 登录模式
type LoginMode int

const (
	SingleLogin LoginMode = iota // 单端登录
	MultiLogin                   // 多端登录
	MutexLogin                   // 同端互斥登录
)

// LoginRequest 登录请求
type LoginRequest struct {
	UserID string                 `json:"user_id"`
	Device string                 `json:"device,omitempty"`
	IP     string                 `json:"ip,omitempty"`
	Extra  map[string]interface{} `json:"extra,omitempty"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpireTime   time.Time `json:"expire_time"`
	UserInfo     *UserInfo `json:"user_info"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       string                 `json:"id"`
	Username string                 `json:"username"`
	Roles    []string               `json:"roles"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

// TokenInfo Token信息
type TokenInfo struct {
	UserID     string                 `json:"user_id"`
	ExpireTime time.Time              `json:"expire_time"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// LoginInfo 登录信息
type LoginInfo struct {
	UserID     string                 `json:"user_id"`
	Token      string                 `json:"token"`
	Device     string                 `json:"device"`
	IP         string                 `json:"ip"`
	LoginTime  time.Time              `json:"login_time"`
	LastAccess time.Time              `json:"last_access"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// RefreshTokenInfo 刷新Token信息
type RefreshTokenInfo struct {
	RefreshToken string                 `json:"refresh_token"`
	UserID       string                 `json:"user_id"`
	Device       string                 `json:"device"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	Extra        map[string]interface{} `json:"extra"`
}

// Role 角色信息
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// Session 会话信息
type Session struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Token      string                 `json:"token"`
	Device     string                 `json:"device"`
	IP         string                 `json:"ip"`
	LoginTime  time.Time              `json:"login_time"`
	LastAccess time.Time              `json:"last_access"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// Config 配置信息
type Config struct {
	// Token配置
	TokenExpire   time.Duration `json:"token_expire"`
	RefreshExpire time.Duration `json:"refresh_expire"`
	TokenStyle    TokenStyle    `json:"token_style"`

	// 登录配置
	LoginMode    LoginMode `json:"login_mode"`
	AutoRenew    bool      `json:"auto_renew"`    // 自动续期
	RememberDays int       `json:"remember_days"` // 记住登录天数

	// 存储配置
	Storage  StorageConfig  `json:"storage"`
	Redis    RedisConfig    `json:"redis"`
	Database DatabaseConfig `json:"database"`

	// 键前缀配置
	KeyPrefix string `json:"key_prefix"` // 存储键前缀，默认为 "gstoken"

	// 用户角色提供者（不序列化到JSON）
	UserRoleProvider UserRoleProvider `json:"-"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type string `json:"type"` // redis, memory, database
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}
