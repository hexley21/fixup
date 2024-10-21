package domain

type (
	Subcategory struct {
		ID   int32
		Info SubcategoryInfo
	}

	SubcategoryInfo struct {
		CategoryID int32
		Name       string
	}
)

func NewSubcategory(id int32, categoryID int32, name string) Subcategory {
	info := NewSubcategoryInfo(categoryID, name)
	return Subcategory{
		ID:   id,
		Info: info,
	}
}

func NewSubcategoryInfo(categoryID int32, name string) SubcategoryInfo {
	return SubcategoryInfo{
		CategoryID: categoryID,
		Name:       name,
	}
}
