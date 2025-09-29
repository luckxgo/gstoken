package web

import (
	"context"
	"fmt"
	"gstoken/core"
	"reflect"
	"runtime"
	"strings"
)

// AuthDecorator 方法式鉴权装饰器
type AuthDecorator struct {
	gsToken GSTokenAdapter
	config  *AuthConfig
}

// NewAuthDecorator 创建方法式鉴权装饰器
func NewAuthDecorator(gsToken GSTokenAdapter, config *AuthConfig) *AuthDecorator {
	if config == nil {
		config = DefaultAuthConfig()
	}

	return &AuthDecorator{
		gsToken: gsToken,
		config:  config,
	}
}

// AuthContext 认证上下文
type AuthContext struct {
	context.Context
	UserID   string
	Token    string
	UserInfo *core.UserInfo
}

// NewAuthContext 创建认证上下文
func NewAuthContext(ctx context.Context, userID, token string, userInfo *core.UserInfo) *AuthContext {
	return &AuthContext{
		Context:  ctx,
		UserID:   userID,
		Token:    token,
		UserInfo: userInfo,
	}
}

// RequireAuth 要求认证的方法装饰器
// 使用示例：
// decoratedFunc := decorator.RequireAuth(originalFunc)
// result := decoratedFunc(ctx, token, args...)
func (d *AuthDecorator) RequireAuth(fn interface{}) interface{} {
	return d.wrapFunction(fn, func(ctx context.Context, token string) (*AuthContext, error) {
		if token == "" {
			return nil, core.ErrTokenNotFound
		}

		userInfo, err := d.gsToken.Verify(ctx, token)
		if err != nil {
			return nil, err
		}

		return NewAuthContext(ctx, userInfo.ID, token, userInfo), nil
	})
}

// RequirePermission 要求特定权限的方法装饰器
func (d *AuthDecorator) RequirePermission(permission string) func(interface{}) interface{} {
	return func(fn interface{}) interface{} {
		return d.wrapFunction(fn, func(ctx context.Context, token string) (*AuthContext, error) {
			if token == "" {
				return nil, core.ErrTokenNotFound
			}

			userInfo, err := d.gsToken.Verify(ctx, token)
			if err != nil {
				return nil, err
			}

			hasPermission, err := d.gsToken.CheckPermission(ctx, userInfo.ID, permission)
			if err != nil {
				return nil, err
			}

			if !hasPermission {
				return nil, core.ErrPermissionDenied
			}

			return NewAuthContext(ctx, userInfo.ID, token, userInfo), nil
		})
	}
}

// RequireRole 要求特定角色的方法装饰器
func (d *AuthDecorator) RequireRole(role string) func(interface{}) interface{} {
	return func(fn interface{}) interface{} {
		return d.wrapFunction(fn, func(ctx context.Context, token string) (*AuthContext, error) {
			if token == "" {
				return nil, core.ErrTokenNotFound
			}

			userInfo, err := d.gsToken.Verify(ctx, token)
			if err != nil {
				return nil, err
			}

			hasRole, err := d.gsToken.CheckRole(ctx, userInfo.ID, role)
			if err != nil {
				return nil, err
			}

			if !hasRole {
				return nil, core.ErrRoleNotFound
			}

			return NewAuthContext(ctx, userInfo.ID, token, userInfo), nil
		})
	}
}

// RequireAnyPermission 要求任意权限的方法装饰器
func (d *AuthDecorator) RequireAnyPermission(permissions ...string) func(interface{}) interface{} {
	return func(fn interface{}) interface{} {
		return d.wrapFunction(fn, func(ctx context.Context, token string) (*AuthContext, error) {
			if token == "" {
				return nil, core.ErrTokenNotFound
			}

			userInfo, err := d.gsToken.Verify(ctx, token)
			if err != nil {
				return nil, err
			}

			for _, permission := range permissions {
				hasPermission, err := d.gsToken.CheckPermission(ctx, userInfo.ID, permission)
				if err == nil && hasPermission {
					return NewAuthContext(ctx, userInfo.ID, token, userInfo), nil
				}
			}

			return nil, core.ErrPermissionDenied
		})
	}
}

// RequireAllPermissions 要求所有权限的方法装饰器
func (d *AuthDecorator) RequireAllPermissions(permissions ...string) func(interface{}) interface{} {
	return func(fn interface{}) interface{} {
		return d.wrapFunction(fn, func(ctx context.Context, token string) (*AuthContext, error) {
			if token == "" {
				return nil, core.ErrTokenNotFound
			}

			userInfo, err := d.gsToken.Verify(ctx, token)
			if err != nil {
				return nil, err
			}

			for _, permission := range permissions {
				hasPermission, err := d.gsToken.CheckPermission(ctx, userInfo.ID, permission)
				if err != nil || !hasPermission {
					return nil, core.ErrPermissionDenied
				}
			}

			return NewAuthContext(ctx, userInfo.ID, token, userInfo), nil
		})
	}
}

// RequireAnyRole 要求任意角色的方法装饰器
func (d *AuthDecorator) RequireAnyRole(roles ...string) func(interface{}) interface{} {
	return func(fn interface{}) interface{} {
		return d.wrapFunction(fn, func(ctx context.Context, token string) (*AuthContext, error) {
			if token == "" {
				return nil, core.ErrTokenNotFound
			}

			userInfo, err := d.gsToken.Verify(ctx, token)
			if err != nil {
				return nil, err
			}

			for _, role := range roles {
				hasRole, err := d.gsToken.CheckRole(ctx, userInfo.ID, role)
				if err == nil && hasRole {
					return NewAuthContext(ctx, userInfo.ID, token, userInfo), nil
				}
			}

			return nil, core.ErrRoleNotFound
		})
	}
}

// RequireAllRoles 要求所有角色的方法装饰器
func (d *AuthDecorator) RequireAllRoles(roles ...string) func(interface{}) interface{} {
	return func(fn interface{}) interface{} {
		return d.wrapFunction(fn, func(ctx context.Context, token string) (*AuthContext, error) {
			if token == "" {
				return nil, core.ErrTokenNotFound
			}

			userInfo, err := d.gsToken.Verify(ctx, token)
			if err != nil {
				return nil, err
			}

			for _, role := range roles {
				hasRole, err := d.gsToken.CheckRole(ctx, userInfo.ID, role)
				if err != nil || !hasRole {
					return nil, core.ErrRoleNotFound
				}
			}

			return NewAuthContext(ctx, userInfo.ID, token, userInfo), nil
		})
	}
}

// wrapFunction 包装函数的通用方法
func (d *AuthDecorator) wrapFunction(fn interface{}, authFunc func(context.Context, string) (*AuthContext, error)) interface{} {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// 验证函数签名
	if fnType.Kind() != reflect.Func {
		panic("decorator can only be applied to functions")
	}

	if fnType.NumIn() < 2 {
		panic("function must have at least 2 parameters: (context.Context, token string, ...)")
	}

	// 检查第一个参数是否为 context.Context
	if !fnType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		panic("first parameter must be context.Context")
	}

	// 检查第二个参数是否为 string (token)
	if fnType.In(1).Kind() != reflect.String {
		panic("second parameter must be string (token)")
	}

	// 创建包装函数
	return reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		ctx := args[0].Interface().(context.Context)
		token := args[1].String()

		// 执行认证
		authCtx, err := authFunc(ctx, token)
		if err != nil {
			// 如果函数返回 error，将认证错误作为最后一个返回值
			if fnType.NumOut() > 0 && fnType.Out(fnType.NumOut()-1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				results := make([]reflect.Value, fnType.NumOut())
				for i := 0; i < fnType.NumOut()-1; i++ {
					results[i] = reflect.Zero(fnType.Out(i))
				}
				results[fnType.NumOut()-1] = reflect.ValueOf(err)
				return results
			}

			// 如果函数不返回 error，panic
			panic(fmt.Sprintf("authentication failed: %v", err))
		}

		// 替换第一个参数为认证上下文
		newArgs := make([]reflect.Value, len(args))
		newArgs[0] = reflect.ValueOf(authCtx)
		copy(newArgs[1:], args[1:])

		// 调用原函数
		return fnValue.Call(newArgs)
	}).Interface()
}

// GetFunctionName 获取函数名称（用于调试）
func GetFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

// GetShortFunctionName 获取简短的函数名称
func GetShortFunctionName(fn interface{}) string {
	fullName := GetFunctionName(fn)
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}
