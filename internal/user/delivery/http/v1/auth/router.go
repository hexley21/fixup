package auth

import (
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	"github.com/labstack/echo/v4"
)

func (h *authHandler) MapRoutes(e *echo.Group, refreshJwt jwt.Jwt, accessJwt jwt.Jwt, verJwtVerifier verifier.JwtVerifier) *echo.Group {
	authGroup := e.Group("/auth")

	authGroup.POST("/register/customer", h.registerCustomer)
	authGroup.POST("/register/provider", h.registerProvider)

	authGroup.POST("/refresh", h.refresh(refreshJwt), middleware.JWT(refreshJwt))
	authGroup.POST("/login", h.login(accessJwt, refreshJwt))
	authGroup.POST("/logout", h.logout)

	authGroup.GET("/verify", h.verifyEmail(verJwtVerifier))

	return authGroup
}
