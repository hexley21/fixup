package dto

type (
	Subcategory struct {
		ID string `json:"id"`
		SubcategoryInfo
	}
	SubcategoryInfo struct {
		Name       string `json:"name" validate:"alpha,min=2,max=100,required"`
		CategoryID string `json:"category_id" validate:"number"`
	}
)

func NewSubcategoryDTO(id string, name string, categoryId string) Subcategory {
	return Subcategory{
		ID:              id,
		SubcategoryInfo: NewSubcategoryInfoDTO(name, categoryId),
	}
}

func NewSubcategoryInfoDTO(name string, categoryId string) SubcategoryInfo {
	return SubcategoryInfo{
		Name:       name,
		CategoryID: categoryId,
	}
}
