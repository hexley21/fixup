package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/enum"
)

var (
	maxFileSize int64 = 10 << 20
)

func MapRoutes(
	mf *middleware.MiddlewareFactory,
	hf *HandlerFactory,
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

		r.Get("/{id}", hf.FindUserById)
		r.Patch("/{id}", hf.UpdateUserData)
		r.Delete("/{id}", hf.DeleteUser)

		r.Group(func(r chi.Router) {
			r.Use(
				mf.NewAllowFilesAmount(maxFileSize, "image", 1),
				mf.NewAllowContentType(maxFileSize, "image", "image/jpeg", "image/png"),
			)
			r.Patch("/{id}/pfp", hf.UploadProfilePicture)
		})
	})

	router.Group(func(r chi.Router) {
		r.Use(jWTAccessMiddleware)
		r.Patch("/me/change-password", hf.ChangePassword)
	})

	router.Patch("/profile/{id}", hf.FindUserProfileById)
}
