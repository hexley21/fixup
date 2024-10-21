package dto

type CategoryType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name CategoryType

type CategoryTypeInfo struct {
	Name string `json:"name" validate:"alpha,min=2,max=30,required"`
} // @name CategoryTypeInfo


func NewCategoryType(id string, name string) CategoryType {
	return CategoryType{
		ID: id,
		Name: name,
	}
}