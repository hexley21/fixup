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

			r.Post("/", h.Create)
			r.Patch("/{category_id}", h.Update)
			r.Delete("/{category_id}", h.Delete)
		})

		r.Get("/", h.List)
		r.Get("/{category_id}", h.Get)
	})

	router.Get("/category-types/{type_id}/categories", h.ListByTypeId)
}
