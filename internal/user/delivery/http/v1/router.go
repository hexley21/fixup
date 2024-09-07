package v1

import (
	"github.com/hexley21/handy/internal/common/jwt"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/user"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/pkg/config"
	"github.com/labstack/echo/v4"
)

type v1Router struct {
	cfgJwt       config.JWT
	authService  service.AuthService
	userService  service.UserService
}

func NewRouter(cfgJwt config.JWT, authService service.AuthService, userService service.UserService) *v1Router {
	return &v1Router{
		cfgJwt:       cfgJwt,
		authService:  authService,
		userService:  userService,
	}
}

func (v1r *v1Router) MapV1Routes(echo *echo.Echo) *echo.Group {
	accessAuthJwt := jwt.NewAuthJwtImpl(v1r.cfgJwt.AccessSecret, v1r.cfgJwt.AccessTTL)
	refreshAuthJwt := jwt.NewAuthJwtImpl(v1r.cfgJwt.RefreshSecret, v1r.cfgJwt.RefreshTTL)

	accessJwtGenerator := accessAuthJwt
	refreshJwtGenerator := refreshAuthJwt

	v1Group := echo.Group("/v1")

	auth.NewAuthHandler(v1r.authService, accessJwtGenerator, refreshJwtGenerator).MapRoutes(v1Group)
	user.NewUserHandler(v1r.userService).MapRoutes(v1Group, v1r.cfgJwt.AccessSecret)

	return v1Group
}
