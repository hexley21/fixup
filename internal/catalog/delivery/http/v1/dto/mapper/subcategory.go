package mapper

import (
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/entity"
)

func SubcategoryToDTO(entity entity.Subcategory) dto.Subcategory {
	return dto.Subcategory{
		ID:              entity.ID,
		SubcategoryInfo: SubcategoryInfoToDTO(entity.SubcategoryInfo),
	}

}

func SubcategoryInfoToDTO(entity entity.SubcategoryInfo) dto.SubcategoryInfo {
	return dto.SubcategoryInfo{
		CategoryID: entity.CategoryID,
		Name:       entity.Name,
	}
}
