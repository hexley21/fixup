package user

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/domain"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/rest"
)

// TODO: manage who can access certain endpoint & add profile endpoints

var (
	maxPfpSize int64 = 1 << 20
)

type Handler struct {
	*handler.Components
	service service.UserService
}

func NewHandler(components *handler.Components, service service.UserService) *Handler {
	return &Handler{
		Components: components,
		service:    service,
	}
}

// Get
// @Summary Find user by ID
// @Description Retrieve user details by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} rest.ApiResponse[dto.User] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Security access_token
// @Router /users/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(paramIdKey).(int64)
	if !ok {
		h.Writer.WriteError(w, ErrParamIdNotSet)
		return
	}

	user, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Fetch user - U-ID: %d", id)
	h.Writer.WriteData(w, http.StatusOK, user)
}

// UploadProfilePicture
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
// @Router /users/{id}/pfp [put]
func (h *Handler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(paramIdKey).(int64)
	if !ok {
		h.Writer.WriteError(w, ErrParamIdNotSet)
		return
	}

	form, errResp := h.Binder.BindMultipartForm(r, maxPfpSize)
	if errResp != nil {
		h.Writer.WriteError(w, rest.NewReadFileError(errResp))
		return
	}

	formFile := form.File["image"]
	if len(formFile) < 1 {
		h.Writer.WriteError(w, rest.NewBadRequestError(rest.ErrNoFile))
		return
	}

	imageFile := formFile[0]

	file, err := imageFile.Open()
	if err != nil {
		h.Writer.WriteError(w, rest.NewReadFileError(err))
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			h.Logger.Errorf("failed to close file: %v", err)
		}
	}(file)

	err = h.service.UpdateProfilePicture(
		r.Context(),
		id,
		file,
		"",
		imageFile.Size,
		imageFile.Header.Get("Content-Type"),
	)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Upload user profile picture - U-ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}

// UpdatePersonalInfo
// @Summary Update user data
// @Description Update user data by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dto.UpdateUser true "User data"
// @Success 200 {object} rest.ApiResponse[dto.User] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Security access_token
// @Router /users/{id} [patch]
func (h *Handler) UpdatePersonalInfo(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(paramIdKey).(int64)
	if !ok {
		h.Writer.WriteError(w, ErrParamIdNotSet)
		return
	}
	var updateDTO dto.UpdateUser
	if errResp := h.Binder.BindJSON(r, &updateDTO); errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	if errResp := h.Validator.Validate(updateDTO); errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	user, err := h.service.UpdatePersonalInfo(
		r.Context(),
		id,
		domain.NewUserPersonalInfo(updateDTO.Email, updateDTO.PhoneNumber, updateDTO.FirstName, updateDTO.LastName),
	)
	if err != nil {
		if errors.Is(err, service.ErrUserNotUpdated) {
			h.Writer.WriteError(w, rest.NewBadRequestError(err))
			return
		}

		if errors.Is(err, service.ErrUserNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Update user data - U-ID: %d", id)
	h.Writer.WriteData(w, http.StatusOK, user)
}

// Delete
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
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(paramIdKey).(int64)
	if !ok {
		h.Writer.WriteError(w, ErrParamIdNotSet)
		return
	}
	
	err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Delete user - U-ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}

// ChangePassword
// @Summary Update user password
// @Description Update the password of an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param password 	body dto.UpdatePassword true "Update Password DTO"
// @Success 204 "No Content"
// @Failure 400 {object} rest.ErrorResponse "Invalid arguments"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 404 {object} rest.ErrorResponse "User not found"
// @Failure 500 {object} rest.ErrorResponse "Internal server error"
// @Router /user/me/change-password [patch]
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	jwt, ok := r.Context().Value(auth_jwt.AuthJWTKey).(auth_jwt.UserData)
	if !ok {
		h.Writer.WriteError(w, auth_jwt.ErrJWTNotSet)
		return
	}

	id, err := strconv.ParseInt(jwt.ID, 10, 64)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	var updateDTO dto.UpdatePassword
	errResp := h.Binder.BindJSON(r, &updateDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	errResp = h.Validator.Validate(updateDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	err = h.service.UpdatePassword(r.Context(), id, updateDTO.OldPassword, updateDTO.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}
		if errors.Is(err, service.ErrIncorrectPassword) {
			h.Writer.WriteError(w, rest.NewUnauthorizedError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Change user password - U-ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}
