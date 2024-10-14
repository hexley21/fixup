package dto

type (
	Subcategory struct {
		ID string `json:"id"`
		SubcategoryInfo
	}
	SubcategoryInfo struct {
		Name       string `json:"name" validate:"alpha,min=2,max=100,required"`
		CategoryID string  `json:"category_id" validate:"number"`
	}
)
