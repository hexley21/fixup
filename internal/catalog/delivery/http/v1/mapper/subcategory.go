package mapper

import (
	"strconv"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/domain"
)

func MapSubcategoryToDTO(entity domain.Subcategory) dto.Subcategory {
	return dto.Subcategory{
		ID:              strconv.Itoa(int(entity.ID)),
		SubcategoryInfo: MapSubcategoryInfoToDTO(entity.Info),
	}

}

func MapSubcategoryInfoToDTO(entity domain.SubcategoryInfo) dto.SubcategoryInfo {
	return dto.SubcategoryInfo{
		CategoryID: strconv.Itoa(int(entity.CategoryID)),
		Name:       entity.Name,
	}
}

func MapSubcategoryToEntity(subcategoryDTO dto.Subcategory) (domain.Subcategory, error) {
	intId, err := strconv.ParseInt(subcategoryDTO.ID, 10, 32)
	if err != nil {
		return domain.Subcategory{}, err
	}

	info, err := MapSubcategoryInfoToEntity(subcategoryDTO.SubcategoryInfo)
	if err != nil {
		return domain.Subcategory{}, err
	}

	return domain.Subcategory{
		ID:   int32(intId),
		Info: info,
	}, nil

}

func MapSubcategoryInfoToEntity(dto dto.SubcategoryInfo) (domain.SubcategoryInfo, error) {
	intId, err := strconv.ParseInt(dto.CategoryID, 10, 32)
	if err != nil {
		return domain.SubcategoryInfo{}, err
	}

	return domain.SubcategoryInfo{
		CategoryID: int32(intId),
		Name:       dto.Name,
	}, nil
}
