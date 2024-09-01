package user

import (
	"github.com/hexley21/handy/internal/common/middleware"
	"github.com/labstack/echo/v4"
)

func (h *userHandler) MapRoutes(e *echo.Group, secretKey string) *echo.Group {
	usersGroup := e.Group("/users")

	
	
	usersGroup.Use(middleware.EchoJWTMiddleware(secretKey))
	usersGroup.Use(middleware.EchoIsSelfOrAdminMiddleware())
	usersGroup.GET("/:id", h.FindUserById())

	return usersGroup
}
