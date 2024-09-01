package auth

import (
	"github.com/labstack/echo/v4"
)

func (h *authHandler) MapRoutes(e *echo.Group) *echo.Group {
	registerGroup := e.Group("/register")
	registerGroup.POST("/customer", h.RegisterCustomer())

	return registerGroup
}
