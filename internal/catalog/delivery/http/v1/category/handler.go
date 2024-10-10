package category

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
	MsgCategoryNotFound = "Category not found"
)

type Handler struct {
	*handler.Components
	service service.CategoryService
}

func NewHandler(handlerComponents *handler.Components, service service.CategoryService) *Handler {
	return &Handler{
		Components: handlerComponents,
		service:    service,
	}
}

// @Summary Create a new category
// @Description Creates a new category with the provided data.
// @Tags Category
// @Param dto body dto.CreateCategoryDTO true "Category data"
// @Success 201 {object} rest.ApiResponse[dto.CategoryDTO] "Created - Successfully created the category"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while creating the category"
// @Router /categories [post]
// @Security access_token
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var createDTO dto.CreateCategoryDTO
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

	category, err := h.service.CreateCategory(r.Context(), createDTO)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.RaiseException {
			h.Writer.WriteError(w, rest.NewConflictError(err, app_error.MsgNameAlreadyTaken))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Create category: %s, ID: %s", category.Name, category.ID)
	h.Writer.WriteData(w, http.StatusCreated, category)
}

// @Summary Retrieve categories
// @Description Retrieves a category range
// @Tags Category
// @Param page query int true "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} rest.ApiResponse[[]dto.CategoryDTO] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving categories"
// @Router /categories [get]
// @Security access_token
func (h *Handler) GetCategoryies(w http.ResponseWriter, r *http.Request) {
	errResp, page, perPage := request_util.ParsePagination(r)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	category, err := h.service.GetCategories(r.Context(), int32(page), int32(perPage))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, MsgCategoryNotFound))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Fetch category  - elements: %d", len(category))
	h.Writer.WriteData(w, http.StatusOK, category)
}

// @Summary Retrieve categories
// @Description Retrieves a category range
// @Tags Category
// @Param id path int true "Category Type id"
// @Param page query int true "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} rest.ApiResponse[[]dto.CategoryDTO] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving categories"
// @Router /category-types/{id}/categories [get]
// @Security access_token
func (h *Handler) GetCategoryiesByTypeId(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
		return
	}

	errResp, page, perPage := request_util.ParsePagination(r)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	categories, err := h.service.GetCategoriesByTypeId(r.Context(), int32(id), int32(page), int32(perPage))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, MsgCategoryNotFound))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Fetch category  - elements: %d", len(categories))
	h.Writer.WriteData(w, http.StatusOK, categories)
}

// @Summary Retrieve a category by ID
// @Description Retrieves a category specified by the ID.
// @Tags Category
// @Param id path int true "The ID of the category to retrieve"
// @Success 200 {object} rest.ApiResponse[dto.CategoryDTO] "OK - Successfully retrieved the category"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving the category"
// @Router /categories/{id} [get]
// @Security access_token
func (h *Handler) GetCategoryById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
		return
	}

	categoryDTO, err := h.service.GetCategoryById(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, MsgCategoryNotFound))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Fetch category: %s, ID: %s", categoryDTO.Name, categoryDTO.ID)
	h.Writer.WriteData(w, http.StatusOK, categoryDTO)
}

// @Summary Update a category by ID
// @Description Updates a category specified by the ID.
// @Tags Category
// @Param id path int true "The ID of the category to update"
// @Param dto body dto.PatchCategoryDTO true "Category data"
// @Success 200 {object} rest.ApiResponse[dto.CategoryDTO] "OK - Successfully updated the category"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while updating the category"
// @Router /categories/{id} [patch]
// @Security access_token
func (h *Handler) PatchCategoryById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
		return
	}

	var patchDto dto.PatchCategoryDTO
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

	updated, err := h.service.UpdateCategoryById(r.Context(), int32(id), patchDto)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, MsgCategoryNotFound))
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.RaiseException {
			h.Writer.WriteError(w, rest.NewConflictError(err, app_error.MsgNameAlreadyTaken))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Patch category: %s, ID: %d", patchDto.Name, id)
	h.Writer.WriteData(w, http.StatusOK, updated)
}

// @Summary Delete a category by ID
// @Description Deletes a category specified by the ID.
// @Tags Category
// @Param id path int true "The ID of the category to delete"
// @Success 204 {string} string "No Content - Successfully deleted the category"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while deleting the category"
// @Router /categories/{id} [delete]
// @Security access_token
func (h *Handler) DeleteCategoryById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
		return
	}

	err = h.service.DeleteCategoryById(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, pg_error.ErrNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, MsgCategoryNotFound))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Delete category - ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}
