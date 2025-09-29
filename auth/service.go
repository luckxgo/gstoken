package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gstoken/core"
)

// Service 认证服务实现
type Service struct {
	storage        core.Storage
	tokenGenerator core.TokenGenerator
	sessionService core.SessionService
	config         *core.Config
	keyService     *core.KeyService
}

// NewAuthService 创建新的认证服务
func NewAuthService(storage core.Storage, tokenGenerator core.TokenGenerator, sessionService core.SessionService, config *core.Config, keyService *core.KeyService) core.AuthService {
	return &Service{
		storage:        storage,
		tokenGenerator: tokenGenerator,
		sessionService: sessionService,
		config:         config,
		keyService:     keyService,
	}
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, req *core.LoginRequest) (*core.LoginResponse, error) {
	// 处理登录模式
	if err := s.handleLoginMode(ctx, req); err != nil {
		return nil, fmt.Errorf("处理登录模式失败: %w", err)
	}

	// 生成Token
	tokenExtra := map[string]interface{}{
		"user_id": req.UserID,
		"device":  req.Device,
		"ip":      req.IP,
	}

	// 合并额外参数
	for k, v := range req.Extra {
		tokenExtra[k] = v
	}

	token, err := s.tokenGenerator.Generate(tokenExtra)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	// 生成刷新Token（如果配置了）
	var refreshToken string
	if s.config.RefreshExpire > 0 {
		refreshTokenExtra := map[string]interface{}{
			"user_id": req.UserID,
			"type":    "refresh",
		}
		refreshToken, err = s.tokenGenerator.Generate(refreshTokenExtra)
		if err != nil {
			return nil, fmt.Errorf("生成刷新Token失败: %w", err)
		}
	}

	// 创建会话
	now := time.Now()
	session := &core.Session{
		ID:         token,
		UserID:     req.UserID,
		Token:      token,
		Device:     req.Device,
		IP:         req.IP,
		LoginTime:  now,
		LastAccess: now,
		Extra:      req.Extra,
	}

	if err := s.sessionService.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	// 存储登录信息
	loginInfo := &core.LoginInfo{
		UserID:     req.UserID,
		Token:      token,
		Device:     req.Device,
		IP:         req.IP,
		LoginTime:  now,
		LastAccess: now,
	}

	if err := s.storeLoginInfo(ctx, token, loginInfo); err != nil {
		return nil, fmt.Errorf("存储登录信息失败: %w", err)
	}

	// 创建用户会话映射，用于根据userID查找Token
	if err := s.storeUserSessionMapping(ctx, req.UserID, token); err != nil {
		return nil, fmt.Errorf("存储用户会话映射失败: %w", err)
	}

	// 存储刷新Token（如果生成了）
	if refreshToken != "" {
		refreshInfo := &core.RefreshTokenInfo{
			RefreshToken: refreshToken,
			UserID:       req.UserID,
			Device:       req.Device,
			CreatedAt:    now,
			ExpiresAt:    now.Add(s.config.RefreshExpire),
			Extra:        req.Extra,
		}
		if err := s.storeRefreshToken(ctx, refreshToken, refreshInfo); err != nil {
			return nil, fmt.Errorf("存储刷新Token失败: %w", err)
		}
	}

	// 构造响应
	response := &core.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpireTime:   now.Add(s.config.TokenExpire),
		UserInfo: &core.UserInfo{
			ID:       req.UserID,
			Username: req.UserID, // 简化处理
			Extra:    req.Extra,
		},
	}

	return response, nil
}

// Logout 用户登出
func (s *Service) Logout(ctx context.Context, token string) error {
	// 获取登录信息以便删除用户会话映射
	loginInfo, err := s.GetLoginInfo(ctx, token)
	if err == nil && loginInfo != nil {
		// 删除用户会话映射
		userSessionKey := s.keyService.UserSessionKey(loginInfo.UserID, token)
		s.storage.Delete(ctx, userSessionKey)
	}

	// 删除会话
	if err := s.sessionService.DeleteSession(ctx, token); err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}

	// 删除登录信息
	loginKey := s.keyService.LoginInfoKey(token)
	if err := s.storage.Delete(ctx, loginKey); err != nil {
		return fmt.Errorf("删除登录信息失败: %w", err)
	}

	return nil
}

// LogoutByUserID 根据用户ID登出所有会话
func (s *Service) LogoutByUserID(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("用户ID不能为空")
	}

	// 获取用户的所有会话Token
	userSessionPattern := s.keyService.UserSessionPattern(userID)
	sessionKeys, err := s.storage.Keys(ctx, userSessionPattern)
	if err != nil {
		return fmt.Errorf("获取用户会话键失败: %w", err)
	}

	// 删除每个Token对应的会话和登录信息
	for _, sessionKey := range sessionKeys {
		tokenData, err := s.storage.Get(ctx, sessionKey)
		if err != nil {
			continue
		}

		var token string
		if tokenBytes, ok := tokenData.([]byte); ok {
			token = string(tokenBytes)
		} else {
			token = fmt.Sprintf("%v", tokenData)
		}

		// 删除会话
		s.sessionService.DeleteSession(ctx, token)

		// 删除对应的登录信息
		loginKey := s.keyService.LoginInfoKey(token)
		s.storage.Delete(ctx, loginKey)

		// 删除用户会话映射
		s.storage.Delete(ctx, sessionKey)
	}

	return nil
}

// GetLoginInfo 获取登录信息
func (s *Service) GetLoginInfo(ctx context.Context, token string) (*core.LoginInfo, error) {
	loginKey := s.keyService.LoginInfoKey(token)
	data, err := s.storage.Get(ctx, loginKey)
	if err != nil {
		return nil, fmt.Errorf("获取登录信息失败: %w", err)
	}

	if data == nil {
		return nil, errors.New("登录信息不存在")
	}

	var loginInfo core.LoginInfo
	if dataBytes, ok := data.([]byte); ok {
		if err := json.Unmarshal(dataBytes, &loginInfo); err != nil {
			return nil, fmt.Errorf("解析登录信息失败: %w", err)
		}
	} else {
		// 处理其他类型的数据
		dataStr := fmt.Sprintf("%v", data)
		if err := json.Unmarshal([]byte(dataStr), &loginInfo); err != nil {
			return nil, fmt.Errorf("解析登录信息失败: %w", err)
		}
	}

	return &loginInfo, nil
}

// handleLoginMode 处理登录模式
func (s *Service) handleLoginMode(ctx context.Context, req *core.LoginRequest) error {
	switch s.config.LoginMode {
	case core.SingleLogin:
		// 单端登录：踢出该用户的所有其他会话
		return s.sessionService.KickOut(ctx, req.UserID)
	case core.MutexLogin:
		// 同端互斥登录：踢出该用户在同一设备的其他会话
		return s.kickOutSameDevice(ctx, req.UserID, req.Device)
	case core.MultiLogin:
		// 多端登录：不做处理
		return nil
	default:
		return nil
	}
}

// kickOutSameDevice 踢出同一设备的会话
func (s *Service) kickOutSameDevice(ctx context.Context, userID, device string) error {
	// 获取用户的所有会话Token
	pattern := s.keyService.UserSessionPattern(userID)
	keys, err := s.storage.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		tokenData, err := s.storage.Get(ctx, key)
		if err != nil {
			continue
		}

		var token string
		if tokenBytes, ok := tokenData.([]byte); ok {
			token = string(tokenBytes)
		} else {
			token = fmt.Sprintf("%v", tokenData)
		}

		// 获取会话详情
		session, err := s.sessionService.GetSession(ctx, token)
		if err != nil {
			continue
		}

		// 如果是同一设备，则删除会话
		if session.Device == device {
			s.sessionService.DeleteSession(ctx, session.Token)
		}
	}

	return nil
}

// storeLoginInfo 存储登录信息
func (s *Service) storeLoginInfo(ctx context.Context, token string, loginInfo *core.LoginInfo) error {
	loginKey := s.keyService.LoginInfoKey(token)
	data, err := json.Marshal(loginInfo)
	if err != nil {
		return err
	}

	return s.storage.Set(ctx, loginKey, data, s.config.TokenExpire)
}

// storeRefreshToken 存储刷新Token信息
func (s *Service) storeRefreshToken(ctx context.Context, refreshToken string, refreshInfo *core.RefreshTokenInfo) error {
	refreshKey := s.keyService.RefreshTokenKey(refreshToken)
	data, err := json.Marshal(refreshInfo)
	if err != nil {
		return err
	}

	return s.storage.Set(ctx, refreshKey, data, s.config.RefreshExpire)
}

// getRefreshTokenInfo 获取刷新Token信息
func (s *Service) getRefreshTokenInfo(ctx context.Context, refreshToken string) (*core.RefreshTokenInfo, error) {
	refreshKey := s.keyService.RefreshTokenKey(refreshToken)
	data, err := s.storage.Get(ctx, refreshKey)
	if err != nil {
		return nil, fmt.Errorf("获取刷新Token信息失败: %w", err)
	}

	if data == nil {
		return nil, errors.New("刷新Token不存在")
	}

	var refreshInfo core.RefreshTokenInfo
	if dataBytes, ok := data.([]byte); ok {
		if err := json.Unmarshal(dataBytes, &refreshInfo); err != nil {
			return nil, fmt.Errorf("解析刷新Token信息失败: %w", err)
		}
	} else {
		dataStr := fmt.Sprintf("%v", data)
		if err := json.Unmarshal([]byte(dataStr), &refreshInfo); err != nil {
			return nil, fmt.Errorf("解析刷新Token信息失败: %w", err)
		}
	}

	return &refreshInfo, nil
}

// RefreshAccessToken 刷新访问Token
func (s *Service) RefreshAccessToken(ctx context.Context, refreshToken string) (*core.LoginResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("刷新Token不能为空")
	}

	// 获取刷新Token信息
	refreshInfo, err := s.getRefreshTokenInfo(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("获取刷新Token信息失败: %w", err)
	}

	// 检查刷新Token是否过期
	if time.Now().After(refreshInfo.ExpiresAt) {
		// 删除过期的刷新Token
		s.storage.Delete(ctx, s.keyService.RefreshTokenKey(refreshToken))
		return nil, errors.New("刷新Token已过期")
	}

	// 生成新的访问Token
	tokenExtra := map[string]interface{}{
		"user_id": refreshInfo.UserID,
		"refresh": true,
	}

	newAccessToken, err := s.tokenGenerator.Generate(tokenExtra)
	if err != nil {
		return nil, fmt.Errorf("生成新访问Token失败: %w", err)
	}

	// 生成新的刷新Token
	newRefreshTokenExtra := map[string]interface{}{
		"user_id": refreshInfo.UserID,
		"type":    "refresh",
	}

	newRefreshToken, err := s.tokenGenerator.Generate(newRefreshTokenExtra)
	if err != nil {
		return nil, fmt.Errorf("生成新刷新Token失败: %w", err)
	}

	// 删除旧的刷新Token
	s.storage.Delete(ctx, s.keyService.RefreshTokenKey(refreshToken))

	// 创建新的会话
	now := time.Now()
	session := &core.Session{
		ID:         newAccessToken,
		UserID:     refreshInfo.UserID,
		Token:      newAccessToken,
		Device:     "", // 刷新时设备信息可能不可用
		IP:         "", // 刷新时IP信息可能不可用
		LoginTime:  now,
		LastAccess: now,
		Extra:      make(map[string]interface{}),
	}

	if err := s.sessionService.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("创建新会话失败: %w", err)
	}

	// 存储新的登录信息
	loginInfo := &core.LoginInfo{
		UserID:     refreshInfo.UserID,
		Token:      newAccessToken,
		Device:     "",
		IP:         "",
		LoginTime:  now,
		LastAccess: now,
	}

	if err := s.storeLoginInfo(ctx, newAccessToken, loginInfo); err != nil {
		return nil, fmt.Errorf("存储新登录信息失败: %w", err)
	}

	// 存储新的刷新Token
	newRefreshInfo := &core.RefreshTokenInfo{
		RefreshToken: newRefreshToken,
		UserID:       refreshInfo.UserID,
		Device:       refreshInfo.Device,
		CreatedAt:    now,
		ExpiresAt:    now.Add(s.config.RefreshExpire),
		Extra:        refreshInfo.Extra,
	}

	if err := s.storeRefreshToken(ctx, newRefreshToken, newRefreshInfo); err != nil {
		return nil, fmt.Errorf("存储新刷新Token失败: %w", err)
	}

	// 构造响应
	response := &core.LoginResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		ExpireTime:   now.Add(s.config.TokenExpire),
		UserInfo: &core.UserInfo{
			ID:       refreshInfo.UserID,
			Username: refreshInfo.UserID, // 简化处理
		},
	}

	return response, nil
}

// storeUserSessionMapping 存储用户会话映射
func (s *Service) storeUserSessionMapping(ctx context.Context, userID, token string) error {
	userSessionKey := s.keyService.UserSessionKey(userID, token)
	return s.storage.Set(ctx, userSessionKey, token, s.config.TokenExpire)
}
