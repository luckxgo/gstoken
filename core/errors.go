package core

import "errors"

// 认证相关错误
var (
	// ErrTokenNotFound Token未找到
	ErrTokenNotFound = errors.New("token not found")

	// ErrTokenExpired Token已过期
	ErrTokenExpired = errors.New("token expired")

	// ErrTokenInvalid Token无效
	ErrTokenInvalid = errors.New("token invalid")

	// ErrUserNotLogin 用户未登录
	ErrUserNotLogin = errors.New("user not login")

	// ErrPermissionDenied 权限不足
	ErrPermissionDenied = errors.New("permission denied")

	// ErrRoleNotFound 角色未找到
	ErrRoleNotFound = errors.New("role not found")

	// ErrUserNotFound 用户未找到
	ErrUserNotFound = errors.New("user not found")

	// ErrSessionNotFound 会话未找到
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired 会话已过期
	ErrSessionExpired = errors.New("session expired")

	// ErrRefreshTokenExpired 刷新Token已过期
	ErrRefreshTokenExpired = errors.New("refresh token expired")

	// ErrRefreshTokenInvalid 刷新Token无效
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
)

// 配置相关错误
var (
	// ErrConfigInvalid 配置无效
	ErrConfigInvalid = errors.New("config invalid")

	// ErrStorageNotConfigured 存储未配置
	ErrStorageNotConfigured = errors.New("storage not configured")

	// ErrTokenGeneratorNotConfigured Token生成器未配置
	ErrTokenGeneratorNotConfigured = errors.New("token generator not configured")
)

// 业务逻辑错误
var (
	// ErrUserAlreadyLogin 用户已登录
	ErrUserAlreadyLogin = errors.New("user already login")

	// ErrLoginModeNotSupported 登录模式不支持
	ErrLoginModeNotSupported = errors.New("login mode not supported")

	// ErrDeviceNotSupported 设备类型不支持
	ErrDeviceNotSupported = errors.New("device not supported")
)
