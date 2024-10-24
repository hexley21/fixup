package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/user"
	"github.com/hexley21/fixup/internal/user/jwt/refresh_jwt"
	"github.com/hexley21/fixup/internal/user/jwt/verify_jwt"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/infra/cdn"
)

type RouterArgs struct {
	AuthService            service.AuthService
	UserService            service.UserService
	Middleware             *middleware.Middleware
	HandlerComponents      *handler.Components
	AccessJWTManager       auth_jwt.Manager
	RefreshJWTManager      refresh_jwt.Manager
	VerificationJWTManager verify_jwt.Manager
	CdnUrlSigner           cdn.URLSigner
}

// MapV1Routes maps version 1 routes to the provided router.
// It initializes handlers for authentication and user services, sets up JWT and verification middlewares,
// and maps the routes for authentication and user endpoints.
func MapV1Routes(args RouterArgs, router chi.Router) {
	authHandler := auth.NewHandler(
		args.HandlerComponents,
		args.AuthService,
	)

	userHandler := user.NewHandler(
		args.HandlerComponents,
		args.UserService,
		args.CdnUrlSigner,
	)

	accessJWTMiddleware := args.Middleware.NewJWT(args.AccessJWTManager)
	onlyVerifiedMiddleware := args.Middleware.NewAllowVerified(true)

	router.Route("/v1", func(r chi.Router) {
		auth.MapRoutes(authHandler, args.AccessJWTManager, args.RefreshJWTManager, args.VerificationJWTManager, r)
		user.MapRoutes(args.Middleware, userHandler, accessJWTMiddleware, onlyVerifiedMiddleware, r)
	})
}
