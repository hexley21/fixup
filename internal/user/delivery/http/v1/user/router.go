package user

import "github.com/go-chi/chi/v5"

func (h *userHandler) MapRoutes() chi.Router{
	return h.router.Route("/users", func(r chi.Router) {
		r.Get("/{id}", h.FindUserById())
	})
}