package user

import (
	"github.com/go-chi/chi/v5"
	"github.com/hexley21/handy/pkg/http/handler"
)

func (h *userHandler) MapRoutes(f *handler.HandlerFactory) chi.Router{
	return h.router.Route("/users", func(r chi.Router) {
		r.Get("/{id}", f.NewHandlerFunc(h.FindUserById()))
	})
}