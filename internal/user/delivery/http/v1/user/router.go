package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/common/enum"
)

var (
	maxFileSize int64 = 10 << 20
)

func MapRoutes(
	mf *middleware.MiddlewareFactory,
	h *Handler,
	jWTAccessMiddleware func(http.Handler) http.Handler,
	onlyVerifiedMiddleware func(http.Handler) http.Handler,
	router chi.Router,
) {
	router.Route("/user", func(r chi.Router) {
		r.Use(
			jWTAccessMiddleware,
			onlyVerifiedMiddleware,
			mf.NewAllowSelfOrRole(enum.UserRoleADMIN, enum.UserRoleMODERATOR),
		)

		r.Get("/{id}", h.FindUserById)
		r.Patch("/{id}", h.UpdateUserData)
		r.Delete("/{id}", h.DeleteUser)

		r.Group(func(r chi.Router) {
			r.Use(
				mf.NewAllowFilesAmount(maxFileSize, "image", 1),
				mf.NewAllowContentType(maxFileSize, "image", "image/jpeg", "image/png"),
			)
			r.Patch("/{id}/pfp", h.UploadProfilePicture)
		})
	})

	router.Group(func(r chi.Router) {
		r.Use(jWTAccessMiddleware)
		r.Patch("/me/change-password", h.ChangePassword)
	})

	router.Patch("/profile/{id}", h.FindUserProfileById)
}
