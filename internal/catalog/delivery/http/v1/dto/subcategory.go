package dto

type (
	Subcategory struct {
		ID int32 `json:"id"`
		SubcategoryInfo
	}
	SubcategoryInfo struct {
		Name       string `json:"name"`
		CategoryID int32  `json:"category_id"`
	}
)
