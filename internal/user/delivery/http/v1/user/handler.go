package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/rest"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type userHandler struct {
	service      service.UserService
}

func NewUserHandler(service service.UserService) *userHandler {
	return &userHandler{
		service:      service,
	}
}

func (h *userHandler) findUserById(c echo.Context) error {
	idParam := c.Param("id")

	if idParam == "me" {
		user, ok := c.Get("user").(jwt.UserClaims)
		if !ok {
			return rest.ErrJwtNotImplemented
		}
		idParam = user.ID
	}

	userId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return rest.NewInternalServerError(err)
	}

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

func (h *userHandler) uploadProfilePicture(c echo.Context) error {
	idParam := c.Param("id")

	if idParam == "me" {
		user, ok := c.Get("user").(jwt.UserClaims)
		if !ok {
			return rest.ErrJwtNotImplemented
		}
		idParam = user.ID
	}

	userId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return rest.NewInternalServerError(err)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return rest.NewReadFileError(err)
	}

	files := form.File["image"]

	if len(files) > 1 {
		return rest.ErrTooManyFiles
	}
	if len(files) < 1 {
		return rest.ErrNoFile
	}

	imageFile := files[0]

	src, err := imageFile.Open()
	if err != nil {
		return rest.NewReadFileError(err)
	}

	defer src.Close()

	contentType := imageFile.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		return rest.ErrInvalidFileType
	}

	err = h.service.SetProfilePicture(context.Background(), userId, src, "", imageFile.Size, contentType)
	if err != nil {
		return rest.NewInternalServerError(err)
	}

	return c.NoContent(http.StatusOK)
}

func (h *userHandler) updateUserData(c echo.Context) error {
	idParam := c.Param("id")

	if idParam == "me" {
		user, ok := c.Get("user").(jwt.UserClaims)
		if !ok {
			return rest.ErrJwtNotImplemented
		}
		idParam = user.ID
	}
	
	userId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return rest.NewInternalServerError(err)
	}

	dto := new(dto.UpdateUser)
	if err := c.Bind(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	if err := c.Validate(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	user, err := h.service.UpdateUserDataById(context.Background(), userId, *dto)
	if err != nil {
		return rest.NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, rest.NewApiResponse(user))

}
