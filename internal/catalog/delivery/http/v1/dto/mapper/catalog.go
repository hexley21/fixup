package mapper

import (
	"strconv"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/entity"
)

func MapCategoryTypeToDTO(entity entity.CategoryType) dto.CategoryTypeDTO {
	return dto.CategoryTypeDTO{
		ID: strconv.FormatInt(int64(entity.ID), 10),
		Name: entity.Name,
	}
}

func MapCategoryToDTO(entity entity.Category) dto.CategoryDTO {
	return dto.CategoryDTO{
		ID: strconv.Itoa(int(entity.ID)),
		TypeID: strconv.Itoa(int(entity.TypeID)),
		Name: entity.Name,
	}
}