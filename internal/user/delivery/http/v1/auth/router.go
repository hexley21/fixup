package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/service/verifier"
)

func MapRoutes(
	mf *middleware.MiddlewareFactory,
	hf *HandlerFactory,
	accessJwtManager auth_jwt.JWTManager,
	refreshJWTManager auth_jwt.JWTManager,
	vrfJWTManager verifier.JWTManager,
	r chi.Router,
) chi.Router {
	r.Route("/auth", func(auth chi.Router) {
		auth.Post("/register/customer", hf.RegisterCustomer(vrfJWTManager))
		auth.Post("/register/provider", hf.RegisterProvider(vrfJWTManager))
		auth.Post("/resend-confirmation", hf.ResendConfirmationLetter(vrfJWTManager))

		auth.With(mf.NewJWT(refreshJWTManager)).Post("/refresh", hf.Refresh(accessJwtManager))
		auth.Post("/login", hf.Login(accessJwtManager, refreshJWTManager))
		auth.Post("/logout", hf.Logout)

		auth.Get("/verify", hf.VerifyEmail(vrfJWTManager))
	})

	return r
}
