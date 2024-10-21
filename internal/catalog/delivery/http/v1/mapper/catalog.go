package mapper

import (
	"strconv"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/domain"
)

func MapCategoryTypeToDTO(entity domain.CategoryType) dto.CategoryType {
	return dto.NewCategoryType(strconv.FormatInt(int64(entity.ID), 10), entity.Name)
}

// func MapCategoryToDTO(entity domain.Category) dto.CategoryDTO {
// 	return dto.CategoryDTO{
// 		ID: strconv.Itoa(int(entity.ID)),
// 		TypeID: strconv.Itoa(int(entity.Info.TypeID)),
// 		Name: entity.Info.Name,
// 	}
// }