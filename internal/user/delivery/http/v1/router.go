package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/user"
	"github.com/hexley21/fixup/internal/user/jwt/refresh_jwt"
	refresh_middleware "github.com/hexley21/fixup/internal/user/jwt/refresh_jwt/middleware"
	"github.com/hexley21/fixup/internal/user/jwt/verify_jwt"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/http/handler"
)

type RouterArgs struct {
	AuthService            service.AuthService
	UserService            service.UserService
	Middleware             *middleware.Middleware
	HandlerComponents      *handler.Components
	AccessJWTManager       auth_jwt.JWTManager
	RefreshJWTManager      refresh_jwt.JWTManager
	VerificationJWTManager verify_jwt.JWTManager
}

func MapV1Routes(args RouterArgs, router chi.Router) {
	authHandlermiddleware := auth.NewHandler(
		args.HandlerComponents,
		args.AuthService,
	)

	userHandlermiddleware := user.NewHandler(
		args.HandlerComponents,
		args.UserService,
	)

	accessJWTMiddleware := args.Middleware.NewJWT(args.AccessJWTManager)
	onlyVerifiedMiddleware := args.Middleware.NewAllowVerified(true)
	refreshJWTMiddleware := refresh_middleware.NewJWT(args.HandlerComponents.Writer, args.RefreshJWTManager)

	router.Route("/v1", func(r chi.Router) {
		auth.MapRoutes(args.Middleware, authHandlermiddleware, refreshJWTMiddleware, args.AccessJWTManager, args.RefreshJWTManager, args.VerificationJWTManager, r)
		user.MapRoutes(args.Middleware, userHandlermiddleware, accessJWTMiddleware, onlyVerifiedMiddleware, r)
	})
}
