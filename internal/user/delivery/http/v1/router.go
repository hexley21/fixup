package v1

import (
	"github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/user"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/labstack/echo/v4"
)

type V1RouterArgs struct {
	AuthService service.AuthService
	UserService service.UserService
	CfgJwt config.JWT
}

func MapV1Routes(echo *echo.Echo, args V1RouterArgs) *echo.Group {
	accessAuthJwt := jwt.NewAuthJwtImpl(args.CfgJwt.AccessSecret, args.CfgJwt.AccessTTL)
	refreshAuthJwt := jwt.NewAuthJwtImpl(args.CfgJwt.RefreshSecret, args.CfgJwt.RefreshTTL)

	verificationJwt := verifier.NewVerificationJwt(args.CfgJwt.VerificationSecret, args.CfgJwt.VerificationTTL)

	accessJwtMiddleware := middleware.JWT(accessAuthJwt)
	onlyVerifiedMiddleware := middleware.AllowVerified(true)

	v1Group := echo.Group("/v1")

	auth.NewAuthHandler(args.AuthService).MapRoutes(v1Group, accessAuthJwt, refreshAuthJwt, verificationJwt)
	user.NewUserHandler(args.UserService).MapRoutes(v1Group, accessJwtMiddleware, onlyVerifiedMiddleware)

	return v1Group
}
