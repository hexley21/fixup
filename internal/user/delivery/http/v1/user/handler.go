package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	"github.com/jackc/pgx/v5"
)

// TODO: manage who can access certain endpoint

var (
	maxPfpSize int64 = 1 << 20
)

type HandlerFactory struct {
	*handler.Components
	service service.UserService
}

func NewFactory(components *handler.Components, service service.UserService) *HandlerFactory {
	return &HandlerFactory{
		Components: components,
		service:    service,
	}
}

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
func (f *HandlerFactory) FindUserById(w http.ResponseWriter, r *http.Request) {
	id, errResp := ctx_util.GetParamId(r.Context())
	if errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	user, err := f.service.FindUserById(r.Context(), id)
	if err != nil {
		if errors.Is(err, pg_error.ErrNotFound) {
			f.Writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
			return
		}

		f.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	f.Logger.Infof("Fetch user - U-ID: %d", id)
	f.Writer.WriteData(w, http.StatusOK, user)
}

// @Summary Find user profile by ID
// @Description Retrieve profile details by user ID
// @Tags profile
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} rest.apiResponse[dto.Profile] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Security access_token
// @Router /profile/{id} [get]
func (f *HandlerFactory) FindUserProfileById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, errResp := ctx_util.GetParamId(r.Context())
	if errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	profile, err := f.service.FindUserProfileById(ctx, id)
	if err != nil {
		if errors.Is(err, pg_error.ErrNotFound) {
			f.Writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
			return
		}

		f.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	f.Logger.Infof("Fetch user profile - U-ID: %d", id)
	f.Writer.WriteData(w, http.StatusOK, profile)
}

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
func (f *HandlerFactory) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	id, errResp := ctx_util.GetParamId(r.Context())
	if errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	form, errResp := f.Binder.BindMultipartForm(r, maxPfpSize)
	if errResp != nil {
		f.Writer.WriteError(w, rest.NewReadFileError(errResp))
		return
	}

	formFile := form.File["image"]
	if len(formFile) < 1 {
		f.Writer.WriteError(w, rest.NewBadRequestError(nil, rest.MsgNoFile))
		return
	}

	imageFile := formFile[0]

	file, err := imageFile.Open()
	if err != nil {
		f.Writer.WriteError(w, rest.NewReadFileError(err))
		return
	}
	defer file.Close()

	err = f.service.SetProfilePicture(r.Context(), id, file, "", imageFile.Size, imageFile.Header.Get("Content-Type"))
	if err != nil {
		if errors.Is(err, pg_error.ErrNotFound) {
			f.Writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
			return
		}

		f.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	f.Logger.Infof("Upload user profile picture - U-ID: %d", id)
	f.Writer.WriteNoContent(w, http.StatusNoContent)
}

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
func (f *HandlerFactory) UpdateUserData(w http.ResponseWriter, r *http.Request) {
	id, errResp := ctx_util.GetParamId(r.Context())
	if errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	dto := new(dto.UpdateUser)
	if errResp := f.Binder.BindJSON(r, dto); errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	if errResp := f.Validator.Validate(dto); errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	user, err := f.service.UpdateUserDataById(r.Context(), id, *dto)
	if err != nil {
		if errors.Is(err, pg_error.ErrNotFound) {
			f.Writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
			return
		}

		if errors.Is(err, pgx.ErrNoRows) {
			f.Writer.WriteError(w, rest.NewBadRequestError(err, app_error.MsgNoChanges))
			return
		}

		f.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	f.Logger.Infof("Update user data - U-ID: %d", id)
	f.Writer.WriteData(w, http.StatusOK, user)
}

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
func (f *HandlerFactory) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, errResp := ctx_util.GetParamId(r.Context())
	if errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	err := f.service.DeleteUserById(r.Context(), id)
	if err != nil {
		if errors.Is(err, pg_error.ErrNotFound) {
			f.Writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
			return
		}

		f.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	f.Logger.Infof("Delete user - U-ID: %d", id)
	f.Writer.WriteNoContent(w, http.StatusNoContent)
}

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
// @Router /user/me/change-password [patch]
func (f *HandlerFactory) ChangePassword(w http.ResponseWriter, r *http.Request) {
	id, errResp := ctx_util.GetJWTId(r.Context())
	if errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		f.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	dto := new(dto.UpdatePassword)
	errResp = f.Binder.BindJSON(r, dto)
	if errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	errResp = f.Validator.Validate(dto)
	if errResp != nil {
		f.Writer.WriteError(w, errResp)
		return
	}

	err = f.service.ChangePassword(r.Context(), userId, *dto)
	if err != nil {
		if errors.Is(err, pg_error.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			f.Writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
			return
		}
		if errors.Is(err, hasher.ErrPasswordMismatch) {
			f.Writer.WriteError(w, rest.NewUnauthorizedError(err, app_error.MsgIncorrectPassword))
			return
		}

		f.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	f.Logger.Infof("Change user password - U-ID: %d", userId)
	f.Writer.WriteNoContent(w, http.StatusNoContent)
}
