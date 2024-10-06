package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/jwt/refresh_jwt"
	"github.com/hexley21/fixup/internal/user/jwt/verify_jwt"
)

func MapRoutes(
	mf *middleware.MiddlewareFactory,
	h *Handler,
	refreshJWTMiddleware func(http.Handler) http.Handler,
	accessJwtManager auth_jwt.JWTManager,
	refreshJwtManager refresh_jwt.JWTManager,
	vrfJWTManager verify_jwt.JWTManager,
	r chi.Router,
) chi.Router {
	r.Route("/auth", func(auth chi.Router) {
		auth.Post("/register/customer", h.RegisterCustomer(vrfJWTManager))
		auth.Post("/register/provider", h.RegisterProvider(vrfJWTManager))
		auth.Post("/resend-confirmation", h.ResendConfirmationLetter(vrfJWTManager))

		auth.With(refreshJWTMiddleware).Post("/refresh", h.Refresh(accessJwtManager))
		auth.Post("/login", h.Login(accessJwtManager, refreshJwtManager))
		auth.Post("/logout", h.Logout)

		auth.Get("/verify", h.VerifyEmail(vrfJWTManager))
	})

	return r
}
