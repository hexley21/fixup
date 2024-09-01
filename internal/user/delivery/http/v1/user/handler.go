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
		userId := c.Param("id")

		user, err := h.service.FindUserById(context.Background(), userId)
		if errors.Is(err, pgx.ErrNoRows) {
			return rest.NewNotFoundError(err, "User not found")
		}
		if errors.Is(err, strconv.ErrSyntax) {
			return rest.NewInvalidArgumentsError(err)
		}
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, rest.NewApiResponse(user))
	}
}
