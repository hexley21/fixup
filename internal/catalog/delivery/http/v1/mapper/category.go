package mapper

import (
	"strconv"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/domain"
)

func MapCategoryInfoToVO(infoDTO dto.CategoryInfo) (domain.CategoryInfo, error) {
	intId, err := strconv.ParseInt(infoDTO.TypeID, 10, 32)
	if err != nil {
		return domain.CategoryInfo{}, err
	}

	return domain.CategoryInfo{
		TypeID: int32(intId),
		Name:   infoDTO.Name,
	}, nil
}

func MapCategoryToDTO(entity domain.Category) dto.Category {
	return dto.NewCategoryDTO(strconv.FormatInt(int64(entity.ID), 10), entity.Info.Name, strconv.FormatInt(int64(entity.Info.TypeID), 10))
}
