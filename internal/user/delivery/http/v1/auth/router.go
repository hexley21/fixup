package auth

import "github.com/go-chi/chi/v5"

func (h *authHandler) MapRoutes() chi.Router {
	h.router.Route("/register", func(r chi.Router) {
		r.Get("/customer", h.RegisterCustomer())
	})
	
	return h.router
}