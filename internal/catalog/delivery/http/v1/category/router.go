package category

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
	router.Route("/categories", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jWTAccessMiddleware, onlyVerifiedMiddleware, onlyAdminMiddleware)

			r.Post("/", h.CreateCategory)
			r.Patch("/{id}", h.PatchCategoryById)
			r.Delete("/{id}", h.DeleteCategoryById)
		})

		r.Get("/", h.GetCategories)
		r.Get("/{id}", h.GetCategoryById)
	})

	router.Get("/category-types/{id}/categories", h.GetCategoriesByTypeId)
}
