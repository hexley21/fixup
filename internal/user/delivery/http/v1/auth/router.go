package auth

import (
	"github.com/labstack/echo/v4"
)

func (h *authHandler) MapRoutes(e *echo.Group) {
	registerGroup := e.Group("/register")
	registerGroup.GET("/customer", h.RegisterCustomer())
}
