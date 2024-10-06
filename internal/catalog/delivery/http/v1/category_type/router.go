package category_type

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func MapRoutes(
	h *Handler,
	jWTAccessMiddleware func(http.Handler) http.Handler,
	onlyVerifiedMiddleware func(http.Handler) http.Handler,
	onlyAdminMiddleware func(http.Handler) http.Handler,
	router chi.Router,
) {
	router.Route("/category-types", func(r chi.Router) {
		r.Use(
			jWTAccessMiddleware,
			onlyAdminMiddleware,
			onlyAdminMiddleware,
		)

		r.Post("/", h.CreateCategoryType)
		r.Get("/", h.GetCategoryTypes)
		r.Get("/{id}", h.GetCategoryTypeById)
		r.Patch("/{id}", h.PatchCategoryTypeById)
		r.Delete("/{id}", h.DeleteCategoryTypeById)
	})
}
