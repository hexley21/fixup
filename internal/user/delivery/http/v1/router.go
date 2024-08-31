package v1

import (
	"github.com/hexley21/handy/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/user"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/labstack/echo/v4"
)

type v1Router struct {
	authService service.AuthService
	userService service.UserService
}

func NewRouter(authService service.AuthService, userService service.UserService) *v1Router {
	return &v1Router{
		authService: authService,
		userService: userService,
	}
}

func (v1r *v1Router) MapV1Routes(echo *echo.Echo) *echo.Group {
	v1Group := echo.Group("/v1")

	auth.NewAuthHandler(v1r.authService).MapRoutes(v1Group)
	user.NewUserHandler(v1r.userService).MapRoutes(v1Group)

	return v1Group
}
