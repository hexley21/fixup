package category_type

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/internal/common/util/request_util"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	MsgCategoryTypeNotFound = "Category type not found"
)

type Handler struct {
	*handler.Components
	service service.CategoryTypeService
}

func NewHandler(handlerComponents *handler.Components, service service.CategoryTypeService) *Handler {
	return &Handler{
		Components: handlerComponents,
		service:    service,
	}
}

// @Summary Create a new category type
// @Description Creates a new category type with the provided data.
// @Tags CategoryType
// @Param dto body dto.CreateCategoryTypeDTO true "Category type data"
// @Success 201 {object} rest.ApiResponse[dto.CategoryTypeDTO] "Created - Successfully created the category type"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while creating the category type"
// @Router /category-types [post]
// @Security access_token
func (h *Handler) CreateCategoryType(w http.ResponseWriter, r *http.Request) {
	var createDTO dto.CreateCategoryTypeDTO
	errResp := h.Binder.BindJSON(r, &createDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	errResp = h.Validator.Validate(createDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	categoryType, err := h.service.CreateCategoryType(r.Context(), createDTO)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			h.Writer.WriteError(w, rest.NewConflictError(err, app_error.MsgNameAlreadyTaken))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Create category type: %s, ID: %s", categoryType.Name, categoryType.ID)
	h.Writer.WriteData(w, http.StatusCreated, categoryType)
}

// @Summary Retrieve a category types
// @Description Retrieves a category type range
// @Tags CategoryType
// @Param page query int true "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} rest.ApiResponse[[]dto.CategoryTypeDTO] "OK - Successfully retrieved the category types"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving the category type"
// @Router /category-types [get]
// @Security access_token
func (h *Handler) GetCategoryTypes(w http.ResponseWriter, r *http.Request) {
	errResp, page, perPage := request_util.ParsePagination(r)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	categoryTypes, err := h.service.GetCategoryTypes(r.Context(), int32(page), int32(perPage))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, MsgCategoryTypeNotFound))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Fetch category types - elements: %d", len(categoryTypes))
	h.Writer.WriteData(w, http.StatusOK, categoryTypes)
}

// @Summary Retrieve a category type by ID
// @Description Retrieves a category type specified by the ID.
// @Tags CategoryType
// @Param id path int true "The ID of the category type to retrieve"
// @Success 200 {object} rest.ApiResponse[dto.CategoryTypeDTO] "OK - Successfully retrieved the category type"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving the category type"
// @Router /category-types/{id} [get]
// @Security access_token
func (h *Handler) GetCategoryTypeById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
		return
	}

	categoryTypeDTO, err := h.service.GetCategoryTypeById(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, MsgCategoryTypeNotFound))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Fetch category type: %s, ID: %s", categoryTypeDTO.Name, categoryTypeDTO.ID)
	h.Writer.WriteData(w, http.StatusOK, categoryTypeDTO)
}

// @Summary Update a category type by ID
// @Description Updates a category type specified by the ID.
// @Tags CategoryType
// @Param id path int true "The ID of the category type to update"
// @Param dto body dto.PatchCategoryTypeDTO true "Category type data"
// @Success 200 {object} rest.ApiResponse[dto.CategoryTypeDTO] "OK - Successfully updated the category type"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while updating the category type"
// @Router /category-types/{id} [patch]
// @Security access_token
func (h *Handler) PatchCategoryTypeById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
		return
	}

	var patchDto dto.PatchCategoryTypeDTO
	errResp := h.Binder.BindJSON(r, &patchDto)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	errResp = h.Validator.Validate(patchDto)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	err = h.service.UpdateCategoryTypeById(r.Context(), int32(id), patchDto)
	if err != nil {
		if errors.Is(err, pg_error.ErrNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, MsgCategoryTypeNotFound))
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			h.Writer.WriteError(w, rest.NewConflictError(err, app_error.MsgNameAlreadyTaken))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Patch category type: %s, ID: %d", patchDto.Name, id)
	h.Writer.WriteData(w, http.StatusOK, dto.CategoryTypeDTO{ID: strconv.Itoa(id), Name: patchDto.Name})
}

// @Summary Delete a category type by ID
// @Description Deletes a category type specified by the ID.
// @Tags CategoryType
// @Param id path int true "The ID of the category type to delete"
// @Success 204 {string} string "No Content - Successfully deleted the category type"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while deleting the category type"
// @Router /category-types/{id} [delete]
// @Security access_token
func (h *Handler) DeleteCategoryTypeById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
		return
	}

	err = h.service.DeleteCategoryTypeById(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, pg_error.ErrNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, MsgCategoryTypeNotFound))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Delete category type - ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}
