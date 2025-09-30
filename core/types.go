package core

import (
	"time"
)

// Token 相关常量
const (
	// Token 生成时的额外参数键
	TokenExtraKeyUserID = "user_id"
	TokenExtraKeyDevice = "device"
	TokenExtraKeyIP     = "ip"
	TokenExtraKeyType   = "type"

	// Token 类型值
	TokenTypeRefresh = "refresh"
	TokenTypeAccess  = "access"

	// Token 额外标识
	TokenFlagRefresh = "refresh"

	// 权限相关常量
	PermissionWildcard = "*"
	
	// 存储类型常量
	StorageTypeRedis    = "redis"
	StorageTypeMemory   = "memory"
	StorageTypeDatabase = "database"
	
	// 默认配置常量
	DefaultKeyPrefix      = "gstoken"
	DefaultRedisAddr      = "localhost:6379"
	DefaultRedisPassword  = ""
	DefaultRedisDB        = 0
	DefaultRedisPoolSize  = 10
	DefaultDatabaseDriver = "mysql"
	DefaultDatabaseHost   = "localhost"
	DefaultDatabasePort   = 3306
	DefaultDatabaseUser   = "root"
	DefaultDatabasePass   = ""
	DefaultDatabaseName   = "gstoken"
	DefaultRememberDays   = 7
)

// 错误消息常量
const (
	ErrMsgHandleLoginMode         = "处理登录模式失败"
	ErrMsgGenerateToken           = "生成Token失败"
	ErrMsgGenerateRefreshToken    = "生成刷新Token失败"
	ErrMsgCreateSession           = "创建会话失败"
	ErrMsgStoreLoginInfo          = "存储登录信息失败"
	ErrMsgStoreUserSessionMap     = "存储用户会话映射失败"
	ErrMsgStoreRefreshToken       = "存储刷新Token失败"
	ErrMsgDeleteSession           = "删除会话失败"
	ErrMsgDeleteLoginInfo         = "删除登录信息失败"
	ErrMsgGetUserSessionKeys      = "获取用户会话键失败"
	ErrMsgGetLoginInfo            = "获取登录信息失败"
	ErrMsgLoginInfoNotExists      = "登录信息不存在"
	ErrMsgStorageDataFormat       = "存储数据格式错误，期望字节数组"
	ErrMsgParseLoginInfo          = "解析登录信息失败"
	ErrMsgGetRefreshTokenInfo     = "获取刷新Token信息失败"
	ErrMsgRefreshTokenNotExists   = "刷新Token不存在"
	ErrMsgParseRefreshTokenInfo   = "解析刷新Token信息失败"
	ErrMsgRefreshTokenExpired     = "刷新Token已过期"
	ErrMsgGenerateNewAccessToken  = "生成新访问Token失败"
	ErrMsgGenerateNewRefreshToken = "生成新刷新Token失败"
	ErrMsgCreateNewSession        = "创建新会话失败"
	ErrMsgStoreNewLoginInfo       = "存储新登录信息失败"
	ErrMsgStoreNewRefreshToken    = "存储新刷新Token失败"
	ErrMsgUserIDEmpty             = "用户ID不能为空"
	ErrMsgRefreshTokenEmpty       = "刷新Token不能为空"

	// 会话相关错误消息
	ErrMsgSessionInfoEmpty        = "会话信息不能为空"
	ErrMsgTokenEmpty              = "Token不能为空"
	ErrMsgStoreSessionData        = "存储会话数据失败"
	ErrMsgStoreUserSessionMapping = "存储用户会话映射失败"
	ErrMsgGetSessionData          = "获取会话数据失败"
	ErrMsgSessionNotExists        = "会话不存在"
	ErrMsgSessionDataFormat       = "会话数据格式错误，期望字节数组"
	ErrMsgParseSessionData        = "解析会话数据失败"
	ErrMsgCheckSessionExists      = "检查会话是否存在失败"
	ErrMsgUpdateSessionData       = "更新会话数据失败"
	ErrMsgDeleteSessionData       = "删除会话数据失败"
	ErrMsgGetUserSessionList      = "获取用户会话列表失败"

	// 认证引擎相关错误消息
	ErrMsgLoginRequestEmpty = "登录请求不能为空"
	ErrMsgTokenExpired      = "Token已过期"
	ErrMsgGetSessionInfo    = "获取会话信息失败"
	ErrMsgPermissionEmpty   = "权限标识不能为空"

	// 权限服务相关错误消息
	ErrMsgRoleIDEmpty           = "角色ID不能为空"
	ErrMsgUserRoleProviderEmpty = "用户角色提供者未设置，请调用 SetUserRoleProvider 方法"
	ErrMsgGetUserRoles          = "获取用户角色失败"
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
