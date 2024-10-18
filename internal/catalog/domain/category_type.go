package domain

type CategoryType struct {
	ID   int32
	Name string
}

func NewCategoryType(id int32, name string) CategoryType {
	return CategoryType{
		ID:   id,
		Name: name,
	}
}
