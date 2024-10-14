package subcategory

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/entity"
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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "subcategory_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
		return
	}

	subcategoryDTO, err := h.service.Get(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrSubcategoryNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, err.Error()))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Fetch subcategory: %s, ID: %d", subcategoryDTO.Name, subcategoryDTO.ID)
	h.Writer.WriteData(w, http.StatusOK, subcategoryDTO)
}

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
		subcategoriesDTO[i] = dto.Subcategory{
			ID: s.ID,
			SubcategoryInfo: dto.SubcategoryInfo{
				Name:       s.Name,
				CategoryID: s.CategoryID,
			},
		}
	}

	h.Logger.Infof("Fetch subcategories - %d", subcategoriesLen)
	h.Writer.WriteData(w, http.StatusOK, subcategoriesDTO)
}

func (h *Handler) ListByCategoryId(w http.ResponseWriter, r *http.Request) {
	categoryId, err := strconv.Atoi(chi.URLParam(r, "category_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
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
		subcategoriesDTO[i] = dto.Subcategory{
			ID: s.ID,
			SubcategoryInfo: dto.SubcategoryInfo{
				Name:       s.Name,
				CategoryID: s.CategoryID,
			},
		}
	}

	h.Logger.Infof("Fetch subcategories by category id: %d - %d", categoryId, subcategoriesLen)
	h.Writer.WriteData(w, http.StatusOK, subcategoriesDTO)
}

func (h *Handler) ListByTypeId(w http.ResponseWriter, r *http.Request) {
	typeId, err := strconv.Atoi(chi.URLParam(r, "type_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
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
		subcategoriesDTO[i] = dto.Subcategory{
			ID: s.ID,
			SubcategoryInfo: dto.SubcategoryInfo{
				Name:       s.Name,
				CategoryID: s.CategoryID,
			},
		}
	}

	h.Logger.Infof("Fetch subcategories by type id: %s - %d", typeId, subcategoriesLen)
	h.Writer.WriteData(w, http.StatusOK, subcategoriesDTO)
}

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

	subcategoryId, err := h.service.Create(r.Context(), entity.SubcategoryInfo{
		Name:       infoDTO.Name,
		CategoryID: infoDTO.CategoryID,
	})

	if err != nil {
		if errors.Is(err, service.ErrSubcateogryNameTaken) {
			h.Writer.WriteError(w, rest.NewConflictError(err, err.Error()))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Create subcategory: %s, Category-ID: %d ID: %d", infoDTO.Name, infoDTO.CategoryID, subcategoryId)
	h.Writer.WriteData(w, http.StatusCreated, dto.Subcategory{
		ID:   subcategoryId,
		SubcategoryInfo: infoDTO,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "subcategory_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
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

	subcategory, err := h.service.Update(r.Context(), int32(id), entity.SubcategoryInfo{
		Name:       infoDTO.Name,
		CategoryID: infoDTO.CategoryID,
	})

	if err != nil {
		if errors.Is(err, service.ErrSubcateogryNameTaken) {
			h.Writer.WriteError(w, rest.NewConflictError(err, err.Error()))
			return
		}

		if errors.Is(err, service.ErrSubcategoryNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, err.Error()))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Update subcategory: %s, Category-ID: %d ID: %d", subcategory.Name, subcategory.CategoryID, subcategory.ID)
	h.Writer.WriteData(w, http.StatusOK, dto.Subcategory{
		ID: subcategory.ID,
		SubcategoryInfo: dto.SubcategoryInfo{
			Name:       subcategory.Name,
			CategoryID: subcategory.CategoryID,
		},
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "subcategory_id"))
	if err != nil {
		h.Writer.WriteError(w, rest.NewBadRequestError(err, rest.MsgInvalidId))
		return
	}

	err = h.service.Delete(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrSubcategoryNotFound) {
			h.Writer.WriteError(w, rest.NewNotFoundError(err, err.Error()))
			return
		}

		h.Writer.WriteError(w, rest.NewInternalServerError(err))
		return
	}

	h.Logger.Infof("Delete subcategory - ID: %d", id)
	h.Writer.WriteNoContent(w, http.StatusNoContent)
}
