package auth

import (
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/labstack/echo/v4"
)

func (h *authHandler) MapRoutes(e *echo.Group, refreshSecretKey string) *echo.Group {
	authGroup := e.Group("auth")

	authGroup.POST("/register/customer", h.registerCustomer)
	authGroup.POST("/register/provider", h.registerProvider)

	authGroup.POST("/refresh", h.refresh, middleware.JWT(refreshSecretKey))
	authGroup.POST("/login", h.login)
	authGroup.POST("/logout", h.logout)

	return authGroup
}
