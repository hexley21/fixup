package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/pkg/rest"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type userHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *userHandler {
	return &userHandler{
		service,
	}
}

func (h *userHandler) FindUserById() echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := h.service.FindUserById(context.Background(), c.Param("id"))
		if errors.Is(err, pgx.ErrNoRows) {
			return rest.NotFound
		}
		if errors.Is(err, strconv.ErrSyntax) {
			return rest.BadRequest
		}
		if err != nil {
			return rest.InternalServerError
		}

		return c.JSON(http.StatusOK, rest.NewApiResponse(user))
	}
}
