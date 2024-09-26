package ctxutil

import (
	"context"
	"errors"

	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/internal/user/enum"
)

type ctxKey string

const (
	jwtIdKey         ctxKey = "jwt_id"
	jwtRoleKey       ctxKey = "jwt_role"
	jwtUserStatusKey ctxKey = "jwt_user_status"
	paramIdKey       ctxKey = "param_id"
)

var (
	ErrJwtNotImplemented     = errors.New("jwt middleware not implemented")
	ErrParamIdNotImplemented = errors.New("param id middleware not implemented")
)

func GetJwtId(ctx context.Context) (string, *rest.ErrorResponse) {
	if value, ok := ctx.Value(jwtIdKey).(string); ok {
		return value, nil
	}

	return "", rest.NewInternalServerError(ErrJwtNotImplemented)
}

func SetJwtId(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, jwtIdKey, value)
}

func GetJwtRole(ctx context.Context) (enum.UserRole, *rest.ErrorResponse) {
	if role, ok := ctx.Value(jwtRoleKey).(enum.UserRole); ok {
		return role, nil
	}

	return "", rest.NewInternalServerError(ErrJwtNotImplemented)
}

func SetJwtRole(ctx context.Context, value enum.UserRole) context.Context {
	return context.WithValue(ctx, jwtRoleKey, value)
}

func GetJwtUserStatus(ctx context.Context) (bool, *rest.ErrorResponse) {
	if role, ok := ctx.Value(jwtUserStatusKey).(bool); ok {
		return role, nil
	}

	return false, rest.NewInternalServerError(ErrJwtNotImplemented)
}

func SetJwtUserStatus(ctx context.Context, value bool) context.Context {
	return context.WithValue(ctx, jwtUserStatusKey, value)
}

func GetParamId(ctx context.Context) (int64, *rest.ErrorResponse) {
	if paramId, ok := ctx.Value(paramIdKey).(int64); ok {
		return paramId, nil
	}

	return int64(0), rest.NewInternalServerError(ErrParamIdNotImplemented)
}

func SetParamId(ctx context.Context, value int64) context.Context {
	return context.WithValue(ctx, paramIdKey, value)
}
