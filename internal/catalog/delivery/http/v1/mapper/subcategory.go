package mapper

import (
	"strconv"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/domain"
)

func MapSubcategoryInfoToVO(dto dto.SubcategoryInfo) (domain.SubcategoryInfo, error) {
	intId, err := strconv.ParseInt(dto.CategoryID, 10, 32)
	if err != nil {
		return domain.SubcategoryInfo{}, err
	}

	return domain.SubcategoryInfo{
		CategoryID: int32(intId),
		Name:       dto.Name,
	}, nil
}

func MapSubcategoryToDTO(entity domain.Subcategory) dto.Subcategory {
	return dto.NewSubcategoryDTO(strconv.Itoa(int(entity.ID)), entity.Info.Name, strconv.Itoa(int(entity.Info.CategoryID)))
}
