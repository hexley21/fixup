package dto

type (
	Subcategory struct {
		ID int32 `json:"id"`
		SubcategoryInfo
	}
	SubcategoryInfo struct {
		Name       string `json:"name" validate:"alpha,min=2,max=100,required"`
		CategoryID int32  `json:"category_id" validate:"number"`
	}
)
