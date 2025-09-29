package core

import "fmt"

// KeyService 键生成服务
type KeyService struct {
	prefix string
}

// NewKeyService 创建键生成服务
func NewKeyService(prefix string) *KeyService {
	if prefix == "" {
		prefix = "gstoken"
	}
	return &KeyService{
		prefix: prefix,
	}
}

// 登录相关键
func (k *KeyService) LoginInfoKey(token string) string {
	return fmt.Sprintf("%s:login:%s", k.prefix, token)
}

func (k *KeyService) RefreshTokenKey(refreshToken string) string {
	return fmt.Sprintf("%s:refresh:%s", k.prefix, refreshToken)
}

// 会话相关键
func (k *KeyService) SessionKey(token string) string {
	return fmt.Sprintf("%s:session:%s", k.prefix, token)
}

func (k *KeyService) UserSessionKey(userID, token string) string {
	return fmt.Sprintf("%s:user_session:%s:%s", k.prefix, userID, token)
}

func (k *KeyService) UserSessionPattern(userID string) string {
	return fmt.Sprintf("%s:user_session:%s:*", k.prefix, userID)
}

// 权限相关键
func (k *KeyService) RoleKey(roleID string) string {
	return fmt.Sprintf("%s:role:%s", k.prefix, roleID)
}

func (k *KeyService) UserRoleKey(userID string) string {
	return fmt.Sprintf("%s:user_role:%s", k.prefix, userID)
}

// 用户相关键
func (k *KeyService) UserInfoKey(userID string) string {
	return fmt.Sprintf("%s:user:%s", k.prefix, userID)
}

func (k *KeyService) UserTokensKey(userID string) string {
	return fmt.Sprintf("%s:user_tokens:%s", k.prefix, userID)
}

// 设备相关键
func (k *KeyService) DeviceKey(userID, device string) string {
	return fmt.Sprintf("%s:device:%s:%s", k.prefix, userID, device)
}

func (k *KeyService) DevicePattern(userID string) string {
	return fmt.Sprintf("%s:device:%s:*", k.prefix, userID)
}

// SSO相关键
func (k *KeyService) SSOTicketKey(ticket string) string {
	return fmt.Sprintf("%s:sso:ticket:%s", k.prefix, ticket)
}

func (k *KeyService) SSOSessionKey(sessionID string) string {
	return fmt.Sprintf("%s:sso:session:%s", k.prefix, sessionID)
}

// 自定义键生成
func (k *KeyService) CustomKey(category string, keys ...string) string {
	key := fmt.Sprintf("%s:%s", k.prefix, category)
	for _, k := range keys {
		key += ":" + k
	}
	return key
}

// 获取前缀
func (k *KeyService) GetPrefix() string {
	return k.prefix
}

// 设置前缀
func (k *KeyService) SetPrefix(prefix string) {
	k.prefix = prefix
}
