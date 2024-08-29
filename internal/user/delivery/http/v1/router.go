package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/user"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/pkg/logger"
)

type v1Router struct {
	logger  logger.Logger
	authService service.AuthService
	userService service.UserService
}

func NewRouter(logger  logger.Logger, authService service.AuthService, userService service.UserService) *v1Router {
	return &v1Router{
		logger:  logger,
		authService: authService,
		userService: userService,
	}
}

func (v1r *v1Router) MapV1Routes(router chi.Router) chi.Router {
	return router.Route("/v1", func(r chi.Router) {
		auth.NewAuthHandler(r, v1r.logger, v1r.authService).MapRoutes()
		user.NewUserHandler(r, v1r.logger, v1r.userService).MapRoutes()
	})
}
