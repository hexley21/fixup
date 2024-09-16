package auth

import (
	"github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	"github.com/labstack/echo/v4"
)

func (h *authHandler) MapRoutes(e *echo.Group, refreshJwt jwt.Jwt, accessJwt jwt.Jwt, verificationJwt verifier.Jwt) *echo.Group {
	authGroup := e.Group("/auth")

	authGroup.POST("/register/customer", h.registerCustomer(verificationJwt))
	authGroup.POST("/register/provider", h.registerProvider(verificationJwt))
	authGroup.POST("/resend-confirmation", h.resendConfirmationLetter(verificationJwt))

	authGroup.POST("/refresh", h.refresh(refreshJwt), middleware.JWT(refreshJwt))
	authGroup.POST("/login", h.login(accessJwt, refreshJwt))
	authGroup.POST("/logout", h.logout)

	authGroup.GET("/verify", h.verifyEmail(verificationJwt))

	return authGroup
}
