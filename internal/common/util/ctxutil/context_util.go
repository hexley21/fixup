package ctxutil

import (
	"errors"
	"strconv"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/labstack/echo/v4"
)

const (
	jwtIdKey = "jwt_id"
	jwtRoleKey = "jwt_role"
	jwtVerificationKey = "jwt_verified"
	paramIdKey = "param_id"
)

var (
	errJwtNotImplemented = rest.NewInternalServerError(errors.New("jwt middleware not implemented"))
	errParamIdNotImplemented = rest.NewInternalServerError(errors.New("param id middleware not implemented"))
)



func GetJwtId(c echo.Context) (string, error) {
	if id, ok := c.Get(jwtIdKey).(string); ok {
		return id, nil
	}

	return "", errJwtNotImplemented
}

func SetJwtId(c echo.Context, value string) {
	c.Set(jwtIdKey, value)
}

func GetJwtRole(c echo.Context) (enum.UserRole, error) {
	if role, ok := c.Get(jwtRoleKey).(enum.UserRole); ok {
		return role, nil
	}

	return "", errJwtNotImplemented
}

func SetJwtRole(c echo.Context, value enum.UserRole) {
	c.Set(jwtRoleKey, value)
}

func GetJwtVerification(c echo.Context) (bool, error) {
	if verification, ok := c.Get(jwtVerificationKey).(bool); ok {
		return verification, nil
	}

	return false, errJwtNotImplemented
}

func SetJwtVerification(c echo.Context, value bool) {
	c.Set(jwtVerificationKey, value)
}

func GetParamId(c echo.Context) (int64, error) {
	if paramId, ok := c.Get(paramIdKey).(int64); ok {
		return paramId, nil
	}

	return 0, errParamIdNotImplemented
}

func SetParamId(c echo.Context, value string) error {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return rest.NewBadRequestError(err, "Invalid id parameter")
	}

	c.Set(paramIdKey, id)
	return nil
}