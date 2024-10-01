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
