package category_type

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/mapper"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/hexley21/fixup/internal/common/util/request_util"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/rest"
)

type Handler struct {
	*handler.Components
	service        service.CategoryTypeService
	defaultPerPage int64
	maxPerPage     int64
}

func NewHandler(
	handlerComponents *handler.Components,
	service service.CategoryTypeService,
	defaultPerPage int64,
	maxPerPage int64,
) *Handler {
	return &Handler{
		Components:     handlerComponents,
		service:        service,
		defaultPerPage: defaultPerPage,
		maxPerPage:     maxPerPage,
	}
}

// Create
// @Summary Create a new category type
// @Description Creates a new category type with the provided data.
// @Tags CategoryType
// @Param dto body dto.CategoryTypeInfo true "Category type data"
// @Success 201 {object} rest.ApiResponse[dto.CategoryType] "Created - Successfully created the category type"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while creating the category type"
// @Router /category-types [post]
// @Security access_token
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var infoDTO dto.CategoryTypeInfo
	errResp := h.Binder.BindJSON(r, &infoDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	errResp = h.Validator.Validate(&infoDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	categoryType, err := h.service.Create(r.Context(), infoDTO.Name)
	if err != nil {
		if errors.Is(err, service.ErrCateogryTypeNameTaken) {
			h.Writer.WriteError(w, rest.NewConflictError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to create category type: %w", err))
		return
	}

	h.Logger.Infof("Create category type: %s, ID: %d", categoryType.Name, categoryType.ID)
	h.Writer.WriteData(w, http.StatusCreated, categoryType)
}

// List
// @Summary Retrieve a category types
// @Description Retrieves a category type range
// @Tags CategoryType
// @Param page query int true "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} rest.ApiResponse[[]dto.CategoryType] "OK - Successfully retrieved the category types"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving the category type"
// @Router /category-types [get]
// @Security access_token
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset, errResp := request_util.ParseLimitAndOffset(r, h.maxPerPage, h.defaultPerPage)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	typeEntities, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to fetch list of cateogry types: %w", err))
		return
	}

	typesLen := len(typeEntities)
	typeDTOs := make([]dto.CategoryType, typesLen)
	for i, ct := range typeEntities {
		typeDTOs[i] = mapper.MapCategoryTypeToDTO(ct)
	}

	h.Logger.Infof("Fetch category types - %d", typesLen)
	h.Writer.WriteData(w, http.StatusOK, typeDTOs)
}

// Get
// @Summary Retrieve a category type by ID
// @Description Retrieves a category type specified by the ID.
// @Tags CategoryType
// @Param type_id path int true "The ID of the category type to retrieve"
// @Success 200 {object} rest.ApiResponse[dto.CategoryType] "OK - Successfully retrieved the category type"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving the category type"
// @Router /category-types/{id} [get]
// @Security access_token
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "type_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	typeEntity, err := h.service.Get(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrCategoryTypeNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to fetch category type - id: %d, error: %w", id, err))
		return
	}

	h.Logger.Infof("Fetch category type: %s, ID: %d", typeEntity.Name, typeEntity.ID)
	h.Writer.WriteData(w, http.StatusOK, mapper.MapCategoryTypeToDTO(typeEntity))
}

// Update
// @Summary Update a category type by ID
// @Description Updates a category type specified by the ID.
// @Tags CategoryType
// @Param type_id path int true "The ID of the category type to update"
// @Param dto body dto.CategoryTypeInfo true "Category type data"
// @Success 200 {object} rest.ApiResponse[dto.CategoryType] "OK - Successfully updated the category type"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while updating the category type"
// @Router /category-types/{id} [patch]
// @Security access_token
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "type_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	var infoDTO dto.CategoryTypeInfo
	errResp := h.Binder.BindJSON(r, &infoDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	errResp = h.Validator.Validate(&infoDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	err = h.service.Update(r.Context(), int32(id), infoDTO.Name)
	if err != nil {
		if errors.Is(err, service.ErrCategoryTypeNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		if errors.Is(err, service.ErrCateogryTypeNameTaken) {
			h.Writer.WriteError(w, rest.NewConflictError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Update category type: %s, ID: %d", infoDTO.Name, id)
	h.Writer.WriteData(w, http.StatusOK, dto.NewCategoryType(strconv.Itoa(id), infoDTO.Name))
}

// Delete
// @Summary Delete a category type by ID
// @Description Deletes a category type specified by the ID.
// @Tags CategoryType
// @Param type_id path int true "The ID of the category type to delete"
// @Success 204 {string} string "No Content - Successfully deleted the category type"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while deleting the category type"
// @Router /category-types/{id} [delete]
// @Security access_token
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "type_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	err = h.service.Delete(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrCategoryTypeNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to delete category type - id: %d, error: %w", id, err))
		return
	}

	h.Logger.Infof("Delete category type - ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}
