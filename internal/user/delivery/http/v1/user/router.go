package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/enum"
)

// TODO: manage missing routes

var (
	maxFileSize int64 = 10 << 20
)

func MapRoutes(
	middlewareFactory *middleware.MiddlewareFactory,
	jWTAccessMiddleware func(http.Handler) http.Handler,
	onlyVerifiedMiddleware func(http.Handler) http.Handler,
	router chi.Router,
) {
	router.Route("/user", func(r chi.Router) {
		r.Use(
			jWTAccessMiddleware,
			onlyVerifiedMiddleware,
			middlewareFactory.NewAllowSelfOrRole(enum.UserRoleADMIN, enum.UserRoleMODERATOR),
		)

		r.Group(func(r chi.Router) {
			r.Use(
				middlewareFactory.NewAllowFilesAmount(maxFileSize, "image", 1),
				middlewareFactory.NewAllowContentType(maxFileSize, "image", "image/jpeg", "image/png"),
			)
		})

	})
}