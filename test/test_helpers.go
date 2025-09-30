package test

import (
	"context"
	"encoding/json"
	"fmt"

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
		roleIDsBytes, ok := roleIDsData.([]byte)
		if !ok {
			return nil, fmt.Errorf("用户角色数据格式错误")
		}
		err = json.Unmarshal(roleIDsBytes, &roleIDs)
		if err != nil {
			return nil, err
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
		roleDataBytes, ok := roleData.([]byte)
		if !ok {
			continue
		}
		err = json.Unmarshal(roleDataBytes, &role)
		if err != nil {
			// 如果解析失败，创建简单角色
			role = core.Role{
				ID:   roleID,
				Name: string(roleDataBytes),
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
