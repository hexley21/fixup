package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/category_type"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/pkg/http/handler"
)

type RouterArgs struct {
	CategoryTypeService service.CategoryTypeService
	MiddlewareFactory   *middleware.MiddlewareFactory
	HandlerComponents   *handler.Components
	AccessJWTManager    auth_jwt.JWTManager
}

func MapV1Routes(args RouterArgs, router chi.Router) {
	accessJWTMiddleware := args.MiddlewareFactory.NewJWT(args.AccessJWTManager)
	onlyVerifiedMiddleware := args.MiddlewareFactory.NewAllowVerified(true)
	onlyAdminMiddleware := args.MiddlewareFactory.NewAllowRoles(enum.UserRoleADMIN)

	categoryTypesHandler := category_type.NewCategoryTypeHandler(
		args.HandlerComponents,
		args.CategoryTypeService,
	)

	router.Route("/v1", func(r chi.Router) {
		category_type.MapRoutes(categoryTypesHandler, accessJWTMiddleware, onlyVerifiedMiddleware, onlyAdminMiddleware, r)
	})
}
