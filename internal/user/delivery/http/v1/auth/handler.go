package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/pkg/http/handler"
	"github.com/hexley21/handy/pkg/logger"
	"github.com/hexley21/handy/pkg/rest"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type authHandler struct {
	router  chi.Router
	logger  logger.Logger
	service service.AuthService
}

func NewAuthHandler(
	router chi.Router,
	logger logger.Logger,
	service service.AuthService,
) *authHandler {
	return &authHandler{
		router,
		logger,
		service,
	}
}

func (h *authHandler) RegisterCustomer() handler.ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var dto dto.RegisterUser

		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			return err
		}

		user, err := h.service.RegisterCustomer(context.Background(), &dto)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return err
		}
		if err != nil {
			return err
		}

		if err := rest.WriteOkResponse(w, user); err != nil {
			h.logger.Error(err)
			return err
		}
		return nil
	}
}
