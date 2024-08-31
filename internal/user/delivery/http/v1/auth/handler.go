package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/pkg/rest"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
)

type authHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *authHandler {
	return &authHandler{
		service,
	}
}

func (h *authHandler) RegisterCustomer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var dto dto.RegisterUser
		if err := c.Bind(&dto); err != nil {
			return err
		}

		user, err := h.service.RegisterCustomer(context.Background(), &dto)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return rest.Conflict
		}
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, rest.NewApiResponse(user))
	}
}
