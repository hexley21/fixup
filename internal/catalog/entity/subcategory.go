package entity

type (
	Subcategory struct {
		ID   int32
		SubcategoryInfo
	}

	SubcategoryInfo struct {
		CategoryID int32
		Name       string
	}
)
