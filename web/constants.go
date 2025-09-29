package web

// 上下文键常量
const (
	ContextKeyUserID   = "user_id"
	ContextKeyToken    = "token"
	ContextKeyUserInfo = "user_info"
)

// HTTP头常量
const (
	HeaderAuthorization = "Authorization"
	HeaderXToken        = "X-Token"
	BearerPrefix        = "Bearer "
)

// 查询参数常量
const (
	QueryParamToken = "token"
)

// 错误响应常量
const (
	ErrorUnauthorized = "unauthorized"
	ErrorForbidden    = "forbidden"
	ErrorMessage      = "message"
)
