package domain

type (
	Category struct {
		ID   int32
		Info CategoryInfo
	}

	CategoryInfo struct {
		TypeID int32
		Name   string
	}
)

func NewCategory(id int32, typeID int32, name string) Category {
	info := NewCategoryInfo(typeID, name)
	return Category{
		ID:   id,
		Info: info,
	}
}

func NewCategoryInfo(typeID int32, name string) CategoryInfo {
	return CategoryInfo{
		TypeID: typeID,
		Name:   name,
	}
}
