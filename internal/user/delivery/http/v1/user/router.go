package user

import (
	"github.com/labstack/echo/v4"
)

func (h *userHandler) MapRoutes(e *echo.Group) {
	usersGroup := e.Group("/users")
	usersGroup.GET("/:id", h.FindUserById())
}
