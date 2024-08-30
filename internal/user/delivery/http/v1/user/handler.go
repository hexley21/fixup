package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/pkg/http/handler"
	"github.com/hexley21/handy/pkg/logger"
	"github.com/hexley21/handy/pkg/rest"
	"github.com/jackc/pgx/v5"
)

type userHandler struct {
	router  chi.Router
	logger  logger.Logger
	service service.UserService
}

func NewUserHandler(
	router chi.Router,
	logger logger.Logger,
	service service.UserService,
) *userHandler {
	return &userHandler{
		router,
		logger,
		service,
	}
}

func (h *userHandler) FindUserById() handler.ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, err := h.service.FindUserById(context.Background(), chi.URLParam(r, "id"))
		if errors.Is(err, pgx.ErrNoRows) {
			return rest.NotFound
		}
		if errors.Is(err, strconv.ErrSyntax) {
			return rest.BadRequest
		}
		if err != nil {
			return rest.InternalServerError
		}

		return rest.WriteResponse(w, user, http.StatusOK)
	}
}
