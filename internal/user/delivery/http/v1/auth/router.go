package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/jwt/verify_jwt"
)

func MapRoutes(
	mf *middleware.MiddlewareFactory,
	h *Handler,
	accessJwtManager auth_jwt.JWTManager,
	refreshJWTManager auth_jwt.JWTManager,
	vrfJWTManager verify_jwt.JWTManager,
	r chi.Router,
) chi.Router {
	r.Route("/auth", func(auth chi.Router) {
		auth.Post("/register/customer", h.RegisterCustomer(vrfJWTManager))
		auth.Post("/register/provider", h.RegisterProvider(vrfJWTManager))
		auth.Post("/resend-confirmation", h.ResendConfirmationLetter(vrfJWTManager))

		auth.With(mf.NewJWT(refreshJWTManager)).Post("/refresh", h.Refresh(accessJwtManager))
		auth.Post("/login", h.Login(accessJwtManager, refreshJWTManager))
		auth.Post("/logout", h.Logout)

		auth.Get("/verify", h.VerifyEmail(vrfJWTManager))
	})

	return r
}
