package user

import (
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/labstack/echo/v4"
)

func (h *userHandler) MapRoutes(e *echo.Group, accessSecretKey string) *echo.Group {
	usersGroup := e.Group("/users")

	usersGroup.Use(middleware.EchoJWTMiddleware(accessSecretKey))
	usersGroup.Use(middleware.EchoIsSelfOrAdminMiddleware())
	
	usersGroup.GET("/:id", h.findUserById)
	usersGroup.POST("/:id/pfp", h.uploadProfilePicture)

	return usersGroup
}
