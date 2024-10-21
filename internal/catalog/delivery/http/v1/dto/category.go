package dto

type Category struct {
	ID     string `json:"id"`
	CategoryInfo
} // @name Category

type CategoryInfo struct {
	Name   string `json:"name" validate:"alpha,min=2,max=30,required"`
	TypeID string `json:"type_id" validate:"number"`
} // @name CategoryInfo

func NewCategoryDTO(id string, name string, typeId string) Category {
	return Category{
		ID: id,
		CategoryInfo: NewCategoryInfoVO(name, typeId),
	}
}

func NewCategoryInfoVO(name string, typeId string) CategoryInfo {
	return CategoryInfo{
		Name: name,
		TypeID: typeId,
	}
}
