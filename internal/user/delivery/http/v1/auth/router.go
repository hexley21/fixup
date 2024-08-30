package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/hexley21/handy/pkg/http/handler"
)

func (h *authHandler) MapRoutes(f *handler.HandlerFactory) chi.Router {
	return h.router.Route("/register", func(r chi.Router) {
		r.Get("/customer", f.NewHandlerFunc(h.RegisterCustomer()))
	})
}
