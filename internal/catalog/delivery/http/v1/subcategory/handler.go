package subcategory

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/hexley21/fixup/internal/common/util/request_util"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/rest"
)

type Handler struct {
	*handler.Components
	service        service.Subcategory
	defaultPerPage int
	maxPerPage     int
}

func NewHandler(
	handlerComponents *handler.Components,
	service service.Subcategory,
	defaultPerPage int,
	maxPerPage int,
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
		h.Writer.WriteError(w, rest.NewBadRequestError(err))
		return
	}

	subcategory, err := h.service.Get(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrSubcategoryNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Fetch subcategory: %s, ID: %d", subcategory.Name, subcategory.ID)
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
	offset, limit, errResp := request_util.ParseOffsetAndLimit(r, h.maxPerPage, h.defaultPerPage)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	subcategories, err := h.service.List(r.Context(), offset, limit)
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

// List
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
		h.Writer.WriteError(w, rest.NewBadRequestError(err))
		return
	}

	offset, limit, errResp := request_util.ParseOffsetAndLimit(r, h.maxPerPage, h.defaultPerPage)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	subcategories, err := h.service.ListByCategoryId(r.Context(), int32(categoryId), offset, limit)
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


// List
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
		h.Writer.WriteError(w, rest.NewBadRequestError(err))
		return
	}

	offset, limit, errResp := request_util.ParseOffsetAndLimit(r, h.maxPerPage, h.defaultPerPage)
	if errResp != nil {
		h.Writer.WriteError(w, errResp)
		return
	}

	subcategories, err := h.service.ListByTypeId(r.Context(), int32(typeId), offset, limit)
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

	infoEntity, err := mapper.MapSubcategoryInfoToEntity(infoDTO)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidArgumentsError(err))
		return
	}

	subcategoryId, err := h.service.Create(r.Context(), infoEntity)

	if err != nil {
		if errors.Is(err, service.ErrSubcateogryNameTaken) {
			h.Writer.WriteError(w, rest.NewConflictError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Create subcategory: %s, Category-ID: %s ID: %d", infoDTO.Name, infoDTO.CategoryID, subcategoryId)
	h.Writer.WriteData(w, http.StatusCreated, dto.Subcategory{
		ID:   strconv.Itoa(int(subcategoryId)),
		SubcategoryInfo: infoDTO,
	})
}

// Create
// @Summary Create a new subcategory
// @Description Creates a new subcategory with the provided data.
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
		h.Writer.WriteError(w, rest.NewBadRequestError(err))
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

	infoEntity, err := mapper.MapSubcategoryInfoToEntity(infoDTO)
	if err != nil {
		h.Writer.WriteError(w, rest.NewInvalidArgumentsError(err))
		return
	}

	subcategory, err := h.service.Update(r.Context(), int32(id), infoEntity)

	if err != nil {
		if errors.Is(err, service.ErrSubcateogryNameTaken) {
			h.Writer.WriteError(w, rest.NewConflictError(err))
			return
		}

		if errors.Is(err, service.ErrSubcategoryNotFound) || errors.Is(err, service.ErrCategoryNotFound){
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Update subcategory: %s, Category-ID: %d ID: %d", subcategory.Name, subcategory.CategoryID, subcategory.ID)
	h.Writer.WriteData(w, http.StatusOK, mapper.MapSubcategoryToDTO(subcategory))
}

// Delete
// @Summary Delete a subcategory by ID
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
		h.Writer.WriteError(w, rest.NewBadRequestError(err))
		return
	}

	err = h.service.Delete(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrSubcategoryNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Delete subcategory - ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}
