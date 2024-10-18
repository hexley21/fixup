package ctx_util

type ctxKey string

const (
	// jWTIdKey         ctxKey = "jwt_id"
	// jWTRoleKey       ctxKey = "jwt_role"
	// jWTUserStatusKey ctxKey = "jwt_user_status"
	// paramIdKey       ctxKey = "param_id"
)

var (
	// ErrJWTIdNotSet         = errors.New("jwt id was not set")
	// ErrJWTRoleNotSet       = errors.New("jwt role was not set")
	// ErrJWTUserStatusNotSet = errors.New("jwt user status was not set")
	// ErrParamIdNotSet = errors.New("param id was not set")
)

// func GetJWTId(ctx context.Context) (string, *rest.ErrorResponse) {
// 	if value, ok := ctx.Value(jWTIdKey).(string); ok {
// 		return value, nil
// 	}

// 	return "", rest.NewInternalServerError(ErrJWTIdNotSet)
// }

// func SetJWTId(ctx context.Context, value string) context.Context {
// 	return context.WithValue(ctx, jWTIdKey, value)
// }

// func GetJWTRole(ctx context.Context) (enum.UserRole, *rest.ErrorResponse) {
// 	if role, ok := ctx.Value(jWTRoleKey).(enum.UserRole); ok {
// 		return role, nil
// 	}

// 	return "", rest.NewInternalServerError(ErrJWTRoleNotSet)
// }

// func SetJWTRole(ctx context.Context, value enum.UserRole) context.Context {
// 	return context.WithValue(ctx, jWTRoleKey, value)
// }

// func GetJWTUserStatus(ctx context.Context) (bool, *rest.ErrorResponse) {
// 	if role, ok := ctx.Value(jWTUserStatusKey).(bool); ok {
// 		return role, nil
// 	}

// 	return false, rest.NewInternalServerError(ErrJWTUserStatusNotSet)
// }

// func SetJWTUserStatus(ctx context.Context, value bool) context.Context {
// 	return context.WithValue(ctx, jWTUserStatusKey, value)
// }

// func GetParamId(ctx context.Context) (int64, *rest.ErrorResponse) {
// 	if paramId, ok := ctx.Value(paramIdKey).(int64); ok {
// 		return paramId, nil
// 	}

// 	return int64(0), rest.NewInternalServerError(ErrParamIdNotSet)
// }

// func SetParamId(ctx context.Context, value int64) context.Context {
// 	return context.WithValue(ctx, paramIdKey, value)
// }
