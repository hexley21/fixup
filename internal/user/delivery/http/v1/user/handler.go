package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/handy/internal/user/service"
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

func (h *userHandler) FindUserById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := h.service.FindUserById(context.Background(), chi.URLParam(r, "id"))
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.ErrorCause(rest.WriteNotFoundError(w), err)
			return
		}
		if errors.Is(err, strconv.ErrSyntax) {
			h.logger.ErrorCause(rest.WriteBadRequestError(w), err)
			return
		}
		if err != nil {
			h.logger.ErrorCause(rest.WriteInternalServerError(w), err)
			return
		}

		if err := rest.WriteOkResponse(w, user); err != nil {
			h.logger.Error(err)
		}
	}
}
