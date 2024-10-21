package category

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
	service        service.CategoryService
	defaultPerPage int64
	maxPerPage     int64
}

func NewHandler(
	handlerComponents *handler.Components,
	service service.CategoryService,
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
// @Summary Create a new category
// @Description Creates a new category with the provided data.
// @Tags Category
// @Param dto body dto.CategoryInfo true "Category data"
// @Success 201 {object} rest.ApiResponse[dto.Category] "Created - Successfully created the category"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while creating the category"
// @Router /categories [post]
// @Security access_token
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var infoDTO dto.CategoryInfo
	errResp := h.Binder.BindJSON(r, &infoDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	errResp = h.Validator.Validate(infoDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	infoVO, err := mapper.MapCategoryInfoToVO(infoDTO)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to create category due to wrong validation: %w", err))
		return
	}

	categoryId, err := h.service.Create(r.Context(), infoVO)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNameTaken) {
			h.Writer.WriteError(w, rest.NewConflictError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to create category: %w", err))
		return
	}

	h.Logger.Infof("Create category: %s, Type-ID: %d ID: %d", infoVO.Name, infoVO.TypeID, categoryId)
	h.Writer.WriteData(
		w, http.StatusCreated,
		dto.NewCategoryDTO(strconv.FormatInt(int64(categoryId), 10), infoDTO.Name, infoDTO.TypeID),
	)
}

// List
// @Summary Retrieve categories
// @Description Retrieves a category range
// @Tags Category
// @Param page query int true "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} rest.ApiResponse[[]dto.Category] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving categories"
// @Router /categories [get]
// @Security access_token
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset, errResp := request_util.ParseLimitAndOffset(r, h.maxPerPage, h.defaultPerPage)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	categoryEntities, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to fetch categories: %w", err))
		return
	}

	categoriesLen := len(categoryEntities)
	categoryDTOs := make([]dto.Category, categoriesLen)
	for i, c := range categoryEntities {
		categoryDTOs[i] = mapper.MapCategoryToDTO(c)
	}

	h.Logger.Infof("Fetch categories - %d", categoriesLen)
	h.Writer.WriteData(w, http.StatusOK, categoryDTOs)
}

// ListByTypeId
// @Summary Retrieve categories
// @Description Retrieves a category range
// @Tags Category
// @Param type_id path int true "Category Type id"
// @Param page query int true "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} rest.ApiResponse[[]dto.Category] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving categories"
// @Router /category-types/{type_id}/categories [get]
// @Security access_token
func (h *Handler) ListByTypeId(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "type_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	limit, offset, errResp := request_util.ParseLimitAndOffset(r, h.maxPerPage, h.defaultPerPage)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	categoryEntities, err := h.service.ListByTypeId(r.Context(), int32(id), limit, offset)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to fetch categories - type id: %d, error: %w", id, err))
		return
	}

	categoriesLen := len(categoryEntities)
	categoryDTOs := make([]dto.Category, categoriesLen)
	for i, c := range categoryEntities {
		categoryDTOs[i] = mapper.MapCategoryToDTO(c)
	}

	h.Logger.Infof("Fetch categories - %d", categoriesLen)
	h.Writer.WriteData(w, http.StatusOK, categoryDTOs)
}

// Get
// @Summary Retrieve a category by ID
// @Description Retrieves a category specified by the ID.
// @Tags Category
// @Param category_id path int true "The ID of the category to retrieve"
// @Success 200 {object} rest.ApiResponse[dto.Category] "OK - Successfully retrieved the category"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while retrieving the category"
// @Router /categories/{category_id} [get]
// @Security access_token
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "category_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	categoryEntity, err := h.service.Get(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to get category - id: %d, error: %w", id, err))
		return
	}

	h.Logger.Infof("Fetch category: %s, ID: %d", categoryEntity.Info.Name, categoryEntity.ID)
	h.Writer.WriteData(w, http.StatusOK, mapper.MapCategoryToDTO(categoryEntity))
}

// Update
// @Summary Update a category by ID
// @Description Updates a category specified by the ID.
// @Tags Category
// @Param category_id path int true "The ID of the category to update"
// @Param dto body dto.CategoryInfo true "Category data"
// @Success 200 {object} rest.ApiResponse[dto.Category] "OK - Successfully updated the category"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while updating the category"
// @Router /categories/{category_id} [patch]
// @Security access_token
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "category_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	var infoDTO dto.CategoryInfo
	errResp := h.Binder.BindJSON(r, &infoDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	errResp = h.Validator.Validate(infoDTO)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	infoVO, err := mapper.MapCategoryInfoToVO(infoDTO)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to update category due to wrong validation: %w", err))
		return
	}

	caetgoryEntity, err := h.service.Update(r.Context(), int32(id), infoVO)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		if errors.Is(err, service.ErrCategoryNameTaken) {
			h.Writer.WriteError(w, rest.NewConflictError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to update category - id: %d, error: %w", caetgoryEntity.ID, err))
		return
	}

	h.Logger.Infof("Update category: %s, ID: %d", caetgoryEntity.Info.Name, id)
	h.Writer.WriteData(w, http.StatusOK, mapper.MapCategoryToDTO(caetgoryEntity))
}

// Delete
// @Summary Delete a category by ID
// @Description Deletes a category specified by the ID.
// @Tags Category
// @Param category_id path int true "The ID of the category to delete"
// @Success 204 {string} string "No Content - Successfully deleted the category"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error - An error occurred while deleting the category"
// @Router /categories/{category_id} [delete]
// @Security access_token
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "category_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	err = h.service.Delete(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to delete category - id: %d, error: %w", id, err))
		return
	}

	h.Logger.Infof("Delete category - ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}
