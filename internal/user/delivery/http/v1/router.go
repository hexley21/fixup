package v1

import (
	"github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/user"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/labstack/echo/v4"
)

type v1Router struct {
	cfgJwt      config.JWT
	verifierJwt verifier.Jwt
	authService service.AuthService
	userService service.UserService
}

func NewRouter(cfgJwt config.JWT, verifierJwt verifier.Jwt, authService service.AuthService, userService service.UserService) *v1Router {
	return &v1Router{
		cfgJwt:      cfgJwt,
		verifierJwt: verifierJwt,
		authService: authService,
		userService: userService,
	}
}

func (v1r *v1Router) MapV1Routes(echo *echo.Echo) *echo.Group {
	accessAuthJwt := jwt.NewAuthJwtImpl(v1r.cfgJwt.AccessSecret, v1r.cfgJwt.AccessTTL)
	refreshAuthJwt := jwt.NewAuthJwtImpl(v1r.cfgJwt.RefreshSecret, v1r.cfgJwt.RefreshTTL)

	accessJwtMiddleware := middleware.JWT(accessAuthJwt)

	v1Group := echo.Group("/v1")

	auth.NewAuthHandler(v1r.authService).MapRoutes(v1Group, accessAuthJwt, refreshAuthJwt, v1r.verifierJwt)
	user.NewUserHandler(v1r.userService).MapRoutes(v1Group, accessJwtMiddleware)

	return v1Group
}
