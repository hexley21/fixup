package subcategory

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
	service        service.SubcategoryService
	defaultPerPage int64
	maxPerPage     int64
}

func NewHandler(
	handlerComponents *handler.Components,
	service service.SubcategoryService,
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

// Get
// @Summary Retrieve subcategory
// @Description Retrieves a subcategory range
// @Tags Subcategory
// @Param subcategory_id path int true "Subcategory id"
// @Success 200 {object} rest.ApiResponse[dto.Subcategory] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /subcategories/{subcategory_id} [get]
// @Security access_token
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "subcategory_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	subcategory, err := h.service.Get(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSubcategoryNotFound):
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
		default:
			h.Writer.WriteError(w, rest.NewInternalServerError(err))
		}
		return
	}

	h.Logger.Infof("Fetch subcategory: %s, ID: %d", subcategory.Info.Name, subcategory.ID)
	h.Writer.WriteData(w, http.StatusOK, mapper.MapSubcategoryToDTO(subcategory))
}

// List
// @Summary Retrieve subcategory
// @Description Retrieves a subcategory range
// @Tags Subcategory
// @Param page query int true "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} rest.ApiResponse[[]dto.Subcategory] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /subcategories [get]
// @Security access_token
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset, errResp := request_util.ParseLimitAndOffset(r, h.maxPerPage, h.defaultPerPage)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	subcategories, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	if subcategories == nil {
		h.Logger.Info("Fetch subcategories - 0")
		h.Writer.WriteData(w, http.StatusOK, []dto.Subcategory{})
		return
	}

	subcategoriesLen := len(subcategories)
	subcategoriesDTO := make([]dto.Subcategory, subcategoriesLen)
	for i, s := range subcategories {
		subcategoriesDTO[i] = mapper.MapSubcategoryToDTO(s)
	}

	h.Logger.Infof("Fetch subcategories - %d", subcategoriesLen)
	h.Writer.WriteData(w, http.StatusOK, subcategoriesDTO)
}

// ListByCategoryId
// @Summary Retrieve subcategory
// @Description Retrieves a subcategory range
// @Tags Subcategory
// @Param category_id path int true "Category id"
// @Param page query int true "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} rest.ApiResponse[[]dto.Subcategory] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /categories/{category_id}/subcategories [get]
// @Security access_token
func (h *Handler) ListByCategoryId(w http.ResponseWriter, r *http.Request) {
	categoryId, err := strconv.Atoi(chi.URLParam(r, "category_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	limit, offset, errResp := request_util.ParseLimitAndOffset(r, h.maxPerPage, h.defaultPerPage)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	subcategories, err := h.service.ListByCategoryId(r.Context(), int32(categoryId), limit, offset)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	if subcategories == nil {
		h.Logger.Infof("Fetch subcategories by category id: %d - 0", categoryId)
		h.Writer.WriteData(w, http.StatusOK, []dto.Subcategory{})
		return
	}

	subcategoriesLen := len(subcategories)
	subcategoriesDTO := make([]dto.Subcategory, subcategoriesLen)
	for i, s := range subcategories {
		subcategoriesDTO[i] = mapper.MapSubcategoryToDTO(s)
	}

	h.Logger.Infof("Fetch subcategories by category id: %d - %d", categoryId, subcategoriesLen)
	h.Writer.WriteData(w, http.StatusOK, subcategoriesDTO)
}

// ListByTypeId
// @Summary Retrieve subcategory
// @Description Retrieves a subcategory range
// @Tags Subcategory
// @Param type_id path int true "Category Type id"
// @Param page query int true "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} rest.ApiResponse[[]dto.Subcategory] "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /category-types/{type_id}/subcategories [get]
// @Security access_token
func (h *Handler) ListByTypeId(w http.ResponseWriter, r *http.Request) {
	typeId, err := strconv.Atoi(chi.URLParam(r, "type_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	limit, offset, errResp := request_util.ParseLimitAndOffset(r, h.maxPerPage, h.defaultPerPage)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	subcategories, err := h.service.ListByTypeId(r.Context(), int32(typeId), limit, offset)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	if subcategories == nil {
		h.Logger.Infof("Fetch subcategories by type id: %d - 0", typeId)
		h.Writer.WriteData(w, http.StatusOK, []dto.Subcategory{})
		return
	}

	subcategoriesLen := len(subcategories)
	subcategoriesDTO := make([]dto.Subcategory, subcategoriesLen)
	for i, s := range subcategories {
		subcategoriesDTO[i] = mapper.MapSubcategoryToDTO(s)
	}

	h.Logger.Infof("Fetch subcategories by type id: %d - %d", typeId, subcategoriesLen)
	h.Writer.WriteData(w, http.StatusOK, subcategoriesDTO)
}

// Create
// @Summary Create a new subcategory
// @Description Creates a new subcategory with the provided data.
// @Tags Subcategory
// @Param dto body dto.SubcategoryInfo true "Subcategory info"
// @Success 201 {object} rest.ApiResponse[dto.Subcategory]
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /subcategories [post]
// @Security access_token
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var infoDTO dto.SubcategoryInfo
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

	infoVO, err := mapper.MapSubcategoryInfoToVO(infoDTO)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidArgumentsError(err))
		return
	}

	subcategoryId, err := h.service.Create(r.Context(), infoVO)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSubcategoryNameTaken):
			h.Writer.WriteError(w, rest.NewConflictError(err))
		default:
			h.Writer.WriteError(w, rest.NewInternalServerError(err))
		}
		return
	}

	h.Logger.Infof("Create subcategory: %s, Category-ID: %s ID: %d", infoDTO.Name, infoDTO.CategoryID, subcategoryId)
	h.Writer.WriteData(w, http.StatusCreated, dto.Subcategory{
		ID:              strconv.Itoa(int(subcategoryId)),
		SubcategoryInfo: infoDTO,
	})
}

// Update
// @Summary Updates a new subcategory
// @Description Updates a new subcategory with the provided data.
// @Tags Subcategory
// @Param subcategory_id path int true "The ID of the subcategory to update"
// @Param dto body dto.SubcategoryInfo true "Subcategory info"
// @Success 200 {object} rest.ApiResponse[dto.Subcategory]
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /subcategories/{subcategory_id} [patch]
// @Security access_token
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "subcategory_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	var infoDTO dto.SubcategoryInfo
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

	infoVO, err := mapper.MapSubcategoryInfoToVO(infoDTO)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidArgumentsError(err))
		return
	}

	subcategory, err := h.service.Update(r.Context(), int32(id), infoVO)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSubcategoryNameTaken):
			h.Writer.WriteError(w, rest.NewConflictError(err))
		case errors.Is(err, service.ErrSubcategoryNotFound), errors.Is(err, service.ErrCategoryNotFound):
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
		default:
			h.Writer.WriteError(w, rest.NewInternalServerError(err))
		}
		return
	}

	h.Logger.Infof("Update subcategory: %s, Category-ID: %d ID: %d", subcategory.Info.Name, subcategory.Info.CategoryID, subcategory.ID)
	h.Writer.WriteData(w, http.StatusOK, mapper.MapSubcategoryToDTO(subcategory))
}

// Delete
// @Summary Deletes a subcategory by ID
// @Description Deletes a subcategory specified by the ID.
// @Tags Subcategory
// @Param subcategory_id path int true "The ID of the subcategory to delete"
// @Success 204 {string} string "No Content - Successfully deleted"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 403 {object} rest.ErrorResponse "Forbidden"
// @Failure 404 {object} rest.ErrorResponse "Not Found"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /subcategories/{subcategory_id} [delete]
// @Security access_token
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "subcategory_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidIdError(err))
		return
	}

	err = h.service.Delete(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSubcategoryNotFound):
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
		default:
			h.Writer.WriteError(w, rest.NewInternalServerError(err))
		}
		return
	}

	h.Logger.Infof("Delete subcategory - ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}
