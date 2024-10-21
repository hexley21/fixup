package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/common/middleware"
)

var (
	maxFileSize int64 = 10 << 20
)

func MapRoutes(
	mw *middleware.Middleware,
	h *Handler,
	jWTAccessMiddleware func(http.Handler) http.Handler,
	onlyVerifiedMiddleware func(http.Handler) http.Handler,
	router chi.Router,
) {
	router.Route("/user", func(r chi.Router) {
		r.Use(
			jWTAccessMiddleware,
			onlyVerifiedMiddleware,
			NewUserMiddleware(h.Writer).AllowSelfOrRole(enum.UserRoleADMIN, enum.UserRoleMODERATOR),
		)

		r.Get("/{user_id}", h.Get)
		r.Patch("/{user_id}", h.UpdatePersonalInfo)
		r.Delete("/{user_id}", h.Delete)

		r.Group(func(r chi.Router) {
			r.Use(
				mw.NewAllowFilesAmount(maxFileSize, "image", 1),
				mw.NewAllowContentType(maxFileSize, "image", "image/jpeg", "image/png"),
			)
			r.Patch("/{user_id}/pfp", h.UploadProfilePicture)
		})
	})

	router.Group(func(r chi.Router) {
		r.Use(jWTAccessMiddleware)
		r.Patch("/me/change-password", h.ChangePassword)
	})

	// router.Get("/profile/{id}", h.FindUserProfileById)
}
