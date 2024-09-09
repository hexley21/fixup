package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/hasher"
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

// findUserById godoc
// @Summary Find user by ID
// @Description Retrieve user details by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} rest.apiResponse[dto.User] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Security access_token
// @Router /users/{id} [get]
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

// uploadProfilePicture godoc
// @Summary Upload profile picture
// @Description Upload a profile picture for the user by ID
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "User ID"
// @Param image formData file true "Profile picture file"
// @Success 204 "No Content"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Security access_token
// @Router /users/{id}/pfp [patch]
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

	return c.NoContent(http.StatusNoContent)
}

// updateUserData godoc
// @Summary Update user data
// @Description Update user data by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dto.UpdateUser true "User data"
// @Success 200 {object} rest.apiResponse[dto.User] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Security access_token
// @Router /users/{id} [patch]
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
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "NotFound"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Security access_token
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

// updatePassword godoc
// @Summary Update user password
// @Description Update the password of an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body dto.UpdatePassword true "Update Password DTO"
// @Success 204 "No Content"
// @Failure 400 {object} rest.ErrorResponse "Invalid arguments"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 404 {object} rest.ErrorResponse "User not found"
// @Failure 500 {object} rest.ErrorResponse "Internal server error"
// @Router /v1/user/{id}/change-password [patch]
func (h *userHandler) updatePassword(c echo.Context) error {
	id, err := ctxutil.GetJwtId(c)
	if err != nil {
		return err
	}

	userId, err := strconv.ParseInt(id,  10, 64)
	if err != nil {
		return rest.NewInternalServerError(err)
	}

	dto := new(dto.UpdatePassword)
	if err := c.Bind(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	if err := c.Validate(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	if err := h.service.ChangePassword(context.Background(), userId, *dto); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rest.NewNotFoundError(err, "User was not found")
		}
		if errors.Is(err, hasher.ErrPasswordMismatch) {
			return rest.NewUnauthorizedError(err, "Old password is not correct")
		}
		return rest.NewInternalServerError(err)
	}

	return c.NoContent(http.StatusNoContent)
}