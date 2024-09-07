package auth

import (
	"github.com/labstack/echo/v4"
)

func (h *authHandler) MapRoutes(e *echo.Group) *echo.Group {
	registerGroup := e.Group("/register")
	
	registerGroup.POST("/customer", h.registerCustomer)
	registerGroup.POST("/provider", h.registerProvider)

	e.POST("/login", h.Login())

	return registerGroup
}
