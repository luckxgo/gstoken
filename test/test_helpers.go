package test

import (
	"context"
	"encoding/json"
	"github.com/luckxgo/gstoken/core"
)

// testUserRoleProvider 测试用的用户角色提供者
type testUserRoleProvider struct {
	storage core.Storage
}

func (p *testUserRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]core.Role, error) {
	// 从存储中获取用户角色关联，使用与测试用例一致的键格式
	userRoleKey := "gstoken:user_role:" + userID
	roleIDsData, err := p.storage.Get(ctx, userRoleKey)
	if err != nil {
		return nil, err
	}

	var roleIDs []string
	if roleIDsData != nil {
		// 尝试解析为字符串数组
		if roleIDsBytes, ok := roleIDsData.([]byte); ok {
			// 如果是字节数组，直接解析
			err = json.Unmarshal(roleIDsBytes, &roleIDs)
			if err != nil {
				return nil, err
			}
		} else if roleIDsStr, ok := roleIDsData.(string); ok {
			// 如果是字符串，尝试解析JSON
			err = json.Unmarshal([]byte(roleIDsStr), &roleIDs)
			if err != nil {
				return nil, err
			}
		}
	}

	var roles []core.Role
	for _, roleID := range roleIDs {
		roleKey := "gstoken:role:" + roleID
		roleData, err := p.storage.Get(ctx, roleKey)
		if err != nil {
			continue
		}

		var role core.Role
		if roleDataBytes, ok := roleData.([]byte); ok {
			// 如果是字节数组，直接解析
			err = json.Unmarshal(roleDataBytes, &role)
			if err != nil {
				// 如果解析失败，创建简单角色
				role = core.Role{
					ID:   roleID,
					Name: string(roleDataBytes),
				}
			}
		} else if roleDataStr, ok := roleData.(string); ok {
			// 如果是字符串，尝试解析JSON
			err = json.Unmarshal([]byte(roleDataStr), &role)
			if err != nil {
				// 如果解析失败，创建简单角色
				role = core.Role{
					ID:   roleID,
					Name: roleDataStr,
				}
			}
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// newTestUserRoleProvider 创建测试用的用户角色提供者
func newTestUserRoleProvider(storage core.Storage) *testUserRoleProvider {
	return &testUserRoleProvider{storage: storage}
}
