package auth

import (
	"github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	"github.com/labstack/echo/v4"
)

func (h *AuthHandler) MapRoutes(e *echo.Group, refreshJwt jwt.Jwt, accessJwt jwt.Jwt, verificationJwt verifier.Jwt) *echo.Group {
	authGroup := e.Group("/auth")

	authGroup.POST("/register/customer", h.RegisterCustomer(verificationJwt))
	authGroup.POST("/register/provider", h.RegisterProvider(verificationJwt))
	authGroup.POST("/resend-confirmation", h.ResendConfirmationLetter(verificationJwt))

	authGroup.POST("/refresh", h.Refresh(refreshJwt), middleware.JWT(refreshJwt))
	authGroup.POST("/login", h.Login(accessJwt, refreshJwt))
	authGroup.POST("/logout", h.Logout)

	authGroup.GET("/verify", h.VerifyEmail(verificationJwt))

	return authGroup
}
