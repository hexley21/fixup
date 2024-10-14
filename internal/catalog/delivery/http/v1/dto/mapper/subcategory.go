package mapper

import (
	"strconv"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/entity"
)

func MapSubcategoryToDTO(entity entity.Subcategory) dto.Subcategory {
	return dto.Subcategory{
		ID:              strconv.Itoa(int(entity.ID)),
		SubcategoryInfo: MapSubcategoryInfoToDTO(entity.SubcategoryInfo),
	}

}

func MapSubcategoryInfoToDTO(entity entity.SubcategoryInfo) dto.SubcategoryInfo {
	return dto.SubcategoryInfo{
		CategoryID: strconv.Itoa(int(entity.CategoryID)),
		Name:       entity.Name,
	}
}

func MapSubcategoryToEntity(dto dto.Subcategory) (entity.Subcategory, error) {
	intId, err := strconv.ParseInt(dto.ID, 10, 32)
	if err != nil {
		return entity.Subcategory{}, err
	}

	info, err := MapSubcategoryInfoToEntity(dto.SubcategoryInfo)
	if err != nil {
		return entity.Subcategory{}, err
	}

	return entity.Subcategory{
		ID:              int32(intId),
		SubcategoryInfo: info,
	}, nil

}

func MapSubcategoryInfoToEntity(dto dto.SubcategoryInfo) (entity.SubcategoryInfo, error) {
	intId, err := strconv.ParseInt(dto.CategoryID, 10, 32)
	if err != nil {
		return entity.SubcategoryInfo{}, err
	}

	return entity.SubcategoryInfo{
		CategoryID: int32(intId),
		Name:       dto.Name,
	}, nil
}
