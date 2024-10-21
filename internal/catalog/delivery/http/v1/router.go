package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/category"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/category_type"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/subcategory"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/http/handler"
)

type RouterArgs struct {
	CategoryTypeService service.CategoryTypeService
	CategoryService     service.CategoryService
	SubcategoryService  service.SubcategoryService
	Middleware          *middleware.Middleware
	HandlerComponents   *handler.Components
	AccessJWTManager    auth_jwt.Manager
	PaginationConfig    *config.Pagination
}

func MapV1Routes(args RouterArgs, router chi.Router) {
	accessJWTMiddleware := args.Middleware.NewJWT(args.AccessJWTManager)
	onlyVerifiedMiddleware := args.Middleware.NewAllowVerified(true)
	onlyAdminMiddleware := args.Middleware.NewAllowRoles(enum.UserRoleADMIN)

	categoryTypesHandler := category_type.NewHandler(
		args.HandlerComponents,
		args.CategoryTypeService,
		args.PaginationConfig.LargePages,
		args.PaginationConfig.XLargePages,
	)

	categoryHandler := category.NewHandler(
		args.HandlerComponents,
		args.CategoryService,
	)

	subcategoryHandler := subcategory.NewHandler(
		args.HandlerComponents,
		args.SubcategoryService,
		args.PaginationConfig.LargePages,
		args.PaginationConfig.XLargePages,
	)

	router.Route("/v1", func(r chi.Router) {
		category_type.MapRoutes(categoryTypesHandler, accessJWTMiddleware, onlyVerifiedMiddleware, onlyAdminMiddleware, r)
		category.MapRoutes(categoryHandler, accessJWTMiddleware, onlyVerifiedMiddleware, onlyAdminMiddleware, r)
		subcategory.MapRoutes(subcategoryHandler, accessJWTMiddleware, onlyAdminMiddleware, onlyAdminMiddleware, r)
	})
}
