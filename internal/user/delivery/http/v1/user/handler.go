package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// TODO: manage contextes

type userHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *userHandler {
	return &userHandler{
		service: service,
	}
}

func (h *userHandler) findUserById(c echo.Context) error {
	id, err := ctxutil.GetParamId(c)
	if err != nil {
		return err
	}

	user, err := h.service.FindUserById(context.Background(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rest.NewNotFoundError(err, "User not found")
		}
		return rest.NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, rest.NewApiResponse(user))
}

func (h *userHandler) uploadProfilePicture(c echo.Context) error {
	id, err := ctxutil.GetParamId(c)
	if err != nil {
		return err
	}

	form, err := c.MultipartForm()
	if err != nil {
		return rest.NewReadFileError(err)
	}

	imageFile := form.File["image"][0]

	src, err := imageFile.Open()
	if err != nil {
		return rest.NewReadFileError(err)
	}
	defer src.Close()

	err = h.service.SetProfilePicture(context.Background(), id, src, "", imageFile.Size, imageFile.Header.Get("Content-Type"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rest.NewNotFoundError(err, "User not found")
		}
		return rest.NewInternalServerError(err)
	}

	return c.NoContent(http.StatusOK)
}

func (h *userHandler) updateUserData(c echo.Context) error {
	id, err := ctxutil.GetParamId(c)
	if err != nil {
		return err
	}

	dto := new(dto.UpdateUser)
	if err := c.Bind(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	if err := c.Validate(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	user, err := h.service.UpdateUserDataById(context.Background(), id, *dto)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rest.NewNotFoundError(err, "User not found")
		}
		return rest.NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, rest.NewApiResponse(user))
}

// deleteUser godoc
// @Summary Delete a user
// @Description Delete a user by ID or the currently authenticated user if "me" is provided
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID or 'me'"
// @Success 204 {object} nil "No Content"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /users/{id} [delete]
func (h *userHandler) deleteUser(c echo.Context) error {
	id, err := ctxutil.GetParamId(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteUserById(context.Background(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rest.NewNotFoundError(err, "User was not found")
		}
		return rest.NewInternalServerError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
