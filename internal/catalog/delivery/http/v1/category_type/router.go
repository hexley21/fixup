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
		r.Group(func(r chi.Router) {
			r.Use(jWTAccessMiddleware, onlyVerifiedMiddleware, onlyAdminMiddleware)

			r.Post("/", h.Create)
			r.Patch("/{type_id}", h.Update)
			r.Delete("/{type_id}", h.Delete)
		})

		r.Get("/", h.List)
		r.Get("/{type_id}", h.Get)
	})
}
