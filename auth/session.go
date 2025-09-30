package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/luckxgo/gstoken/core"
)

// SessionService 会话服务实现
type SessionServiceImpl struct {
	storage    core.Storage
	config     *core.Config
	keyService *core.KeyService
}

// NewSessionService 创建新的会话服务
func NewSessionService(storage core.Storage, config *core.Config, keyService *core.KeyService) core.SessionService {
	return &SessionServiceImpl{
		storage:    storage,
		config:     config,
		keyService: keyService,
	}
}

// CreateSession 创建会话
func (s *SessionServiceImpl) CreateSession(ctx context.Context, session *core.Session) error {
	if session == nil {
		return errors.New("会话信息不能为空")
	}

	if session.Token == "" {
		return errors.New("Token不能为空")
	}

	if session.UserID == "" {
		return errors.New("用户ID不能为空")
	}

	// 存储会话数据 - 直接存储 session 对象，让 storage.Set 内部进行 JSON 序列化
	sessionKey := s.keyService.SessionKey(session.Token)
	if err := s.storage.Set(ctx, sessionKey, session, s.config.TokenExpire); err != nil {
		return fmt.Errorf("存储会话数据失败: %w", err)
	}

	// 存储用户会话映射（用于踢人下线）
	userSessionKey := s.keyService.UserSessionKey(session.UserID, session.Token)
	if err := s.storage.Set(ctx, userSessionKey, session.Token, s.config.TokenExpire); err != nil {
		return fmt.Errorf("存储用户会话映射失败: %w", err)
	}

	return nil
}

// GetSession 获取会话
func (s *SessionServiceImpl) GetSession(ctx context.Context, token string) (*core.Session, error) {
	if token == "" {
		return nil, errors.New("Token不能为空")
	}

	sessionKey := s.keyService.SessionKey(token)
	data, err := s.storage.Get(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("获取会话数据失败: %w", err)
	}

	if data == nil {
		return nil, errors.New("会话不存在")
	}

	var session core.Session
	dataBytes, ok := data.([]byte)
	if !ok {
		return nil, fmt.Errorf("会话数据格式错误，期望字节数组")
	}

	if err := json.Unmarshal(dataBytes, &session); err != nil {
		return nil, fmt.Errorf("解析会话数据失败: %w", err)
	}

	return &session, nil
}

// UpdateSession 更新会话
func (s *SessionServiceImpl) UpdateSession(ctx context.Context, session *core.Session) error {
	if session == nil {
		return errors.New("会话信息不能为空")
	}

	if session.Token == "" {
		return errors.New("Token不能为空")
	}

	// 检查会话是否存在
	exists, err := s.storage.Exists(ctx, s.keyService.SessionKey(session.Token))
	if err != nil {
		return fmt.Errorf("检查会话是否存在失败: %w", err)
	}

	if !exists {
		return errors.New("会话不存在")
	}

	// 更新会话数据 - 直接存储 session 对象，让 storage.Set 内部进行 JSON 序列化
	sessionKey := s.keyService.SessionKey(session.Token)
	if err := s.storage.Set(ctx, sessionKey, session, s.config.TokenExpire); err != nil {
		return fmt.Errorf("更新会话数据失败: %w", err)
	}

	return nil
}

// DeleteSession 删除会话
func (s *SessionServiceImpl) DeleteSession(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("Token不能为空")
	}

	// 先获取会话信息以便删除用户会话映射
	session, err := s.GetSession(ctx, token)
	if err != nil {
		// 如果会话不存在，直接返回成功
		return nil
	}

	// 删除会话数据
	sessionKey := s.keyService.SessionKey(token)
	if err := s.storage.Delete(ctx, sessionKey); err != nil {
		return fmt.Errorf("删除会话数据失败: %w", err)
	}

	// 删除用户会话映射
	userSessionKey := s.keyService.UserSessionKey(session.UserID, token)
	if err := s.storage.Delete(ctx, userSessionKey); err != nil {
		// 删除映射失败不影响主要操作
	}

	return nil
}

// KickOut 踢出用户的所有会话
func (s *SessionServiceImpl) KickOut(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("用户ID不能为空")
	}

	// 获取用户的所有会话Token
	pattern := s.keyService.UserSessionPattern(userID)
	keys, err := s.storage.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("获取用户会话列表失败: %w", err)
	}

	// 删除所有会话
	for _, key := range keys {
		tokenData, err := s.storage.Get(ctx, key)
		if err != nil {
			continue
		}

		tokenBytes, ok := tokenData.([]byte)
		if !ok {
			continue
		}

		// 反序列化 JSON 字符串
		var token string
		if err := json.Unmarshal(tokenBytes, &token); err != nil {
			// 如果反序列化失败，尝试直接使用字节数组
			token = string(tokenBytes)
		}

		// 删除会话
		s.DeleteSession(ctx, token)
	}

	return nil
}

// KickOutByToken 根据Token踢出会话
func (s *SessionServiceImpl) KickOutByToken(ctx context.Context, token string) error {
	return s.DeleteSession(ctx, token)
}
